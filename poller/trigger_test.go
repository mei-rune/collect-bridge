package poller

import (
	"testing"
	"time"
)

func TestTrigger(t *testing.T) {
	action := &testAction{stats: map[string]interface{}{},
		run: func(t time.Time, v interface{}) error {
			return nil
		}}

	tg, e := newTrigger(map[string]interface{}{
		"id":         "test_id",
		"name":       "this is a test trigger",
		"expression": "@every 1ms",
		"$action": []interface{}{map[string]interface{}{
			"id":     "12344",
			"name":   "this is a test acion name",
			"type":   "test",
			"action": action}}}, map[string]interface{}{})

	if nil != e {
		t.Error(e)
		return
	}

	trgger, e := tg.New()
	if nil != e {
		t.Error(e)
		return
	}
	defer trgger.Close()

	select {
	case <-trgger.Channel():
		t.Log("recv")
		break
	case <-time.After(1 * time.Second):
		t.Error("time out")
	}
}
