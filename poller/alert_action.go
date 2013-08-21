package poller

import (
	"bytes"
	"commons"
	"crypto/md5"
	"crypto/rand"
	ds "data_store"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	//"commons/types"
	"encoding/json"
	"errors"
	"flag"
	"sync/atomic"
	"time"
)

const MAX_REPEATED = 9999990

var (
	notification_priority     = flag.Int("alert.notification.priority", 0, "the priority of notification job")
	notification_max_attempts = flag.Int("alert.notification.max_attempts", -1, "the max attempts while notification push failed")
	notification_queue        = flag.String("alert.notification.queue", "", "the default queue name")

	reset_error = errors.New("please reset channel.")
)

var (

	// objectIdCounter is atomically incremented when generating a new ObjectId
	// using NewObjectId() function. It's used as a counter part of an id.
	objectIdCounter uint32 = 0

	// machineId stores machine id generated once and used in subsequent calls
	// to NewObjectId function.
	machineId  = readMachineId()
	currentPid = os.Getpid()
)

// initMachineId generates machine id and puts it into the machineId global
// variable. If this function fails to get the hostname, it will cause
// a runtime error.
func readMachineId() []byte {
	var sum [3]byte
	id := sum[:]
	hostname, err1 := os.Hostname()
	if err1 != nil {
		_, err2 := io.ReadFull(rand.Reader, id)
		if err2 != nil {
			panic(fmt.Errorf("cannot get hostname: %v; %v", err1, err2))
		}
		return id
	}
	hw := md5.New()
	hw.Write([]byte(hostname))
	copy(id, hw.Sum(nil))
	return id
}

// NewObjectId returns a new unique ObjectId.
// This function causes a runtime error if it fails to get the hostname
// of the current machine.
func generateId() string {
	var b [12]byte
	// Timestamp, 4 bytes, big endian
	binary.BigEndian.PutUint32(b[:], uint32(time.Now().Unix()))
	// Machine, first 3 bytes of md5(hostname)
	b[4] = machineId[0]
	b[5] = machineId[1]
	b[6] = machineId[2]
	// Pid, 2 bytes, specs don't specify endianness, but we use big endian.
	b[7] = byte(currentPid >> 8)
	b[8] = byte(currentPid)
	// Increment, 3 bytes, big endian
	i := atomic.AddUint32(&objectIdCounter, 1)
	b[9] = byte(i >> 16)
	b[10] = byte(i >> 8)
	b[11] = byte(i)
	return hex.EncodeToString(b[:])
}

type alertAction struct {
	id          int64
	name        string
	delay_times int

	options     map[string]interface{}
	contex      map[string]interface{}
	publish     chan<- []string
	channel     chan<- *data_object
	cached_data *data_object

	templates       []*template.Template
	information     string
	checker         Checker
	previous_status int
	last_status     int
	repeated        int
	already_send    bool
	last_event_id   string
	sequence_id     int

	notification_group_ids []string
	notification_groups    *ds.Cache

	begin_send_at, wait_response_at, responsed_at, end_send_at int64

	stats_previous_status int
	stats_last_status     int
	stats_repeated        int
	stats_already_send    bool
	stats_last_event_id   string
	stats_sequence_id     int
	stats_information     string
}

func (self *alertAction) Stats() map[string]interface{} {
	stats := map[string]interface{}{
		"type":             "alert",
		"id":               self.id,
		"name":             self.name,
		"information":      self.stats_information,
		"previous_status":  self.stats_previous_status,
		"last_status":      self.stats_last_status,
		"repeated":         self.stats_repeated,
		"already_send":     self.stats_already_send,
		"event_id":         self.stats_last_event_id,
		"sequence_id":      self.stats_sequence_id,
		"begin_send_at":    atomic.LoadInt64(&self.begin_send_at),
		"wait_response_at": atomic.LoadInt64(&self.wait_response_at),
		"responsed_at":     atomic.LoadInt64(&self.responsed_at),
		"end_send_at":      atomic.LoadInt64(&self.end_send_at)}

	if nil != self.notification_groups {
		stats["notification_group_ids"] = self.notification_group_ids
	} else {
		stats["notification_group_ids"] = ""
	}
	return stats
}

func (self *alertAction) RunBefore() {
}

func (self *alertAction) RunAfter() {
	self.stats_previous_status = self.previous_status
	self.stats_last_status = self.last_status
	self.stats_repeated = self.repeated
	self.stats_already_send = self.already_send
	self.stats_last_event_id = self.last_event_id
	self.stats_sequence_id = self.sequence_id
	self.stats_information = self.information
}

