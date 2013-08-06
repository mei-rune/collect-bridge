package poller

import (
	"commons"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestAlertSimple(t *testing.T) {
	publish := make(chan []string, 1000)
	c1 := make(chan *data_object, 10)
	c := make(chan *data_object, 10)

	defer close(c1)
	go func() {
		for v := range c1 {
			c <- v
			v.c <- nil
		}
	}()

	all_tests := []struct {
		delay_times int
	}{{delay_times: 1},
		{delay_times: 2},
		{delay_times: 3},
		{delay_times: 4},
		{delay_times: 5},
		{delay_times: 6},
		{delay_times: 7},
		{delay_times: 8},
		{delay_times: 10},
		{delay_times: 20},
		{delay_times: 30},
		{delay_times: 40},
		{delay_times: 50}}

	for _, test := range all_tests {
		action, e := newAlertAction(map[string]interface{}{
			"type":             "alert",
			"id":               "123",
			"name":             "this is a test alert",
			"delay_times":      test.delay_times,
			"expression_style": "json",
			"expression_code": map[string]interface{}{
				"attribute": "a",
				"operator":  ">",
				"value":     "12"}},
			map[string]interface{}{"managed_id": "1213"},
			map[string]interface{}{"alerts_channel": forward2(c1), "redis_channel": forward(publish)})

		if nil != e {
			t.Error(e)
			return
		}

		alert := action.(*alertAction)

		sendAndNotRecv := func(a map[string]interface{}) {
			e := action.Run(time.Now(), commons.Return(a))
			if nil != e {
				t.Error(e)
				return
			}

			select {
			case <-c:
				t.Error("excepted not recv and actual is recved")
			default:
			}
		}

		sendAndRecv := func(status int, a map[string]interface{}) {
			e = action.Run(time.Now(), commons.Return(a))
			if nil != e {
				t.Error(e)
				return
			}

			select {
			case v := <-c:
				if status != v.attributes["status"] {
					t.Errorf("status != %v, actual is %v", status, v.attributes["status"])
				}
				if "1213" != v.attributes["managed_id"] {
					t.Errorf("managed_id != '1213', actual is '%v'", v.attributes["managed_id"])
				}
				select {
				case <-publish:
				case <-time.After(1 * time.Second):
					t.Error("not recv and last_status is", alert.last_status)
				}
			default:
				t.Error("not recv and last_status is", alert.last_status)
			}
		}

		//////////////////////////////////////////
		// send alert
		for i := 0; i < test.delay_times-1; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "13"})
		}
		sendAndRecv(1, map[string]interface{}{"a": "13"})

		for i := 0; i < test.delay_times; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "13"})
		}

		//////////////////////////////////////////
		// send resume
		for i := 0; i < test.delay_times-1; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "12"})
		}

		sendAndRecv(0, map[string]interface{}{"a": "12"})

		for i := 0; i < test.delay_times; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "12"})
		}
	}
}

