package poller

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"time"
)

type ExecuteAction interface {
	Run(t time.Time, value interface{}) error
}

func newAction(attributes, options, ctx map[string]interface{}) (ExecuteAction, error) {
	switch attributes["type"] {
	case "redis_command":
		return newRedisAction(attributes, options, ctx)
	case "alert":
		return newAlertAction(attributes, options, ctx)
	case "history":
		return newHistoryAction(attributes, options, ctx)
	}
	return nil, fmt.Errorf("unsupported type, - %v", attributes["type"])
}

type actionWrapper struct {
	id, name string
	action   ExecuteAction

	temporary  error
	last_error error
}

func (self *actionWrapper) Run(t time.Time, value interface{}) {
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
