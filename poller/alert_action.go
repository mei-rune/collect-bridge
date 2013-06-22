package poller

import (
	"bytes"
	"commons"
	"fmt"
	"runtime"
	"time"
)

type AlertAction struct {
	name        string
	maxRepeated int

	result  map[string]interface{}
	channel chan map[string]interface{}

	checker    Checker
	lastStatus int
	repeated   int
	last_error error
}

func (self *AlertAction) Run(t time.Time, value interface{}) {
	defer func() {
		if e := recover(); nil != e {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("[panic][alert][%s]%v", self.name, e))
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}
			commons.Log.ERROR.Print(buffer.String())
		}
	}()

	current, e := self.checker.Run(value, self.result)
	if nil != e {
		if nil == self.last_error {
			commons.Log.ERROR.Print("[error]" + self.name + " - " + e.Error())
		}
		self.last_error = e
		return
	} else {

		if nil != self.last_error {
			commons.Log.ERROR.Print("[error]" + self.name + " is ok ")
		}

		self.last_error = nil
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
}

var (
	ExpressionStyleIsRequired    = commons.IsRequired("expression_style")
	ExpressionCodeIsRequired     = commons.IsRequired("expression_code")
	NotificationChannelIsNil     = commons.BadRequest("'notification_channel' is nil")
	NotificationChannelTypeError = commons.BadRequest("'notification_channel' is not a chan map[string]interface{}")
)

func NewAlertAction(attributes, ctx map[string]interface{}) (ExecuteAction, error) {
	name, e := commons.TryGetString(attributes, "name")
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

	return &AlertAction{name: name,
		//description: commons.GetString(attributes, "description", ""),
		maxRepeated: commons.GetInt(attributes, "max_repeated", 0),
		result:      map[string]interface{}{"name": name},
		channel:     channel,
		checker:     checker}, nil
}

func makeChecker(attributes, ctx map[string]interface{}) (Checker, error) {
	style, e := commons.TryGetString(attributes, "expression_style")
	if nil != e {
		return nil, ExpressionStyleIsRequired
	}

	code, e := commons.TryGetString(attributes, "expression_code")
	if nil != e {
		return nil, ExpressionCodeIsRequired
	}

	switch style {
	case "json":
		return makeJsonChecker(code)
	}
	return nil, commons.BadRequest("expression style '" + style + "' is unknown")
}
