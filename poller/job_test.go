package poller

import (
	"commons"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sampling"
	"strings"
	"testing"
	"time"
)

type httpH func(http.ResponseWriter, *http.Request)

func (self httpH) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	self(resp, req)
}

func MockBrokerTest(t *testing.T, metric_name, managedId string, result []*sampling.ExchangeResponse, cb func(broker *sampling.SamplingBroker)) {
	is_test = true
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
		if nil == er || 0 == len(er) {
			resp.WriteHeader(http.StatusNoContent)
			return
		}
		if er[0].ManagedId == managedId && er[0].Name == metric_name {
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

	broker, err := sampling.NewBroker("sampling_broker", hsrv.URL)
	if nil != err {
		t.Error("connect to broker failed,", err)
		return
	}
	defer broker.Close()

	cb(broker)

}

func TestJobWithSamplingFailed(t *testing.T) {
	error_message := "adsfsdfsdfsdfasasdfewtt"
	result := []*sampling.ExchangeResponse{&sampling.ExchangeResponse{Eerror: &commons.ApplicationError{Ecode: 123, Emessage: error_message}}}
	MockBrokerTest(t, "cpu", "12", result, func(broker *sampling.SamplingBroker) {
		action_ch := make(chan string, 1000)
		action := &testAction{stats: map[string]interface{}{},
			run: func(tv time.Time, v interface{}) error {
				//t.Log("called action")
				action_ch <- "ok"
				return nil
			}}
		ch := make(chan []string, 1000)
		tg, e := newJob(map[string]interface{}{
			"id":                "this is a test trigger id",
			"name":              "this is a test trigger",
			"type":              "metric_trigger",
			"metric":            "cpu",
			"managed_object_id": "12",
			"expression":        "@every 1ms",
			"$action": []interface{}{map[string]interface{}{
				"id":     "12344",
				"name":   "this is a test acion name",
				"type":   "test",
				"action": action}}},
			map[string]interface{}{"redis_channel": forward(ch), "sampling_broker": broker})

		if nil != e {
			t.Error(e)
			return
		}

		defer tg.Close(CLOSE_REASON_NORMAL)

		select {
		case <-action_ch:
			break
		case <-time.After(1 * time.Second):
			t.Error("time out -", time.Now())
			return
		}

		it := tg.(*metricJob).baseJob
		if "" == it.last_error {
			t.Error("last_error is nil")
		} else if !strings.Contains(it.last_error, error_message) {
			t.Error("excepted error is ", error_message)
			t.Error("actual error is ", it.last_error)
		}
	})
}

func TestJobFull(t *testing.T) {
	result := []*sampling.ExchangeResponse{&sampling.ExchangeResponse{Evalue: map[string]interface{}{"name": "this is a name", "a": "b"}}}
	MockBrokerTest(t, "cpu", "12", result, func(broker *sampling.SamplingBroker) {
		action_ch := make(chan string, 1000)
		action := &testAction{stats: map[string]interface{}{},
			run: func(t time.Time, v interface{}) error {
				action_ch <- "ok"
				return nil
			}}
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
				"arg6":    "$metric"},
				map[string]interface{}{"id": "12344",
					"name":   "this is a test acion name",
					"type":   "test",
					"action": action}}},
			map[string]interface{}{"redis_channel": forward(ch), "sampling_broker": broker})

		if nil != e {
			t.Error(e)
			return
		}

		defer tg.Close(CLOSE_REASON_NORMAL)

		select {
		case <-action_ch:
			break
		case <-time.After(1 * time.Second):
			t.Error("time out")
			return
		}

		var res []string = nil
		select {
		case res = <-ch:
		default:
			t.Error("result is nil")
			return
		}

		m := result[0].ToMap()
		m["interval"] = 1000000
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
	})
}

func TestJobWithEnabledIsFalse(t *testing.T) {
	result := []*sampling.ExchangeResponse{}
	MockBrokerTest(t, "cpu", "12", result, func(broker *sampling.SamplingBroker) {
		action := &testAction{stats: map[string]interface{}{},
			run: func(t time.Time, v interface{}) error {
				return nil
			}}

		_, e := newJob(map[string]interface{}{
			"type":              "metric_trigger",
			"id":                "test_id",
			"name":              "this is a test trigger",
			"metric":            "cpu",
			"managed_object_id": "12",
			"expression":        "@every 1ms",
			"$action": []interface{}{map[string]interface{}{
				"id":      "12344",
				"name":    "this is a test acion name",
				"type":    "test",
				"enabled": "false",
				"action":  action}}}, map[string]interface{}{"sampling_broker": broker})

		if nil == e {
			t.Error("excepted error is '" + AllDisabled.Error() + "'")
			t.Error("actual error is nil")
		}

		if !strings.Contains(e.Error(), AllDisabled.Error()) {
			t.Error("excepted error contains '" + AllDisabled.Error() + "'")
			t.Error("actual error is", e)
			return
		}
	})
}