func TestAlertSimple2(t *testing.T) {
	publish := make(chan []string, 10000)
	c1 := make(chan *data_object, 10)
	c := make(chan *data_object, 10)

	defer close(c1)
	go func() {
		for v := range c1 {
			c <- v
			v.c <- nil
		}
	}()

	all_tests := []struct {
		delay_times int
	}{{delay_times: 2},
		{delay_times: 3},
		{delay_times: 4},
		{delay_times: 5},
		{delay_times: 6},
		{delay_times: 7},
		{delay_times: 8},
		{delay_times: 10},
		{delay_times: 20},
		{delay_times: 30},
		{delay_times: 40},
		{delay_times: 50}}

	for _, test := range all_tests {
		action, e := newAlertAction(map[string]interface{}{
			"id":               "123",
			"name":             "this is a test alert",
			"delay_times":      test.delay_times,
			"expression_style": "json",
			"expression_code": map[string]interface{}{
				"attribute": "a",
				"operator":  ">",
				"value":     "12"}},
			map[string]interface{}{"managed_id": "1213"},
			map[string]interface{}{"alerts_channel": forward2(c1), "redis_channel": forward(publish)})

		if nil != e {
			t.Error(e)
			return
		}

		alert := action.(*alertAction)

		sendAndNotRecv := func(a map[string]interface{}) {
			e := action.Run(time.Now(), commons.Return(a))
			if nil != e {
				t.Error(e)
				return
			}

			select {
			case <-c:
				t.Error("excepted not recv and actual is recved")
			default:
			}
		}

		sendAndRecv := func(status int, a map[string]interface{}) {
			e = action.Run(time.Now(), commons.Return(a))
			if nil != e {
				t.Error(e)
				return
			}

			select {
			case v := <-c:
				if status != v.attributes["status"] {
					t.Errorf("status != %v, actual is %v", status, v.attributes["status"])
				}
				if "1213" != v.attributes["managed_id"] {
					t.Errorf("managed_id != '1213', actual is '%v'", v.attributes["managed_id"])
				}

				select {
				case <-publish:
				case <-time.After(1 * time.Second):
					t.Error("not recv and last_status is", alert.last_status)
				}

			default:
				t.Error("not recv and last_status is", alert.last_status)
			}
		}

		//////////////////////////////////////////
		// send resume
		for i := 0; i < 2*test.delay_times; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "12"})
		}

		//////////////////////////////////////////
		// send alert
		for i := 0; i < test.delay_times-1; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "13"})
		}

		sendAndNotRecv(map[string]interface{}{"a": "12"})

		for i := 0; i < test.delay_times-1; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "13"})
		}

		sendAndRecv(1, map[string]interface{}{"a": "13"})

		for i := 0; i < test.delay_times; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "13"})
		}

		//////////////////////////////////////////
		// send resume
		for i := 0; i < test.delay_times-1; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "12"})
		}

		sendAndNotRecv(map[string]interface{}{"a": "13"})

		for i := 0; i < test.delay_times-1; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "12"})
		}

		sendAndRecv(0, map[string]interface{}{"a": "12"})

		for i := 0; i < test.delay_times; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "12"})
		}
	}
}

