package poller

import (
	"commons"
	"errors"
	"time"
)

type historyAction struct {
	id           string
	name         string
	description  string
	metric       interface{}
	managed_id   interface{}
	managed_type interface{}
	trigger_id   interface{}
	channel      chan<- map[string]interface{}
	attribute    string
}

func (self *historyAction) Run(t time.Time, value interface{}) error {

	created_at := t
	if current, ok := value.(commons.Result); ok {
		created_at = current.CreatedAt()
	}

	currentValue, e := commons.ToSimpleValue(value, self.attribute)
	if nil != e {
		return e
	}

	self.channel <- map[string]interface{}{
		"action_id":    self.id,
		"sampling_at":  created_at,
		"metric":       self.metric,
		"managed_type": self.managed_type,
		"managed_id":   self.managed_id,
		"trigger_id":   self.trigger_id,
		"value":        currentValue}
	return nil
}

func newHistoryAction(attributes, options, ctx map[string]interface{}) (ExecuteAction, error) {
	id, e := commons.GetString(attributes, "id")
	if nil != e || 0 == len(id) {
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

	c := ctx["history_channel"]
	if nil == c {
		return nil, errors.New("'history_channel' is nil")
	}
	channel, ok := c.(chan<- map[string]interface{})
	if !ok {
		return nil, errors.New("'history_channel' is not a chan []stirng")
	}

	managed_type := options["managed_type"]
	managed_id := options["managed_id"]
	triggger_id := options["trigger_id"]
	metric := options["metric"]

	return &historyAction{id: id,
		name:         name,
		description:  commons.GetStringWithDefault(attributes, "description", ""),
		channel:      channel,
		attribute:    attribute,
		metric:       metric,
		managed_id:   managed_id,
		managed_type: managed_type,
		trigger_id:   triggger_id}, nil
}
