package poller

import (
	"commons"
	"errors"
	"sync/atomic"
	"time"
)

type historyAction struct {
	id           int64
	name         string
	description  string
	metric       interface{}
	managed_id   interface{}
	managed_type interface{}
	trigger_id   interface{}
	channel      chan<- *data_object
	cached_data  *data_object
	attribute    string
	merger       aggregator

	begin_send_at, wait_response_at, end_send_at int64
}

func (self *historyAction) Stats() map[string]interface{} {
	return map[string]interface{}{
		"type":             "history",
		"id":               self.id,
		"name":             self.name,
		"begin_send_at":    atomic.LoadInt64(&self.begin_send_at),
		"wait_response_at": atomic.LoadInt64(&self.wait_response_at),
		"end_send_at":      atomic.LoadInt64(&self.end_send_at)}
}

func (self *historyAction) RunBefore() {
}

func (self *historyAction) RunAfter() {
}

func (self *historyAction) Run(t time.Time, value interface{}) error {
	created_at := t
	if current, ok := value.(commons.Result); ok {
		if current.HasError() {
			return errors.New("sampling failed, " + current.ErrorMessage())
		}
		created_at = current.CreatedAt()
	}

	currentValue, e := commons.ToSimpleValue(value, self.attribute)
	if nil != e {
		return e
	}
	currentValue, ok, e := self.merger.aggregate(currentValue, created_at)
	if nil != e {
		return e
	}
	if !ok {
		return nil
	}
	self.cached_data.attributes = map[string]interface{}{
		"action_id":    self.id,
		"sampled_at":   created_at,
		"metric":       self.metric,
		"managed_type": self.managed_type,
		"managed_id":   self.managed_id,
		"trigger_id":   self.trigger_id,
		"value":        currentValue}

	atomic.StoreInt64(&self.begin_send_at, time.Now().Unix())
	self.channel <- self.cached_data
	atomic.StoreInt64(&self.wait_response_at, time.Now().Unix())
	e = <-self.cached_data.c
	atomic.StoreInt64(&self.end_send_at, time.Now().Unix())
	return e
}

func newHistoryAction(attributes, options, ctx map[string]interface{}) (ExecuteAction, error) {
	id, e := commons.GetInt64(attributes, "id")
	if nil != e || 0 == id {
		return nil, IdIsRequired
	}

	name, e := commons.GetString(attributes, "name")
	if nil != e {
		return nil, NameIsRequired
	}

	attribute, e := commons.GetString(attributes, "attribute")
	if nil != e {
		return nil, CommandIsRequired
	}

	c := ctx["histories_channel"]
	if nil == c {
		return nil, errors.New("'histories_channel' is nil")
	}
	channel, ok := c.(chan<- *data_object)
	if !ok {
		return nil, errors.New("'histories_channel' is not a chan<- *data_object")
	}

	managed_type := options["managed_type"]
	managed_id := options["managed_id"]
	triggger_id := options["trigger_id"]
	interval := 4 * time.Minute
	if v, ok := options["interval"].(time.Duration); ok {
		interval = mergeInterval(v)
	}
	metric := options["metric"]

	return &historyAction{id: id,
		name:         name,
		description:  commons.GetStringWithDefault(attributes, "description", ""),
		channel:      channel,
		cached_data:  &data_object{c: make(chan error, 2)},
		attribute:    attribute,
		merger:       &last_merger{interval: interval},
		metric:       metric,
		managed_id:   managed_id,
		managed_type: managed_type,
		trigger_id:   triggger_id}, nil
}

func mergeInterval(v time.Duration) time.Duration {
	if v <= 15*time.Second {
		return v - 2*time.Second
	}
	if v <= 30*time.Second {
		return v - 5*time.Second
	}
	if v <= 60*time.Second {
		return v - 15*time.Second
	}
	if v <= 120*time.Second {
		return v - 30*time.Second
	}
	if v <= 120*time.Second {
		return v - 30*time.Second
	}
	if v <= 5*time.Minute {
		return v - 1*time.Minute
	}
	if v <= 10*time.Minute {
		return v - 3*time.Minute
	}
	return v - 5*time.Minute
}
