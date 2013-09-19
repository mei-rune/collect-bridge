package poller

import (
	"bytes"
	"commons"
	ds "data_store"
	"fmt"
	"log"
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

	specific_status_names = make([]string, CLOSE_REASON_MAX+1)
)

func init() {
	specific_status_names[CLOSE_REASON_DISABLED] = "disabled"
	specific_status_names[CLOSE_REASON_DELETED] = "deleted"
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

	templates          []*template.Template
	specific_templates []*template.Template
	informations       commons.CircularBuffer
	checker            Checker
	previous_status    int
	last_status        int
	last_value         interface{}
	repeated           int
	already_send       bool
	last_event_id      string
	sequence_id        int
	level              int

	notification_group_ids []string
	notification_groups    *ds.Cache

	begin_send_at, wait_response_at, responsed_at, end_send_at int64

	stats_last_value      interface{}
	stats_previous_status int
	stats_last_status     int
	stats_repeated        int
	stats_already_send    bool
	stats_last_event_id   string
	stats_sequence_id     int
	stats_informations    []interface{}
}

func (self *alertAction) Stats() map[string]interface{} {
	stats := map[string]interface{}{
		"type":             "alert",
		"id":               self.id,
		"name":             self.name,
		"information":      self.stats_informations,
		"previous_status":  self.stats_previous_status,
		"last_status":      self.stats_last_status,
		"last_value":       self.stats_last_value,
		"repeated":         self.stats_repeated,
		"already_send":     self.stats_already_send,
		"event_id":         self.stats_last_event_id,
		"sequence_id":      self.stats_sequence_id,
		"begin_send_at":    time.Unix(atomic.LoadInt64(&self.begin_send_at), 0),
		"wait_response_at": time.Unix(atomic.LoadInt64(&self.wait_response_at), 0),
		"responsed_at":     time.Unix(atomic.LoadInt64(&self.responsed_at), 0),
		"end_send_at":      time.Unix(atomic.LoadInt64(&self.end_send_at), 0)}

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
	self.stats_informations = self.informations.All()
	self.stats_last_value = self.last_value
}

func (self *alertAction) Run(t time.Time, value interface{}) error {
	if res, ok := value.(ValueResult); ok {
		if res.HasError() {
			return errors.New("sampling failed, " + res.ErrorMessage())
		}
	}

	current, current_value, err := self.checker.Run(value, self.contex)
	if nil != err {
		return err
	}
	self.last_value = current_value

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

	evt, err := self.create_event(current_value, current, CLOSE_REASON_NORMAL, t)
	if nil != err {
		return err
	}

	err = self.send(evt, false)
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

func (self *alertAction) create_event(current_value interface{}, current, reason int, t time.Time) (map[string]interface{}, error) {
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
			return nil, errors.New("marshal current value failed," + err.Error())
		}
		if nil != bs {
			evt["current_value"] = string(bs)
		}
	}

	if 0 == self.previous_status {
		self.sequence_id = 1
		self.last_event_id = commons.GenerateId()
	}

	evt["level"] = self.level
	evt["event_id"] = self.last_event_id
	evt["sequence_id"] = self.sequence_id
	evt["previous_status"] = self.previous_status
	evt["status"] = current
	evt["content"] = self.gen_message(current, self.previous_status, reason, evt)
	if nil != self.notification_groups {
		err := self.fillNotificationData(current, evt)
		if nil != err {
			return nil, err
		}
	}

	return evt, nil
}

