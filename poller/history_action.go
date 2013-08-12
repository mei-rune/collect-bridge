package poller

import (
	"commons"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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

	switch v := currentValue.(type) {
	case json.Number:
		currentValue, e = v.Float64()
		if nil != e {
			currentValue, e = v.Int64()
			if nil != e {
				return errors.New("crazy! it is not a number? - " + v.String())
			}
		}
	case int64:
		break
	case int:
		break
	case float64:
		break
	case float32:
		break
	case string:
		currentValue, e = strconv.ParseFloat(v, 10)
		if nil != e {
			return errors.New("crazy! it is not a number? - " + v)
		}
	case bool:
		if v {
			currentValue = 1
		} else {
			currentValue = 0
		}
	case int32:
		break
	case int8:
		break
	case int16:
		break
	case uint:
		break
	case uint8:
		break
	case uint16:
		break
	case uint32:
		break
	case uint64:
		break
	case uintptr:
		break
	default:
		return fmt.Errorf("value is not a number- %T", v)
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
	metric := options["metric"]

	return &historyAction{id: id,
		name:         name,
		description:  commons.GetStringWithDefault(attributes, "description", ""),
		channel:      channel,
		cached_data:  &data_object{c: make(chan error, 2)},
		attribute:    attribute,
		metric:       metric,
		managed_id:   managed_id,
		managed_type: managed_type,
		trigger_id:   triggger_id}, nil
}
