package poller

import (
	"commons"
	"reflect"
	"testing"
	"time"
)

func TestHistorySimple(t *testing.T) {
	c := make(chan *data_object, 10)

	action, e := newHistoryAction(map[string]interface{}{
		"id":        "123",
		"name":      "this is a test alert",
		"type":      "history",
		"attribute": "a"},
		map[string]interface{}{"metric": "cpu",
			"managed_type": "managed_object",
			"managed_id":   "1213",
			"trigger_id":   "43"},
		map[string]interface{}{"histories_channel": forward2(c)})

	if nil != e {
		t.Error(e)
		return
	}

	history := action.(*historyAction)

	result := commons.Return(map[string]interface{}{"a": "13"})
	e = history.Run(time.Now(), result)
	if nil != e {
		t.Error(e)
		return
	}

	excepted := map[string]interface{}{
		"action_id":    "123",
		"sampled_at":   result.CreatedAt(),
		"metric":       "cpu",
		"managed_type": "managed_object",
		"managed_id":   "1213",
		"trigger_id":   "43",
		"value":        "13"}

	select {
	case v := <-c:
		if !reflect.DeepEqual(excepted, v.attributes) {
			t.Error("excepted is", excepted)
			t.Error("actual is", v.attributes)
		}
	default:
		t.Error("not recv")
	}
}