func (self *alertAction) Reset(reason int) error {
	if 0 == self.previous_status {
		self.repeated = 0
		return nil
	}

	evt, err := self.create_event(0, 0, reason, time.Now())
	if nil != err {
		return err
	}

	err = self.send(evt, true)
	if nil == err {
		self.last_status = 0
		self.previous_status = 0
		self.sequence_id++
		self.already_send = true
		self.repeated = 0
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
	if nil == rules || 0 == len(rules) {
		self.informations.Push("notfications is empty.")
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

func (self *alertAction) gen_message(current, previous, reason int, evt map[string]interface{}) string {
	if reason >= 0 {
		if nil != self.specific_templates && len(self.specific_templates) > reason && nil != self.specific_templates[reason] {
			var buffer bytes.Buffer
			e := self.specific_templates[reason].Execute(&buffer, evt)
			if nil == e {
				return buffer.String()
			}
			fmt.Println("execute template failed, " + e.Error())
			self.informations.Push("execute template failed, " + e.Error())
		}
		switch reason {
		case CLOSE_REASON_DELETED:
			return fmt.Sprintf("%v is deleted", self.name)
		case CLOSE_REASON_DISABLED:
			return fmt.Sprintf("%v is disabled", self.name)
		default:
			return fmt.Sprintf("%v is reset, reason is %v, status is %v", self.name, reason, current)
		}
	} else {
		if nil != self.templates && len(self.templates) > current {
			var buffer bytes.Buffer
			e := self.templates[current].Execute(&buffer, evt)
			if nil == e {
				return buffer.String()
			}
			self.informations.Push("execute template failed, " + e.Error())
		}
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

func (self *alertAction) send(evt map[string]interface{}, nonresponse bool) error {
	bs, e := json.Marshal(evt)
	if nil != e {
		return errors.New("marshal alert_event failed, " + e.Error())
	}

	atomic.StoreInt64(&self.begin_send_at, 0)
	atomic.StoreInt64(&self.wait_response_at, 0)
	atomic.StoreInt64(&self.responsed_at, 0)
	atomic.StoreInt64(&self.end_send_at, 0)

	if nonresponse {
		atomic.StoreInt64(&self.begin_send_at, time.Now().Unix())
		self.channel <- &data_object{c: make(chan error, 2), attributes: evt}
		at := time.Now().Unix()
		atomic.StoreInt64(&self.wait_response_at, at)
		atomic.StoreInt64(&self.responsed_at, at)
		self.publish <- []string{"PUBLISH", "tpt_alert_events", string(bs)}
		atomic.StoreInt64(&self.end_send_at, at)
	} else {
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
	}
	return e
}

var (
	ExpressionStyleIsRequired    = commons.IsRequired("expression_style")
	ExpressionCodeIsRequired     = commons.IsRequired("expression_code")
	NotificationChannelIsNil     = errors.New("'alerts_channel' is nil")
	NotificationChannelTypeError = errors.New("'alerts_channel' is not a chan<- *data_object ")

	empty_cookies = map[string]interface{}{}
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

	templates, specific_templates, e := loadTemplates(ctx, attributes, "templates")
	if nil != e {
		return nil, e
	}

	var cookies map[string]interface{} = nil
	if v, ok := ctx["cookies_loader"]; ok {
		if loader, ok := v.(cookiesLoader); ok {
			c, e := loader.loadCookiesWithAcitonId(id, ctx)
			if nil != e {
				return nil, errors.New("load alert cookies with id was " +
					strconv.FormatInt(int64(id), 10) +
					" and name is '" + name + "' failed, " + e.Error())
			}
			cookies = c
		}
	}

	if nil == cookies {
		cookies = empty_cookies
		log.Println("load alert cookies with id was " + strconv.FormatInt(int64(id), 10) + " and name is '" + name + "' is not found")
	} else {
		log.Println("load alert cookies with id was " + strconv.FormatInt(int64(id), 10) + " and name is '" + name + "' is ok")
	}

	action := &alertAction{id: id,
		name:                   name,
		level:                  commons.GetIntWithDefault(attributes, "level", 1),
		already_send:           true,
		options:                options,
		delay_times:            delay_times,
		contex:                 contex,
		publish:                publish,
		channel:                channel,
		cached_data:            &data_object{c: make(chan error, 2)},
		checker:                checker,
		templates:              templates,
		specific_templates:     specific_templates,
		last_status:            commons.GetIntWithDefault(cookies, "status", 0),
		previous_status:        commons.GetIntWithDefault(cookies, "previous_status", 0),
		last_event_id:          commons.GetStringWithDefault(cookies, "event_id", ""),
		sequence_id:            commons.GetIntWithDefault(cookies, "sequence_id", 0),
		notification_group_ids: notification_group_ids,
		notification_groups:    notification_groups}

	if 0 != action.last_status {
		action.previous_status = action.last_status
		action.sequence_id++
		action.already_send = true
	}

	action.informations.Init(make([]interface{}, 10))
	return action, nil
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

func loadTemplates(ctx, args map[string]interface{}, key string) (templates []*template.Template,
	specific_templates []*template.Template, e error) {
	templates, e = templatesWith(args, key)
	if nil != e {
		return nil, nil, e
	}

	pa := commons.GetStringWithDefault(ctx, "alerts_template_path", get_alerts_template_path())
	if 0 == len(pa) {
		return templates, nil, nil
	}

	cat_list := make([]string, 0, 2)
	cat := commons.GetStringWithDefault(args, "catalog", "")
	if 0 != len(cat) {
		cat_list = append(cat_list, cat)
	}
	cat_list = append(cat_list, "default")

	if nil == templates || 0 == len(templates) {
		for _, cat := range cat_list {
			templates, e = loadTemplatesFromDir(pa, cat)
			if nil != e {
				return nil, nil, e
			}

			if nil != templates && 0 != len(templates) {
				break
			}
		}
	}

	specific_templates = make([]*template.Template, len(specific_status_names))
	for _, cat := range cat_list {
		e = loadSpecificTemplatesFromDir(pa, cat, specific_templates)
		if nil != e {
			return nil, nil, e
		}
	}

	return templates, specific_templates, nil
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
	return
}

func loadSpecificTemplatesFromDir(pa, prefix string, templates []*template.Template) (err error) {
	for i, nm := range specific_status_names {
		if nil != templates[i] {
			continue
		}
		nm := filepath.Clean(filepath.Join(pa, prefix+"_"+nm+".tpl"))
		if fi, e := os.Stat(nm); nil != e || fi.IsDir() {
			continue
		}
		t, e := template.ParseFiles(nm)
		if nil != e {
			return fmt.Errorf("load message templates of alerts failed, parse '%v' failed, %v", nm, e.Error())
		}
		templates[i] = t
	}
	return nil
}

func templatesWith(args map[string]interface{}, key string) ([]*template.Template, error) {
	v, ok := args[key]
	if !ok {
		return nil, nil
	}
	var e error
	if s, ok := v.(string); ok {
		if 0 == len(strings.TrimSpace(s)) {
			return nil, nil
		}
		var ss []string
		e = json.Unmarshal([]byte(s), &ss)
		if nil != e {
			return nil, fmt.Errorf("load message templates of alerts failed, %v is not a valid json string, %s - `%v`", key, e.Error(), string(s))
		}
		if nil == ss || 0 == len(ss) {
			return nil, nil
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
	return nil, nil
}
