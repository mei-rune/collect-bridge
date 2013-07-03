package poller

import (
	//"sync/atomic"
	"commons"
	"encoding/json"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	//"time"
)

type httpH func(http.ResponseWriter, *http.Request)

func (self httpH) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	self(resp, req)
}

func TestJobWithSamplingFailed(t *testing.T) {

	l, e := net.Listen("tcp", "127.0.0.1:0")
	if nil != e {
		t.Error(e)
		return
	}

	error_message := "sdfsdfsfasf"
	called := int32(0)
	var handler httpH = func(resp http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.URL.String(), "managed_object/12/cpu") {
			atomic.AddInt32(&called, 1)
		}
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(error_message))
	}

	c := make(chan string)
	go func() {
		http.Serve(l, handler)
		c <- "ok"
	}()

	stop := func() {
		if nil != l {
			l.Close()
			l = nil

			<-c
			close(c)
		}
	}

	defer stop()

	ch := make(chan []string, 1)
	tg, e := newJob(map[string]interface{}{
		"name":              "this is a test trigger",
		"type":              "metric_trigger",
		"metric":            "cpu",
		"managed_object_id": "12",
		"expression":        "@every 1ms",
		"$action": []interface{}{map[string]interface{}{
			"type":    "redis_command",
			"name":    "this is a test redis action",
			"command": "SET",
			"arg0":    "a",
			"arg1":    "b",
			"arg2":    "$managed_type",
			"arg3":    "$managed_id",
			"arg4":    "$metric"}}},
		map[string]interface{}{"redis_channel": forward(ch), "metrics.url": "http://" + l.Addr().String()})

	if nil != e {
		t.Error(e)
		return
	}

	e = tg.Start()
	if nil != e {
		t.Error(e)
		return
	}
	defer tg.Stop()

	for c := 0; c < 1000 && 0 == atomic.LoadInt32(&called); c += 1 {
		time.Sleep(10 * time.Microsecond)
	}

	tg.Stop()
	stop()

	if 0 == called {
		t.Error("not call")
	}

	if nil == tg.(*metricJob).last_error {
		t.Error("last_error is nil")
	} else if !strings.Contains(tg.(*metricJob).last_error.Error(), error_message) {
		t.Error("excepted error is ", error_message)
		t.Error("actual error is ", tg.(*metricJob).last_error.Error())
	}
}

func TestJobFull(t *testing.T) {

	l, e := net.Listen("tcp", "127.0.0.1:0")
	if nil != e {
		t.Error(e)
		return
	}

	result := commons.Return(map[string]interface{}{"name": "this is a name", "a": "b"})
	called := int32(0)
	var handler httpH = func(resp http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.URL.String(), "managed_object/12/cpu") {
			atomic.AddInt32(&called, 1)

			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(result.ToJson()))
		}
	}

	c := make(chan string)
	go func() {
		http.Serve(l, handler)
		c <- "ok"
	}()

	stop := func() {
		if nil != l {
			l.Close()
			l = nil

			<-c
			close(c)
		}
	}

	defer stop()

	ch := make(chan []string, 1)
	tg, e := newJob(map[string]interface{}{
		"name":              "this is a test trigger",
		"type":              "metric_trigger",
		"metric":            "cpu",
		"parent_type":       "managed_object",
		"managed_object_id": "12",
		"expression":        "@every 1ms",
		"$action": []interface{}{map[string]interface{}{
			"type":    "redis_command",
			"name":    "this is a test redis action",
			"command": "SET",
			"arg0":    "sdfs",
			"arg1":    "$$",
			"arg2":    "arg2",
			"arg3":    "$name",
			"arg4":    "$managed_type",
			"arg5":    "$managed_id",
			"arg6":    "$metric"}}},
		map[string]interface{}{"redis_channel": forward(ch), "metrics.url": "http://" + l.Addr().String()})

	if nil != e {
		t.Error(e)
		return
	}

	e = tg.Start()
	if nil != e {
		t.Error(e)
		return
	}
	defer tg.Stop()

	for c := 0; c < 1000 && 0 == atomic.LoadInt32(&called); c += 1 {
		time.Sleep(10 * time.Microsecond)
	}

	tg.Stop()
	stop()

	if 0 == called {
		t.Error("not call")
	}

	var res []string = nil
	select {
	case res = <-ch:
	default:
		t.Error("result is nil")
		return
	}

	m := result.ToMap()
	m["managed_id"] = "12"
	m["managed_type"] = "managed_object"
	m["metric"] = "cpu"

	js, e := json.Marshal(m)
	if nil != e {
		t.Error(e)
		return
	}

	excepted := []string{"SET", "sdfs", string(js), "arg2", "this is a name", "managed_object", "12", "cpu"}
	if !reflect.DeepEqual(res, excepted) {
		t.Error("excepted is ", excepted)
		t.Error("actual is ", res)
	}
}
