package poller

import (
	"log"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestTrigger(t *testing.T) {
	action := &testAction{stats: map[string]interface{}{},
		run: func(t time.Time, v interface{}) error {
			return nil
		}}

	i := int32(0)
	tg, e := newTrigger(map[string]interface{}{
		"id":         "test_id",
		"name":       "this is a test trigger",
		"expression": "@every 1ms",
		"$action": []interface{}{map[string]interface{}{
			"id":     "12344",
			"name":   "this is a test acion name",
			"type":   "test",
			"action": action}}}, nil, map[string]interface{}{}, func(tm time.Time) error {
		atomic.AddInt32(&i, 1)
		t.Log("timeout ", i)
		return nil
	})

	if nil != e {
		t.Error(e)
		return
	}

	tg.Logger.InitLoggerWithCallback(func(bs []byte) {
		t.Log(string(bs))
	}, "test trigger", log.LstdFlags)

	if !tg.Logger.DEBUG.IsEnabled() {
		tg.Logger.DEBUG.Switch()
	}

	e = tg.Start()
	if nil != e {
		t.Error(e)
		return
	}
	defer tg.Stop()

	for c := 0; c < 1000 && 0 == atomic.LoadInt32(&i); c += 1 {
		time.Sleep(10 * time.Microsecond)
	}

	tg.Stop()

	if i <= 0 {
		t.Error("it is not timeout")
	}
}

func TestTriggerWithEnabledIsFalse(t *testing.T) {
	action := &testAction{stats: map[string]interface{}{},
		run: func(t time.Time, v interface{}) error {
			return nil
		}}

	i := int32(0)
	_, e := newTrigger(map[string]interface{}{
		"id":         "test_id",
		"name":       "this is a test trigger",
		"expression": "@every 1ms",
		"$action": []interface{}{map[string]interface{}{
			"id":      "12344",
			"name":    "this is a test acion name",
			"type":    "test",
			"enabled": "false",
			"action":  action}}}, nil, map[string]interface{}{}, func(tm time.Time) error {
		atomic.AddInt32(&i, 1)
		t.Log("timeout ", i)
		return nil
	})

	if nil == e {
		t.Error("excepted error is 'all actions is disable.'")
		t.Error("actual error is nil")
	}

	if !strings.Contains(e.Error(), "all actions is disable.") {
		t.Error("excepted error contains 'all actions is disable.'")
		t.Error("actual error is", e)
		return
	}
}
