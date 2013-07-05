package poller

import (
	"commons"
	"errors"
	"time"
)

type alertAction struct {
	name        string
	maxRepeated int

	options map[string]interface{}
	result  map[string]interface{}
	channel chan map[string]interface{}

	checker    Checker
	lastStatus int
	repeated   int
}

func (self *alertAction) Run(t time.Time, value interface{}) error {
	current, err := self.checker.Run(value, self.result)
	if nil != err {
		return err
	}

	if self.repeated == self.maxRepeated {
		evt := map[string]interface{}{}
		for k, v := range self.result {
			evt[k] = v
		}

		if _, found := evt["triggered_at"]; !found {
			evt["triggered_at"] = t
		}

		evt["status"] = current
		self.channel <- evt
	}

	if current == self.lastStatus {
		self.repeated++

		if self.repeated > 2147483646 || self.repeated < 0 { // inhebit overflow
			self.repeated = self.maxRepeated + 10
		}
	} else {
		self.repeated = 0
		self.lastStatus = current
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
	channel, ok := c.(chan map[string]interface{})
	if !ok {
		return nil, NotificationChannelTypeError
	}

	checker, e := makeChecker(attributes, ctx)
	if nil != e {
		return nil, e
	}

	return &alertAction{name: name,
		//description: commons.GetString(attributes, "description", ""),
		options:     options,
		maxRepeated: commons.GetIntWithDefault(attributes, "max_repeated", 0),
		result:      map[string]interface{}{"name": name},
		channel:     channel,
		checker:     checker}, nil
}

func makeChecker(attributes, ctx map[string]interface{}) (Checker, error) {
	style, e := commons.GetString(attributes, "expression_style")
	if nil != e {
		return nil, ExpressionStyleIsRequired
	}

	code, e := commons.GetString(attributes, "expression_code")
	if nil != e {
		return nil, ExpressionCodeIsRequired
	}

	switch style {
	case "json":
		return makeJsonChecker(code)
	}
	return nil, errors.New("expression style '" + style + "' is unknown")
}