func TestAlertWithSendFailed(t *testing.T) {
	publish := make(chan []string, 10000)
	c1 := make(chan *data_object, 10)
	c := make(chan *data_object, 10)

	returnFailed := int32(0)
	reeturnError := errors.New("TestAlertWithSendFailed")

	defer close(c1)
	go func() {
		for v := range c1 {
			if int32(0) == atomic.LoadInt32(&returnFailed) {
				c <- v
				v.c <- nil
			} else {
				c <- v
				v.c <- reeturnError
			}
		}
	}()

	all_tests := []struct {
		delay_times int
	}{{delay_times: 1},
		{delay_times: 2},
		{delay_times: 3},
		{delay_times: 4},
		{delay_times: 5},
		{delay_times: 6},
		{delay_times: 7},
		{delay_times: 8},
		{delay_times: 10},
		{delay_times: 20},
		{delay_times: 30},
		{delay_times: 40},
		{delay_times: 50}}

	for _, test := range all_tests {

		action, e := newAlertAction(map[string]interface{}{
			"id":               "123",
			"name":             "this is a test alert",
			"delay_times":      test.delay_times,
			"expression_style": "json",
			"expression_code": map[string]interface{}{
				"attribute": "a",
				"operator":  ">",
				"value":     "12"}},
			map[string]interface{}{"managed_id": "1213"},
			map[string]interface{}{"alerts_channel": forward2(c1), "redis_channel": forward(publish)})

		if nil != e {
			t.Error(e)
			return
		}

		alert := action.(*alertAction)

		sendAndNotRecv := func(a map[string]interface{}) {
			e := action.Run(time.Now(), commons.Return(a))
			if nil != e {
				t.Error(e)
				return
			}

			select {
			case <-c:
				t.Error("excepted not recv and actual is recved")
			default:
			}
		}

		sendAndRecv := func(status int, a map[string]interface{}) {
			e := action.Run(time.Now(), commons.Return(a))
			if nil != e {
				if int32(0) == atomic.LoadInt32(&returnFailed) {
					t.Error(e)
					return
				}

				if !strings.Contains(e.Error(), reeturnError.Error()) {
					t.Error("!strings.Contains(e.Error(), reeturnError.Error()),", e)
				}
			}

			select {
			case v := <-c:
				if status != v.attributes["status"] {
					t.Errorf("status != %v, actual is %v", status, v.attributes["status"])
				}
				if "1213" != v.attributes["managed_id"] {
					t.Errorf("managed_id != '1213', actual is '%v'", v.attributes["managed_id"])
				}

				select {
				case <-publish:
					if int32(0) != atomic.LoadInt32(&returnFailed) {
						t.Error("excepted is not recv, but actual is recv")
					}
				default:
					if int32(0) == atomic.LoadInt32(&returnFailed) {
						t.Error("excepted is recv, but actual is not recv")
					}
				}

			default:
				t.Error("not recv and last_status is", alert.last_status)
			}
		}

		//////////////////////////////////////////
		// send alert
		for i := 0; i < test.delay_times-1; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "13"})
		}
		atomic.StoreInt32(&returnFailed, 1)
		for i := 0; i < test.delay_times; i++ {
			sendAndRecv(1, map[string]interface{}{"a": "13"})
		}
		atomic.StoreInt32(&returnFailed, 0)
		sendAndRecv(1, map[string]interface{}{"a": "13"})

		for i := 0; i < test.delay_times; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "13"})
		}
		//////////////////////////////////////////
		// send resume
		for i := 0; i < test.delay_times-1; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "12"})
		}
		atomic.StoreInt32(&returnFailed, 1)
		for i := 0; i < test.delay_times; i++ {
			sendAndRecv(0, map[string]interface{}{"a": "12"})
		}
		atomic.StoreInt32(&returnFailed, 0)
		sendAndRecv(0, map[string]interface{}{"a": "12"})

		for i := 0; i < test.delay_times; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "12"})
		}
	}
}

func TestAlertRepectedOverflow(t *testing.T) {
	publish := make(chan []string, 10000)
	c1 := make(chan *data_object, 10)
	c := make(chan *data_object, 10)

	defer close(c1)
	go func() {
		for v := range c1 {
			c <- v
			v.c <- nil
		}
	}()

	action, e := newAlertAction(map[string]interface{}{
		"id":               "1243",
		"name":             "this is a test alert",
		"delay_times":      1,
		"expression_style": "json",
		"expression_code": map[string]interface{}{
			"attribute": "a",
			"operator":  ">",
			"value":     "12"}},
		map[string]interface{}{"managed_id": "1213"},
		map[string]interface{}{"alerts_channel": forward2(c1), "redis_channel": forward(publish)})

	if nil != e {
		t.Error(e)
		return
	}
	alert := action.(*alertAction)

	e = action.Run(time.Now(), commons.Return(map[string]interface{}{"a": "13"}))
	if nil != e {
		t.Error(e)
		return
	}

	select {
	case v := <-c:
		if 1 != v.attributes["status"] {
			t.Error("status != 1, actual is %v", v.attributes["status"])
		}
		if "1213" != v.attributes["managed_id"] {
			t.Error("managed_id != '1213', actual is '%v'", v.attributes["managed_id"])
		}
	default:
		t.Error("not recv and last_status is", alert.last_status)
	}

	for i := int64(0); i < int64(9999999); i++ {
		e = action.Run(time.Now(), commons.Return(map[string]interface{}{"a": "13"}))
		if nil != e {
			t.Error(e)
			return
		}

		select {
		case <-c:
			t.Error("excepted not recv and actual is recved")
		default:
		}
	}
}

