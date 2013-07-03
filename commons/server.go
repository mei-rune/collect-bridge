package commons

import (
	"bytes"
	"fmt"
	"log"
	"runtime"
	"sync/atomic"
	"time"
)

const (
	SRV_INIT     = 0
	SRV_STARTING = 1
	SRV_RUNNING  = 2
	SRV_STOPPING = 3
)

type SimpleServer struct {
	status        int32
	C             chan func()
	IdledInterval time.Duration
	OnIdle        func()
}

func (s *SimpleServer) Start() {
	if !atomic.CompareAndSwapInt32(&s.status, SRV_INIT, SRV_STARTING) {
		return
	}

	go s.serve()
}

func (s *SimpleServer) IsRunning() bool {
	return SRV_RUNNING == atomic.LoadInt32(&s.status)
}

func (s *SimpleServer) Stop() {
	if atomic.CompareAndSwapInt32(&s.status, SRV_STARTING, SRV_INIT) {
		return
	}

	atomic.CompareAndSwapInt32(&s.status, SRV_RUNNING, SRV_STOPPING)
	s.sendExit()
}

func (s *SimpleServer) sendExit() {
	c := make(chan string, 1)
	defer close(c)

	cb := func() {
		defer func() {
			if e := recover(); nil != e {
				log.Printf("[panic]%v", e)
			}
		}()
		c <- "ok"
	}

	select {
	case s.C <- cb:
		select {
		case <-c:
		case <-time.After(5 * time.Second):
			return
		}
	default: // it is required, becase run() may exited {status != running}
		return
	}
}

func (s *SimpleServer) serve() {
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

	if 0 == s.IdledInterval {
		for SRV_RUNNING == atomic.LoadInt32(&s.status) {
			f := <-s.C
			s.executeCommand(f)
		}
	} else {
		for SRV_RUNNING == atomic.LoadInt32(&s.status) {
			select {
			case f := <-s.C:
				s.executeCommand(f)
			case <-time.After(s.IdledInterval):
				s.fireIdle()
			}
		}
	}
}

func (s *SimpleServer) fireIdle() {
	s.OnIdle()
}

func (s *SimpleServer) executeCommand(cb func()) {
	defer func() {
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

	cb()
}
