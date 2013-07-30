package sampling

import (
	"bytes"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

type worker interface {
	stats() interface{}
	close()
	isTimeout(now int64, default_interval time.Duration) bool
}

type backgroundWorker struct {
	name             string
	c                chan func()
	wait             sync.WaitGroup
	last_error       string
	closers          []func()
	period_interval  time.Duration
	timeout_interval time.Duration

	workers map[string]worker
}

func (self *backgroundWorker) Close() error {
	close(self.c)
	self.wait.Wait()
	return nil
}

func (self *backgroundWorker) shutdown() {
	close(self.c)
}

func (self *backgroundWorker) run() {
	defer func() {
		if e := recover(); nil != e {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("[exited][panic]%v", e))
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}
			msg := buffer.String()
			self.last_error = msg
			log.Println(msg)
		}

		self.wait.Done()
	}()

	if 1000 > self.period_interval {
		self.period_interval = 5 * time.Second
	}

	ticker := time.NewTicker(self.period_interval)
	defer ticker.Stop()

	is_running := true
	for is_running {
		select {
		case <-ticker.C:
			self.timeout()
		case cb, ok := <-self.c:
			if !ok {
				is_running = false
			}
			cb()
		}
	}
}

func (self *backgroundWorker) timeout() {
	now := time.Now().Unix()
	will_delete_keys := make([]string, 0, len(self.workers))
	for k, v := range self.workers {
		if v.isTimeout(now, self.timeout_interval) {
			will_delete_keys = append(will_delete_keys, k)
		}
	}

	for _, k := range will_delete_keys {
		if v, ok := self.workers[k]; ok {
			v.close()
			delete(self.workers, k)
		}
	}
}

func (self *backgroundWorker) add(id string, bw worker) {
	self.c <- func() {
		self.workers[id] = bw
	}
}

func (self *backgroundWorker) remove(id string, bw worker) {
	self.c <- func() {
		delete(self.workers, id)
	}
}

func (self *backgroundWorker) stats() (res interface{}) {
	defer func() {
		if o := recover(); nil != o {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("[panic]%v", o))
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}
			res = buffer.String()
		}
	}()

	c := make(chan interface{})
	self.c <- func() {
		c <- self.stats_impl()
	}
	return <-c
}

func (self *backgroundWorker) stats_impl() (res interface{}) {
	defer func() {
		if o := recover(); nil != o {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("[panic]%v", o))
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}
			res = buffer.String()
		}
	}()

	var results []interface{}
	for _, bw := range self.workers {
		results = append(results, bw.stats())
	}
	return results
}
