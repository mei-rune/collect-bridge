package sampling

import (
	"bytes"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

type simpleWorkers struct {
	l       sync.Mutex
	workers map[string]*worker
}

func (self *simpleWorkers) Add(id string, bw BackgroundWorker) {
	self.l.Lock()
	defer self.l.Unlock()
	self.workers[id] = &worker{bw, ""}
}

func (self *simpleWorkers) Remove(id string) {
	self.l.Lock()
	defer self.l.Unlock()
	delete(self.workers, id)
}

type wrappedWorkers struct {
	backend BackgroundWorkers
}

func (self *wrappedWorkers) Add(id string, bw BackgroundWorker) {
	self.backend.Add(id, bw)
}

func (self *wrappedWorkers) Remove(id string) {
	self.backend.Remove(id)
}

type backgroundWorkers struct {
	c               chan func()
	wait            sync.WaitGroup
	last_error      string
	period_interval time.Duration
	workers         map[string]*worker
}

type worker struct {
	BackgroundWorker
	last_error string
}

func (self *backgroundWorkers) close() {
	for _, w := range self.workers {
		w.Close()
	}

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
	for _, v := range self.workers {
		self.onTick(v)
	}
}

func (self *backgroundWorkers) onTick(w *worker) {
	w.last_error = ""
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
			w.last_error = buffer.String()
		}
	}()

	w.OnTick()
}

func (self *backgroundWorkers) Add(id string, bw BackgroundWorker) {
	self.c <- func() {
		self.workers[id] = &worker{bw, ""}
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
	for k, w := range self.workers {
		res := w.Stats()
		if nil == res {
			res = map[string]interface{}{"id": k}
		}
		if 0 != len(w.last_error) {
			res["error_on_tick"] = w.last_error
		}
		results = append(results, res)
	}
	return results
}
