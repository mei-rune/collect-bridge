package poller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sampling"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type httpH func(http.ResponseWriter, *http.Request)

func (self httpH) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	self(resp, req)
}

func TestJobWithSamplingFailed(t *testing.T) {
	error_message := "sdfsdfsfasf"
	called := int32(0)
	hsrv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if nil == req.Body {
			resp.WriteHeader(http.StatusNoContent)
			return
		}
		var er []*sampling.ExchangeRequest
		if e := json.NewDecoder(req.Body).Decode(&er); nil != e {
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte(e.Error()))
			return
		}
		if nil == er {
			resp.WriteHeader(http.StatusNoContent)
			return
		}
		if er[0].ManagedId == "12" && er[0].Name == "cpu" {
			atomic.AddInt32(&called, 1)
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte(error_message))
			return
		}

		resp.WriteHeader(http.StatusNoContent)
	}))

	defer hsrv.Close()

	broker, err := sampling.NewBroker("sampling_broker", hsrv.URL)
	if nil != err {
		t.Error("connect to broker failed,", err)
		return
	}

	defer broker.Close()

	ch := make(chan []string, 1000)
	tg, e := newJob(map[string]interface{}{
		"id":                "this is a test trigger id",
		"name":              "this is a test trigger",
		"type":              "metric_trigger",
		"metric":            "cpu",
		"managed_object_id": "12",
		"expression":        "@every 1ms",
		"$action": []interface{}{map[string]interface{}{
			"id":      "this is a test action id",
			"type":    "redis_command",
			"name":    "this is a test redis action",
			"command": "SET",
			"arg0":    "a",
			"arg1":    "b",
			"arg2":    "$managed_type",
			"arg3":    "$managed_id",
			"arg4":    "$metric"}}},
		map[string]interface{}{"redis_channel": forward(ch), "sampling_broker": broker})

	if nil != e {
		t.Error(e)
		return
	}

	for c := 0; c < 1000 && 0 == atomic.LoadInt32(&called); c += 1 {
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(10 * time.Millisecond)

	tg.Close(CLOSE_REASON_NORMAL)

	if 0 == called {
		t.Error("not call")
	}

	it := tg.(*metricJob).Trigger.(*intervalTrigger)
	if nil == it.last_error {
		t.Error("last_error is nil")
	} else if !strings.Contains(it.last_error.Error(), error_message) {
		t.Error("excepted error is ", error_message)
		t.Error("actual error is ", it.last_error.Error())
	}
}

func TestJobFull(t *testing.T) {
	result := []*sampling.ExchangeResponse{&sampling.ExchangeResponse{Evalue: map[string]interface{}{"name": "this is a name", "a": "b"}}}
	called := int32(0)

	hsrv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if nil == req.Body {
			resp.WriteHeader(http.StatusNoContent)
			return
		}
		var er []*sampling.ExchangeRequest
		if e := json.NewDecoder(req.Body).Decode(&er); nil != e {
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte(e.Error()))
			return
		}
		if nil == er {
			resp.WriteHeader(http.StatusNoContent)
			return
		}
		if er[0].ManagedId == "12" && er[0].Name == "cpu" {
			atomic.AddInt32(&called, 1)
			resp.WriteHeader(http.StatusAccepted)
			result[0].Id = er[0].Id
			result[0].ChannelName = er[0].ChannelName
			if e := json.NewEncoder(resp).Encode(result); nil != e {
				resp.WriteHeader(http.StatusInternalServerError)
				resp.Write([]byte(e.Error()))
			}
			return
		}

		resp.WriteHeader(http.StatusNoContent)
	}))

	defer hsrv.Close()

	broker, err := sampling.NewBroker("sampling_broker", hsrv.URL)
	if nil != err {
		t.Error("connect to broker failed,", err)
		return
	}

	defer broker.Close()

	ch := make(chan []string, 1000)
	tg, e := newJob(map[string]interface{}{
		"id":                "test_id",
		"name":              "this is a test trigger",
		"type":              "metric_trigger",
		"metric":            "cpu",
		"parent_type":       "managed_object",
		"managed_object_id": "12",
		"expression":        "@every 1ms",
		"$action": []interface{}{map[string]interface{}{
			"id":      "this is a test action id",
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
		map[string]interface{}{"redis_channel": forward(ch), "sampling_broker": broker})

	if nil != e {
		t.Error(e)
		return
	}

	for c := 0; c < 1000 && 0 == atomic.LoadInt32(&called); c += 1 {
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(10 * time.Millisecond)

	tg.Close(CLOSE_REASON_NORMAL)

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

	m := result[0].ToMap()
	m["managed_id"] = 12
	m["managed_type"] = "managed_object"
	m["metric"] = "cpu"
	m["trigger_id"] = "test_id"

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
