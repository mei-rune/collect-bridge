package poller

import (
	"commons"
	"encoding/json"
	"errors"
	"time"
)

const MAX_REPEATED = 9999990

type alertAction struct {
	name         string
	max_repeated int

	options map[string]interface{}
	result  map[string]interface{}
	channel chan<- map[string]interface{}

	checker     Checker
	last_status int
	repeated    int
}

func (self *alertAction) Run(t time.Time, value interface{}) error {
	current, err := self.checker.Run(value, self.result)
	if nil != err {
		return err
	}

	if current == self.last_status {
		self.repeated++

		if self.repeated >= 9999996 || self.repeated < 0 { // inhebit overflow
			self.repeated = self.max_repeated + 10
		}
	} else {
		self.repeated = 1
		self.last_status = current
	}

	if self.repeated == self.max_repeated {

		evt := map[string]interface{}{}
		for k, v := range self.result {
			evt[k] = v
		}
		if nil != self.options {
			for k, v := range self.options {
				evt[k] = v
			}
		}

		if _, found := evt["triggered_at"]; !found {
			evt["triggered_at"] = t
		}

		if _, found := evt["current_value"]; !found {
			evt["current_value"] = value
		}

		evt["status"] = current
		self.channel <- evt
	}

	return nil
}

var (
	ExpressionStyleIsRequired    = commons.IsRequired("expression_style")
	ExpressionCodeIsRequired     = commons.IsRequired("expression_code")
	NotificationChannelIsNil     = errors.New("'notification_channel' is nil")
	NotificationChannelTypeError = errors.New("'notification_channel' is not a chan map[string]interface{}")
)

func newAlertAction(attributes, options, ctx map[string]interface{}) (ExecuteAction, error) {
	name, e := commons.GetString(attributes, "name")
	if nil != e {
		return nil, NameIsRequired
	}

	c := ctx["notification_channel"]
	if nil == c {
		return nil, NotificationChannelIsNil
	}
	channel, ok := c.(chan<- map[string]interface{})
	if !ok {
		return nil, NotificationChannelTypeError
	}

	checker, e := makeChecker(attributes, ctx)
	if nil != e {
		return nil, e
	}

	max_repeated := commons.GetIntWithDefault(attributes, "max_repeated", 1)
	if max_repeated <= 0 {
		max_repeated = 1
	}

	if max_repeated >= MAX_REPEATED {
		max_repeated = MAX_REPEATED - 20
	}

	return &alertAction{name: name,
		//description: commons.GetString(attributes, "description", ""),
		options:      options,
		max_repeated: max_repeated,
		result:       map[string]interface{}{"name": name},
		channel:      channel,
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
