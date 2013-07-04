package commons

import (
	"bytes"
	"errors"
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
	Timeout       time.Duration
}

func (s *SimpleServer) Start() {
	if !atomic.CompareAndSwapInt32(&s.status, SRV_INIT, SRV_STARTING) {
		fmt.Println("start failed")
		return
	}

	if nil == s.C {
		s.C = make(chan func())
	}

	if 0 == s.Timeout {
		s.Timeout = 20 * time.Second
	}

	go s.serve()
}

var (
	notSend = errors.New("send error")
)

func (s *SimpleServer) Call(cb func()) error {
	if !s.IsRunning() {
		return DieError
	}

	s.C <- cb
	return nil
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

	next := true
	for next {
		select {
		case f := <-s.C:
			s.executeCommand(f)
		default:
			next = false
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

func (s *SimpleServer) ReturnError(cb func() error) error {
	c := make(chan error)
	defer close(c)

	s.Call(func() {
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
				c <- errors.New(buffer.String())
			}
		}()

		c <- cb()
	})

	select {
	case res := <-c:
		return res
	case <-time.After(s.Timeout):
		return TimeoutErr
	}
}

func (s *SimpleServer) ReturnString(cb func() string) string {
	c := make(chan string)
	defer close(c)

	s.Call(func() {
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
				c <- buffer.String()
			}
		}()

		c <- cb()
	})

	select {
	case res := <-c:
		return res
	case <-time.After(s.Timeout):
		return "[panic]time out"
	}
}

func (s *SimpleServer) NotReturn(cb func()) {
	c := make(chan string)
	defer close(c)

	s.Call(func() {
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
				c <- buffer.String()
			}
		}()
		cb()
		c <- "ok"
	})

	select {
	case res := <-c:
		if "ok" != res {
			panic(res)
		}
	case <-time.After(s.Timeout):
		panic("time out")
	}
}
