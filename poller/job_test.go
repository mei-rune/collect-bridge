package poller

import (
	//"sync/atomic"
	"net"
	"net/http"
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

func TestJob(t *testing.T) {

	l, e := net.Listen("tcp", "127.0.0.1:0")
	if nil != e {
		t.Error(e)
		return
	}

	called := int32(0)
	var handler httpH = func(resp http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.URL.String(), "managed_object/12/cpu") {
			atomic.AddInt32(&called, 1)
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
		"name":        "this is a test trigger",
		"type":        "metric_trigger",
		"metric":      "cpu",
		"parent_type": "managed_object",
		"parent_id":   "12",
		"expression":  "@every 1ms",
		"$action": []interface{}{map[string]interface{}{
			"type":    "redis_command",
			"name":    "this is a test redis action",
			"command": "SET",
			"arg0":    "a",
			"arg1":    "b"}}},
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

}
