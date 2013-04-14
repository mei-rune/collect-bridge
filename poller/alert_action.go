package poller

import (
	"commons"
	"fmt"
	"time"

	"errors"
)

type AlertAction struct {
	Name        string
	Description string
	MaxRepeated int

	result  map[string]interface{}
	channel chan map[string]interface{}

	checker    Checker
	lastStatus int
	repeated   int
}

func (self *AlertAction) Run(t time.Time, value interface{}) {
	defer func() {
		if e := recover(); nil != e {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("[panic][alert][%s]%v", self.Name, e))
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

	current := self.checker.Run(t, value, self.result)

	if current == self.lastStatus {
		self.repeated++

		if self.repeated > 2147483646 || self.repeated < 0 { // inhebit overflow
			self.repeated = self.MaxRepeated + 10
		}
	} else {
		self.repeated = 0
		self.lastStatus = current
	}

	if self.repeated < self.MaxRepeated {
		return
	}

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

func NewAlertAction(attributes, ctx map[string]interface{}) (ExecuteAction, error) {
	name, e := commons.TryGetString(attributes, "name")
	if nil != e {
		return nil, NameIsRequired
	}

	c := ctx["notification_channel"]
	if nil == c {
		return nil, errors.New("'notification_channel' is nil")
	}
	channel, ok := c.(chan map[string]interface{})
	if !ok {
		return nil, errors.New("'notification_channel' is not a chan map[string]interface{}")
	}

	checker, e := makeChecker(attributes)
	if nil != e {
		return nil, fmt.Errorf("create checker failed - %v", e)
	}

	return &AlertAction{Name: name,
		Description: commons.GetString(attributes, "description", ""),
		MaxRepeated: commons.GetInt(attributes, "max_repeated", 0),
		result:      map[string]interface{}{"name": name},
		channel:     channel,
		checker:     checker}, nil
}

func makeChecker(attributes, ctx map[string]interface{}) Checker {

}
