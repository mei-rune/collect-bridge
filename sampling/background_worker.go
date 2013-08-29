package sampling

import (
	"bytes"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

type backgroundWorkers struct {
	c                  chan func()
	wait               sync.WaitGroup
	last_error         string
	period_interval    time.Duration
	lifecycle_interval time.Duration

	workers map[string]BackgroundWorker
}

func (self *backgroundWorkers) close() {
	close(self.c)
	self.wait.Wait()
}

func (self *backgroundWorkers) shutdown() {
	close(self.c)
}

func (self *backgroundWorkers) run() {
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

func (self *backgroundWorkers) timeout() {
	now := time.Now().Unix()
	will_delete_keys := make([]string, 0, len(self.workers))
	for k, v := range self.workers {
		if v.IsExpired(now, self.lifecycle_interval) {
			will_delete_keys = append(will_delete_keys, k)
		}
	}

	for _, k := range will_delete_keys {
		if v, ok := self.workers[k]; ok {
			v.Close()
			delete(self.workers, k)
		}
	}
}

func (self *backgroundWorkers) Add(id string, bw BackgroundWorker) {
	self.c <- func() {
		self.workers[id] = bw
	}
}

func (self *backgroundWorkers) Remove(id string) {
	self.c <- func() {
		delete(self.workers, id)
	}
}

func (self *backgroundWorkers) stats() (res interface{}) {
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
	select {
	case v := <-c:
		return v
	case <-time.After(20 * time.Second):
		return "stats is time out!"
	}
}

func (self *backgroundWorkers) stats_impl() (res interface{}) {
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
		results = append(results, bw.Stats())
	}
	return results
}
