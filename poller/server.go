package poller

import (
	"bytes"
	"fmt"
	"log"
	"runtime"
	"sync/atomic"
)

type server struct {
	jobs   map[string]Job
	status int32
	c      chan *srv_request
}

type srv_request struct {
	c chan *srv_request
}

func (s *server) register(job Job) {
	s.jobs[job.Id()] = job
}

func (s *server) Start() {
	if !atomic.CompareAndSwapInt32(&s.status, SRV_INIT, SRV_STARTING) {
		return
	}

	go s.serve()
}

func (s *server) Stop() {
	if atomic.CompareAndSwapInt32(&s.status, SRV_STARTING, SRV_STOPPING) {
		return
	}

	atomic.CompareAndSwapInt32(&s.status, SRV_RUNNING, SRV_STOPPING)
}

func (s *server) serve() {
	if !atomic.CompareAndSwapInt32(&s.status, SRV_STARTING, SRV_RUNNING) {
		return
	}

	defer func() {
		atomic.StoreInt32(&s.status, SRV_INIT)

		if e := recover(); nil != e {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("[panic]%v", e))
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}
			log.Println(buffer.String())
		}
	}()

}
