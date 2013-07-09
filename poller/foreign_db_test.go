package poller

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestForeignDB(t *testing.T) {
	var js1 string
	var js2 string
	count := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bs, _ := ioutil.ReadAll(r.Body)
		switch count {
		case 0:
			js1 = string(bs)
			break
		case 1:
			js2 = string(bs)
			break
		}
		count++
		time.Sleep(1 * time.Second)
	}))
	defer ts.Close()

	c := make(chan error, 10)

	fdb, e := newForeignDb("test", ts.URL)
	if nil != e {
		t.Error(e)
		return
	}
	fdb.c <- &data_object{c: nil, attributes: map[string]interface{}{"a": 23}}

	time.Sleep(2 * time.Second)
	for i := 0; i < 10; i++ {
		fdb.c <- &data_object{c: c, attributes: map[string]interface{}{"a": 23}}
	}
	for i := 0; i < 10; i++ {
		if 10 == len(c) {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if 2 != count {
		t.Error("2 != count , it is ", count)
	}

	if 10 != len(c) {
		t.Error("10 != len(c), it is ", len(c))
	} else {
		for i := 0; i < 10; i++ {
			if e := <-c; nil != e {
				t.Errorf("%T, %v", e, e)
			}
		}
	}

	for idx, js := range []string{js1, js2} {
		if 0 == len(js) {
			t.Error("js[", idx, "] is nil")
			continue
		}
		var v []map[string]interface{}
		e := json.Unmarshal([]byte(js), &v)
		if nil != e {
			t.Error(js)
			t.Error("it is not a json, ", e)
		} else {
			t.Log(js)
			t.Log("it is not a ok json")
		}
	}
}
