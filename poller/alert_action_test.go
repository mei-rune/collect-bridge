package poller

import (
	"commons"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestAlertSimple(t *testing.T) {
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
			map[string]interface{}{"alerts_channel": forward2(c1)})

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
			map[string]interface{}{"alerts_channel": forward2(c1)})

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
			default:
				t.Error("not recv and last_status is", alert.last_status)
			}
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
			map[string]interface{}{"alerts_channel": forward2(c1)})

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
		map[string]interface{}{"alerts_channel": forward2(c1)})

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
		map[string]interface{}{"alerts_channel": forward2(c1)})

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
