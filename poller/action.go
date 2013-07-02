package poller

import (
	"fmt"
	"time"
)

type ExecuteAction interface {
	Run(t time.Time, value interface{})
}

func newAction(attributes, options, ctx map[string]interface{}) (ExecuteAction, error) {
	switch attributes["type"] {
	case "redis_command":
		return newRedisAction(attributes, options, ctx)
	case "alert":
		return newAlertAction(attributes, options, ctx)
	}
	return nil, fmt.Errorf("unsupported type, - %v", attributes["type"])
}
