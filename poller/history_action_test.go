package poller

import (
	"commons"
	"testing"
	"time"
)

var all_history_tests = []struct {
	attributes interface{}
	value      interface{}
}{{attributes: map[string]interface{}{"a": "13"}, value: float64(13)},
	{attributes: map[string]interface{}{"a": 13}, value: 13},
	{attributes: map[string]interface{}{"a": int8(13)}, value: int8(13)},
	{attributes: map[string]interface{}{"a": int16(13)}, value: int16(13)},
	{attributes: map[string]interface{}{"a": int32(13)}, value: int32(13)},
	{attributes: map[string]interface{}{"a": int64(13)}, value: int64(13)},
	{attributes: map[string]interface{}{"a": uint(13)}, value: uint(13)},
	{attributes: map[string]interface{}{"a": uint8(13)}, value: uint8(13)},
	{attributes: map[string]interface{}{"a": uint16(13)}, value: uint16(13)},
	{attributes: map[string]interface{}{"a": uint32(13)}, value: uint32(13)},
	{attributes: map[string]interface{}{"a": uint64(13)}, value: uint64(13)},
	{attributes: map[string]interface{}{"a": float32(13)}, value: float32(13)},
	{attributes: map[string]interface{}{"a": float64(13)}, value: float64(13)}}

func TestHistorySimple(t *testing.T) {
	c1 := make(chan *data_object)
	c := make(chan *data_object, 10)

	defer close(c1)
	go func() {
		for v := range c1 {
			c <- v
			v.c <- nil
		}
	}()

	for i, test := range all_history_tests {
		action, e := newHistoryAction(map[string]interface{}{
			"id":        "123",
			"name":      "this is a test alert",
			"type":      "history",
			"attribute": "a"},
			map[string]interface{}{"metric": "cpu",
				"managed_type": "managed_object",
				"managed_id":   "1213",
				"trigger_id":   "43"},
			map[string]interface{}{"histories_channel": forward2(c1)})

		if nil != e {
			t.Error(e)
			return
		}

		history := action.(*historyAction)

		result := commons.Return(test.attributes)
		e = history.Run(time.Now(), result)
		if nil != e {
			t.Error(e)
			return
		}

		select {
		case v := <-c:
			if test.value != v.attributes["value"] {
				t.Error("test all_history_tests[", i, "] failed.")
				t.Errorf("excepted is %T,%v", test.value, test.value)
				t.Errorf("actual is %T,%v", v.attributes["value"], v.attributes["value"])
			}
		default:
			t.Error("not recv")
		}
	}
}

var all_history_failed_tests = []struct {
	attributes interface{}
	e          string
}{{attributes: map[string]interface{}{"a": "1a3"}, e: "crazy! it is not a number? - 1a3"},
	{attributes: map[string]interface{}{"a": map[int]int{}}, e: "value is not a number- map[int]int"},
	{attributes: map[string]interface{}{"a": []interface{}{1, 2}}, e: "value is not a number- []interface {}"}}

func TestHistoryTypeError(t *testing.T) {
	c1 := make(chan *data_object)
	c := make(chan *data_object, 100)

	defer close(c1)
	go func() {
		for v := range c1 {
			c <- v
			v.c <- nil
		}
	}()

	for i, test := range all_history_failed_tests {
		action, e := newHistoryAction(map[string]interface{}{
			"id":        "123",
			"name":      "this is a test alert",
			"type":      "history",
			"attribute": "a"},
			map[string]interface{}{"metric": "cpu",
				"managed_type": "managed_object",
				"managed_id":   "1213",
				"trigger_id":   "43"},
			map[string]interface{}{"histories_channel": forward2(c1)})

		if nil != e {
			t.Error(e)
			return
		}

		history := action.(*historyAction)

		result := commons.Return(test.attributes)
		e = history.Run(time.Now(), result)
		if nil == e {
			t.Error(e)
			return
		}
		if test.e != e.Error() {
			t.Error("test all_history_failed_tests[", i, "] failed.")
			t.Errorf("excepted is '%v'", test.e)
			t.Errorf("actual is '%v'", e)
		}
	}
}
