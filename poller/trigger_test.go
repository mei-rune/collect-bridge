package poller

import (
	"log"
	"sync/atomic"
	"testing"
	"time"
)

func TestTrigger(t *testing.T) {
	i := int32(0)
	tg, e := newTrigger(map[string]interface{}{
		"name":       "this is a test trigger",
		"expression": "@every 1ms",
		"$action":    []interface{}{}}, nil, map[string]interface{}{}, func(tm time.Time) error {
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
