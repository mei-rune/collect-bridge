package poller

import (
	"commons"
	"encoding/json"
	"errors"
	"sync/atomic"
	"time"
)

const MAX_REPEATED = 9999990

var reset_error = errors.New("please reset channel.")

type alertAction struct {
	id          int64
	name        string
	delay_times int

	options     map[string]interface{}
	contex      map[string]interface{}
	publish     chan<- []string
	channel     chan<- *data_object
	cached_data *data_object

	checker      Checker
	last_status  int
	repeated     int
	already_send bool

	begin_send_at, wait_response_at, responsed_at, end_send_at int64

	stats_last_status  int
	stats_repeated     int
	stats_already_send bool
}

func (self *alertAction) Stats() map[string]interface{} {
	return map[string]interface{}{
		"type":             "alert",
		"id":               self.id,
		"name":             self.name,
		"last_status":      self.stats_last_status,
		"repeated":         self.stats_repeated,
		"already_send":     self.stats_already_send,
		"begin_send_at":    atomic.LoadInt64(&self.begin_send_at),
		"wait_response_at": atomic.LoadInt64(&self.wait_response_at),
		"responsed_at":     atomic.LoadInt64(&self.responsed_at),
		"end_send_at":      atomic.LoadInt64(&self.end_send_at)}
}

func (self *alertAction) RunBefore() {
}

func (self *alertAction) RunAfter() {
	self.stats_last_status = self.last_status
	self.stats_repeated = self.repeated
	self.stats_already_send = self.already_send
}

func (self *alertAction) Run(t time.Time, value interface{}) error {
	current, err := self.checker.Run(value, self.contex)
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
		bs, _ := json.Marshal(value)
		if nil != bs {
			evt["current_value"] = string(bs)
		}
	}

	evt["status"] = current

	err = self.send(evt)
	if nil == err {
		self.already_send = true
		return nil
	}

	if err == reset_error {
		self.cached_data = &data_object{c: make(chan error, 2)}
	}
	return err
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

	return &alertAction{id: id,
		name: name,
		//description: commons.GetString(attributes, "description", ""),
		already_send: false,
		options:      options,
		delay_times:  delay_times,
		contex:       contex,
		publish:      publish,
		channel:      channel,
		cached_data:  &data_object{c: make(chan error, 2)},
		checker:      checker}, nil
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