func (self *alertAction) Run(t time.Time, value interface{}) error {
	if res, ok := value.(commons.Result); ok {
		if res.HasError() {
			return errors.New("sampling failed, " + res.ErrorMessage())
		}
	}

	current, current_value, err := self.checker.Run(value, self.contex)
	if nil != err {
		return err
	}

	if current == self.last_status {
		self.repeated++

		if self.repeated >= 9999996 || self.repeated < 0 { // inhebit overflow
			self.repeated = self.delay_times + 10
		}
	} else {
		self.repeated = 1
		self.last_status = current
		self.already_send = false
	}

	if self.repeated < self.delay_times {
		return nil
	}

	if self.already_send {
		return nil
	}

	evt := map[string]interface{}{}
	for k, v := range self.contex {
		evt[k] = v
	}
	if _, found := evt["triggered_at"]; !found {
		evt["triggered_at"] = t
	}

	if _, found := evt["current_value"]; !found {
		bs, err := json.Marshal(current_value)
		if nil != err {
			return errors.New("marshal current value failed," + err.Error())
		}
		if nil != bs {
			evt["current_value"] = string(bs)
		}
	}

	if 0 == self.previous_status {
		self.sequence_id = 1
		self.last_event_id = generateId()
	}

	evt["event_id"] = self.last_event_id
	evt["sequence_id"] = self.sequence_id
	evt["previous_status"] = self.previous_status
	evt["status"] = current
	evt["content"] = self.gen_message(current, self.previous_status, evt)
	if nil != self.notification_groups {
		err := self.fillNotificationData(current, evt)
		if nil != err {
			return err
		}
	}

	err = self.send(evt)
	if nil == err {
		self.previous_status = current
		self.sequence_id++
		self.already_send = true
		return nil
	}

	if err == reset_error {
		self.cached_data = &data_object{c: make(chan error, 2)}
	}
	return err
}

func (self *alertAction) fillNotificationData(current int, evt map[string]interface{}) error {
	if nil == self.notification_group_ids || 0 == len(self.notification_group_ids) {
		return nil
	}

	notification_group_id := self.notification_group_ids[0]
	switch len(self.notification_group_ids) {
	case 1:
		break
	case 2:
		if 0 != current {
			notification_group_id = self.notification_group_ids[1]
		}
		break
	default:
		if current >= len(self.notification_group_ids) {
			notification_group_id = self.notification_group_ids[1]
		} else {
			notification_group_id = self.notification_group_ids[current]
		}
	}

	rules, err := self.notification_groups.GetChildren(notification_group_id, "action", nil)
	if nil != err {
		return errors.New("load notfications failed, " + err.Error())
	}
	if nil == rules {
		return nil
	}

	args := map[string]interface{}{}
	for k, v := range evt {
		args[k] = v
	}
	payload_object := map[string]interface{}{"type": "multiplexed", "arguments": args,
		"rules": rules}
	if 0 < *notification_max_attempts {
		payload_object["max_attempts"] = *notification_max_attempts
	}

	bs, err := json.Marshal(payload_object)
	if nil != err {
		return errors.New("marshal payload_object failed," + err.Error())
	}
	if nil != bs {
		evt["notification"] = map[string]interface{}{"priority": *notification_priority, "queue": *notification_queue, "payload_object": string(bs)}
	}
	return nil
}

func (self *alertAction) gen_message(current, previous int, evt map[string]interface{}) string {
	if len(self.templates) > current {
		var buffer bytes.Buffer
		e := self.templates[current].Execute(&buffer, evt)
		if nil == e {
			return buffer.String()
		}
		self.information = "execute template failed, " + e.Error()
	}

	switch current {
	case 0:
		return fmt.Sprintf("%v is resumed", self.name)
	case 1:
		return fmt.Sprintf("%v is alerted", self.name)
	default:
		return fmt.Sprintf("%v is alerted, status is %v", self.name, current)
	}
}

