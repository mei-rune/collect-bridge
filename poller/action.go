package poller

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"time"
)

type ExecuteAction interface {
	RunBefore()
	Run(t time.Time, value interface{}) error
	RunAfter()

	Stats() map[string]interface{}
}

func newAction(attributes, options, ctx map[string]interface{}) (ExecuteAction, error) {
	switch attributes["type"] {
	case "redis_command":
		return newRedisAction(attributes, options, ctx)
	case "alert":
		return newAlertAction(attributes, options, ctx)
	case "history":
		return newHistoryAction(attributes, options, ctx)
	case "test":
		return newTestAction(attributes, options, ctx)
	}
	return nil, fmt.Errorf("unsupported type, - %v", attributes["type"])
}

type testAction struct {
	stats map[string]interface{}
	run   func(t time.Time, value interface{}) error
}

func (self *testAction) RunBefore() {
}

func (self *testAction) Run(t time.Time, value interface{}) error {
	return self.run(t, value)
}

func (self *testAction) RunAfter() {
}

func (self *testAction) Stats() map[string]interface{} {
	return self.stats
}

func newTestAction(attributes, options, ctx map[string]interface{}) (ExecuteAction, error) {
	return attributes["action"].(ExecuteAction), nil
}

type actionWrapper struct {
	id, name string
	enabled  bool
	action   ExecuteAction

	temporary  error
	last_error error
}

func (self *actionWrapper) Stats() map[string]interface{} {
	stats := self.action.Stats()
	if nil != self.last_error {
		stats["error"] = self.last_error.Error()
	}
	if !self.enabled {
		stats["enabled"] = self.enabled
	}
	return stats
}

func (self *actionWrapper) RunBefore() {
	self.action.RunBefore()
}

func (self *actionWrapper) RunAfter() {
	self.last_error = self.temporary
	self.action.RunAfter()
}

func (self *actionWrapper) Run(t time.Time, value interface{}) {
	if !self.enabled {
		return
	}

	defer func() {
		if e := recover(); nil != e {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("[panic]%v", e))
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}
			self.temporary = errors.New(buffer.String())
		}
	}()

	self.temporary = self.action.Run(t, value)
}