func TestAlertRepectedOverflow2(t *testing.T) {
	publish := make(chan []string, 10000)
	c1 := make(chan *data_object, 10)
	c := make(chan *data_object, 10)

	defer close(c1)
	go func() {
		for v := range c1 {
			c <- v
			v.c <- nil
		}
	}()

	action, e := newAlertAction(map[string]interface{}{
		"id":               "123",
		"name":             "this is a test alert",
		"delay_times":      MAX_REPEATED,
		"expression_style": "json",
		"expression_code": map[string]interface{}{
			"attribute": "a",
			"operator":  ">",
			"value":     "12"}},
		map[string]interface{}{"managed_id": "1213"},
		map[string]interface{}{"alerts_channel": forward2(c1), "redis_channel": forward(publish)})

	if nil != e {
		t.Error(e)
		return
	}
	alert := action.(*alertAction)

	count := 0
	for i := int64(0); i < int64(MAX_REPEATED*2); i++ {
		e = alert.Run(time.Now(), commons.Return(map[string]interface{}{"a": "13"}))
		if nil != e {
			t.Error(e)
			return
		}

		select {
		case <-c:
			count++
		default:
		}
	}

	if count != 1 {
		t.Error("excepted recv count is 1 and actual is ", count)
	}
}

func TestAlertLoadLastStatus(t *testing.T) {
	publish := make(chan []string, 1000)
	c1 := make(chan *data_object, 10)
	c := make(chan *data_object, 10)

	defer close(c1)
	go func() {
		for v := range c1 {
			c <- v
			v.c <- nil
		}
	}()

	all_tests := []struct {
		delay_times int
	}{{delay_times: 1}}

	for _, test := range all_tests {
		action, e := newAlertAction(map[string]interface{}{
			"type":             "alert",
			"id":               "123",
			"name":             "this is a test alert",
			"delay_times":      test.delay_times,
			"last_status":      1,
			"expression_style": "json",
			"expression_code": map[string]interface{}{
				"attribute": "a",
				"operator":  ">",
				"value":     "12"}},
			map[string]interface{}{"managed_id": "1213"},
			map[string]interface{}{"alerts_channel": forward2(c1), "redis_channel": forward(publish)})

		if nil != e {
			t.Error(e)
			return
		}

		alert := action.(*alertAction)

		sendAndNotRecv := func(a map[string]interface{}) {
			e := action.Run(time.Now(), commons.Return(a))
			if nil != e {
				t.Error(e)
				return
			}

			select {
			case <-c:
				t.Error("excepted not recv and actual is recved")
			default:
			}
		}

		sendAndRecv := func(status int, a map[string]interface{}) {
			e = action.Run(time.Now(), commons.Return(a))
			if nil != e {
				t.Error(e)
				return
			}

			select {
			case v := <-c:
				if status != v.attributes["status"] {
					t.Errorf("status != %v, actual is %v", status, v.attributes["status"])
				}
				if "1213" != v.attributes["managed_id"] {
					t.Errorf("managed_id != '1213', actual is '%v'", v.attributes["managed_id"])
				}
				select {
				case <-publish:
				case <-time.After(1 * time.Second):
					t.Error("not recv and last_status is", alert.last_status)
				}
			default:
				t.Error("not recv and last_status is", alert.last_status)
			}
		}

		//////////////////////////////////////////
		// send alert
		for i := 0; i < 2*test.delay_times; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "13"})
		}

		//////////////////////////////////////////
		// send resume
		for i := 0; i < test.delay_times-1; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "12"})
		}

		sendAndRecv(0, map[string]interface{}{"a": "12"})

		for i := 0; i < test.delay_times; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "12"})
		}
	}
}