func (self *alertAction) send(evt map[string]interface{}) error {
	bs, e := json.Marshal(evt)
	if nil != e {
		return errors.New("marshal alert_event failed, " + e.Error())
	}

	atomic.StoreInt64(&self.begin_send_at, 0)
	atomic.StoreInt64(&self.wait_response_at, 0)
	atomic.StoreInt64(&self.responsed_at, 0)
	atomic.StoreInt64(&self.end_send_at, 0)

	self.cached_data.attributes = evt
	atomic.StoreInt64(&self.begin_send_at, time.Now().Unix())
	self.channel <- self.cached_data
	atomic.StoreInt64(&self.wait_response_at, time.Now().Unix())
	e = <-self.cached_data.c
	atomic.StoreInt64(&self.responsed_at, time.Now().Unix())
	if nil == e {
		self.publish <- []string{"PUBLISH", "tpt_alert_events", string(bs)}
	}
	atomic.StoreInt64(&self.end_send_at, time.Now().Unix())
	return e
}

var (
	ExpressionStyleIsRequired    = commons.IsRequired("expression_style")
	ExpressionCodeIsRequired     = commons.IsRequired("expression_code")
	NotificationChannelIsNil     = errors.New("'alerts_channel' is nil")
	NotificationChannelTypeError = errors.New("'alerts_channel' is not a chan<- *data_object ")
)

func newAlertAction(attributes, options, ctx map[string]interface{}) (ExecuteAction, error) {
	id, e := commons.GetInt64(attributes, "id")
	if nil != e || 0 == id {
		return nil, IdIsRequired
	}

	name, e := commons.GetString(attributes, "name")
	if nil != e {
		return nil, NameIsRequired
	}

	c := ctx["alerts_channel"]
	if nil == c {
		return nil, NotificationChannelIsNil
	}
	channel, ok := c.(chan<- *data_object)
	if !ok {
		return nil, NotificationChannelTypeError
	}

	c = ctx["redis_channel"]
	if nil == c {
		return nil, errors.New("'redis_channel' is nil")
	}
	publish, ok := c.(chan<- []string)
	if !ok {
		return nil, errors.New("'redis_channel' is not a chan []stirng")
	}

	checker, e := makeChecker(attributes, ctx)
	if nil != e {
		return nil, e
	}

	delay_times := commons.GetIntWithDefault(attributes, "delay_times", 1)
	if delay_times <= 0 {
		delay_times = 1
	}

	if delay_times >= MAX_REPEATED {
		delay_times = MAX_REPEATED - 20
	}

	contex := map[string]interface{}{"action_id": id, "name": name}
	if nil != options {
		for k, v := range options {
			contex[k] = v
		}
	}

	var notification_group_ids []string
	notification_group_ids_str := commons.GetStringWithDefault(attributes, "notification_group_ids", "")
	var notification_groups *ds.Cache
	if 0 != len(notification_group_ids_str) {
		if c, ok := ctx["notification_groups"]; ok && nil != c {
			notification_groups, _ = c.(*ds.Cache)
		}
		if nil == notification_groups {
			return nil, errors.New("'notification_groups' is missing")
		}
		notification_group_ids = strings.Split(notification_group_ids_str, ",")
		for _, s := range notification_group_ids {
			if _, e = strconv.ParseInt(s, 10, 0); nil != e {
				return nil, errors.New("parse 'notification_group_ids' failed, it is not a int array - '" + notification_group_ids_str + "'")
			}
		}
	}

	templates, e := loadTemplates(ctx, attributes, "templates")
	if nil != e {
		return nil, e
	}

	return &alertAction{id: id,
		name: name,
		//description: commons.GetString(attributes, "description", ""),
		already_send:           true,
		options:                options,
		delay_times:            delay_times,
		contex:                 contex,
		publish:                publish,
		channel:                channel,
		cached_data:            &data_object{c: make(chan error, 2)},
		checker:                checker,
		templates:              templates,
		last_status:            commons.GetIntWithDefault(attributes, "last_status", 0),
		previous_status:        commons.GetIntWithDefault(attributes, "previous_status", 0),
		last_event_id:          commons.GetStringWithDefault(attributes, "event_id", ""),
		sequence_id:            commons.GetIntWithDefault(attributes, "seqence_id", 0),
		notification_group_ids: notification_group_ids,
		notification_groups:    notification_groups}, nil
}

func makeChecker(attributes, ctx map[string]interface{}) (Checker, error) {
	style, e := commons.GetString(attributes, "expression_style")
	if nil != e {
		return nil, ExpressionStyleIsRequired
	}

	code, e := commons.GetString(attributes, "expression_code")
	if nil != e {
		codeObject, e := commons.GetObject(attributes, "expression_code")
		if nil != e {
			return nil, ExpressionCodeIsRequired
		}

		codeBytes, e := json.Marshal(codeObject)
		if nil != e {
			return nil, ExpressionCodeIsRequired
		}

		code = string(codeBytes)
	}

	switch style {
	case "json":
		return makeJsonChecker(code)
	}
	return nil, errors.New("expression style '" + style + "' is unknown")
}

