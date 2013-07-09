package poller

import (
	"commons"
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

	action, e := newAlertAction(map[string]interface{}{
		"name":             "this is a test alert",
		"max_repeated":     1,
		"expression_style": "json",
		"expression_code": map[string]interface{}{
			"attribute": "a",
			"operator":  ">",
			"value":     "12"}},
		map[string]interface{}{"managed_id": "1213"},
		map[string]interface{}{"notification_channel": forward2(c1)})

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

	e = action.Run(time.Now(), commons.Return(map[string]interface{}{"a": "12"}))
	if nil != e {
		t.Error(e)
		return
	}

	select {
	case v := <-c:
		if 0 != v.attributes["status"] {
			t.Error("status != 0, actual is %v", v.attributes["status"])
		}
		if "1213" != v.attributes["managed_id"] {
			t.Error("managed_id != '1213', actual is '%v'", v.attributes["managed_id"])
		}
	default:
		t.Error("not recv and last_status is", alert.last_status)
	}
}

func TestAlertMaxRepected(t *testing.T) {
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
		"name":             "this is a test alert",
		"max_repeated":     3,
		"expression_style": "json",
		"expression_code": map[string]interface{}{
			"attribute": "a",
			"operator":  ">",
			"value":     "12"}},
		map[string]interface{}{"managed_id": "1213"},
		map[string]interface{}{"notification_channel": forward2(c1)})

	if nil != e {
		t.Error(e)
		return
	}
	alert := action.(*alertAction)

	for i := 0; i < 2; i++ {
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
}

func TestAlertMaxRepected2(t *testing.T) {
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
		"name":             "this is a test alert",
		"max_repeated":     3,
		"expression_style": "json",
		"expression_code": map[string]interface{}{
			"attribute": "a",
			"operator":  ">",
			"value":     "12"}},
		map[string]interface{}{"managed_id": "1213"},
		map[string]interface{}{"notification_channel": forward2(c1)})

	if nil != e {
		t.Error(e)
		return
	}
	alert := action.(*alertAction)

	for i := 0; i < 2; i++ {
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

	for i := 0; i < 2; i++ {
		e = action.Run(time.Now(), commons.Return(map[string]interface{}{"a": "12"}))
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

	for i := 0; i < 2; i++ {
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
		"name":             "this is a test alert",
		"max_repeated":     1,
		"expression_style": "json",
		"expression_code": map[string]interface{}{
			"attribute": "a",
			"operator":  ">",
			"value":     "12"}},
		map[string]interface{}{"managed_id": "1213"},
		map[string]interface{}{"notification_channel": forward2(c1)})

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
		"name":             "this is a test alert",
		"max_repeated":     MAX_REPEATED,
		"expression_style": "json",
		"expression_code": map[string]interface{}{
			"attribute": "a",
			"operator":  ">",
			"value":     "12"}},
		map[string]interface{}{"managed_id": "1213"},
		map[string]interface{}{"notification_channel": forward2(c1)})

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