func TestAlertEventId(t *testing.T) {
	publish := make(chan []string, 1000)
	c1 := make(chan *data_object, 10)
	c := make(chan *data_object, 10)

	defer close(c1)
	go func() {
		for v := range c1 {
			c <- v
			v.c <- nil
		}
	}()

	all_tests := []struct {
		delay_times int
	}{{delay_times: 1},
		{delay_times: 2},
		{delay_times: 3},
		{delay_times: 4},
		{delay_times: 5},
		{delay_times: 6},
		{delay_times: 7},
		{delay_times: 8},
		{delay_times: 10},
		{delay_times: 20},
		{delay_times: 30},
		{delay_times: 40},
		{delay_times: 50}}

	for _, test := range all_tests {

		action, e := newAlertAction(map[string]interface{}{
			"type":             "alert",
			"id":               "123",
			"name":             "this is a test alert",
			"delay_times":      test.delay_times,
			"expression_style": "json",
			"expression_code": map[string]interface{}{
				"attribute": "a",
				"operator":  ">",
				"value":     "12"}},
			map[string]interface{}{"managed_id": "1213"},
			map[string]interface{}{"alerts_channel": forward2(c1), "redis_channel": forward(publish)})

		if nil != e {
			t.Error(e)
			return
		}

		alert := action.(*alertAction)

		sendAndNotRecv := func(a map[string]interface{}) {
			e := action.Run(time.Now(), commons.Return(a))
			if nil != e {
				t.Error(e)
				return
			}

			select {
			case <-c:
				t.Error("excepted not recv and actual is recved")
			default:
			}
		}

		sendAndRecv := func(status, previous_status int, event_id string, seq_id int, a map[string]interface{}) string {
			e = action.Run(time.Now(), commons.Return(a))
			if nil != e {
				t.Error(e)
				return ""
			}

			select {
			case v := <-c:
				if previous_status != v.attributes["previous_status"] {
					t.Errorf("previous_status != %v, actual is %v", previous_status, v.attributes["previous_status"])
				}
				if status != v.attributes["status"] {
					t.Errorf("status != %v, actual is %v", status, v.attributes["status"])
				}
				if "1213" != v.attributes["managed_id"] {
					t.Errorf("managed_id != '1213', actual is '%v'", v.attributes["managed_id"])
				}

				if 0 != len(event_id) {
					if event_id != v.attributes["event_id"] {
						t.Errorf("event_id != %v, actual is %v", event_id, v.attributes["event_id"])
					}
				} else if 0 == len(fmt.Sprint(v.attributes["event_id"])) {
					t.Error("event_id is emtpy.")
				}

				if seq_id != v.attributes["sequence_id"] {
					t.Errorf("sequence_id != %v, actual is %v", seq_id, v.attributes["sequence_id"])
				}

				select {
				case <-publish:
				case <-time.After(1 * time.Second):
					t.Error("not recv and last_status is", alert.last_status)
				}
				return fmt.Sprint(v.attributes["event_id"])
			default:
				t.Error("not recv and last_status is", alert.last_status)
				return ""
			}
		}

		//////////////////////////////////////////
		// send alert
		for i := 0; i < test.delay_times-1; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "13"})
		}
		event_id := sendAndRecv(1, 0, "", 1, map[string]interface{}{"a": "13"})

		for i := 0; i < test.delay_times; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "13"})
		}

		//////////////////////////////////////////
		// send resume
		for i := 0; i < test.delay_times-1; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "12"})
		}

		event_id = sendAndRecv(0, 1, event_id, 2, map[string]interface{}{"a": "12"})

		for i := 0; i < test.delay_times; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "12"})
		}

		//////////////////////////////////////////
		// send alert
		for i := 0; i < test.delay_times-1; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "13"})
		}
		event_id2 := sendAndRecv(1, 0, "", 1, map[string]interface{}{"a": "13"})
		if event_id2 == event_id {
			t.Errorf("event_id2 == event_id, event_id is %v, event_id2 is %v", event_id, event_id2)
		}

		for i := 0; i < test.delay_times; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "13"})
		}

		//////////////////////////////////////////
		// send resume
		for i := 0; i < test.delay_times-1; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "12"})
		}

		event_id2 = sendAndRecv(0, 1, event_id2, 2, map[string]interface{}{"a": "12"})

		for i := 0; i < test.delay_times; i++ {
			sendAndNotRecv(map[string]interface{}{"a": "12"})
		}
	}
}