func abs(s string) string {
	r, e := filepath.Abs(s)
	if nil != e {
		return s
	}
	return r
}

func get_alerts_template_path() string {
	dirs := []string{filepath.Join(abs(filepath.Dir(os.Args[0])), "lib/alerts/templates"),
		filepath.Join(abs(filepath.Dir(os.Args[0])), "../lib/alerts/templates"),
		filepath.Join(abs("."), "lib/alerts/templates"),
		filepath.Join(abs("."), "../lib/alerts/templates"),
		filepath.Join(abs(filepath.Dir(os.Args[0])), "conf/alerts_templates"),
		filepath.Join(abs(filepath.Dir(os.Args[0])), "../conf/alerts_templates"),
		filepath.Join(abs("."), "conf/alerts_templates"),
		filepath.Join(abs("."), "../conf/alerts_templates")}
	for _, s := range dirs {
		fi, e := os.Stat(s)
		if nil == e && fi.IsDir() {
			return s
		}
	}
	return ""
}

func loadTemplates(ctx, args map[string]interface{}, key string) ([]*template.Template, error) {
	templates, e := templatesWith(args, key)
	if nil != e {
		return nil, e
	}
	if 0 != len(templates) {
		return templates, nil
	}

	pa := commons.GetStringWithDefault(ctx, "alerts_template_path", get_alerts_template_path())
	if 0 == len(pa) {
		return []*template.Template{}, nil
	}

	cat := commons.GetStringWithDefault(args, "catalog", "")
	if 0 != len(cat) {
		templates, e = loadTemplatesFromDir(pa, cat)
		if nil != e {
			return nil, e
		}
		if 0 != len(templates) {
			return templates, nil
		}
	}

	return loadTemplatesFromDir(pa, "default")
}

func loadTemplatesFromDir(pa, prefix string) (templates []*template.Template, err error) {
	for i := 0; i < 9999; i++ {
		nm := filepath.Clean(filepath.Join(pa, prefix+"_"+strconv.FormatInt(int64(i), 10)+".tpl"))
		if fi, e := os.Stat(nm); nil != e || fi.IsDir() {
			break
		}
		t, e := template.ParseFiles(nm)
		if nil != e {
			return nil, fmt.Errorf("load message templates of alerts failed, parse '%v' failed, %v", nm, e.Error())
		}
		templates = append(templates, t)
	}
	if nil == templates {
		templates = []*template.Template{}
	}
	return
}

func templatesWith(args map[string]interface{}, key string) ([]*template.Template, error) {
	v, ok := args[key]
	if !ok {
		return []*template.Template{}, nil
	}
	var e error
	if s, ok := v.(string); ok {
		if 0 == len(strings.TrimSpace(s)) {
			return []*template.Template{}, nil
		}
		var ss []string
		e = json.Unmarshal([]byte(s), &ss)
		if nil != e {
			return nil, fmt.Errorf("load message templates of alerts failed, %v is not a valid json string, %s - `%v`", key, e.Error(), string(s))
		}
		if nil == ss || 0 == len(ss) {
			return []*template.Template{}, nil
		}
		tt := make([]*template.Template, len(ss))
		for i, s := range ss {
			tt[i], e = template.New("default").Parse(s)
			if nil != e {
				return nil, fmt.Errorf("load message templates of alerts failed, parse %v[%v] failed, %v", key, i, e.Error())
			}
		}
		return tt, nil
	}

	if ii, ok := v.([]interface{}); ok {
		tt := make([]*template.Template, len(ii))
		for i, o := range ii {
			s, ok := o.(string)
			if !ok {
				return nil, fmt.Errorf("load message templates of alerts failed, %v[%v] is not a string", key, i)
			}
			tt[i], e = template.New("default").Parse(s)
			if nil != e {
				return nil, fmt.Errorf("load message templates of alerts failed, parse %v[%v] failed, %v", key, i, e.Error())
			}
		}
		return tt, nil
	}
	if ss, ok := v.([]string); ok {
		tt := make([]*template.Template, len(ss))
		for i, s := range ss {
			tt[i], e = template.New("default").Parse(s)
			if nil != e {
				return nil, fmt.Errorf("load message templates of alerts failed, parse %v[%v] failed, %v", key, i, e.Error())
			}
		}
		return tt, nil
	}
	return []*template.Template{}, nil
}
