package poller

import (
	"commons"
	"fmt"
	"time"

	"errors"
)

type ExecuteAction interface {
	Run(t time.Time, value interface{})
}

func NewAction(attributes, ctx map[string]interface{}) (ExecuteAction, error) {
	switch attributes["type"] {
	case "redis_action":
		return NewRedisAction(attributes, ctx)
	case "alert_action":
		return NewAlertAction(attributes, ctx)
	}
	return nil, fmt.Errorf("unsupported type - %v", attributes["type"])
}
