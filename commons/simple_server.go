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

var (
	notSend = errors.New("send error")
)

const (
	SRV_INIT     = 0
	SRV_STARTING = 1
	SRV_RUNNING  = 2
	SRV_STOPPING = 3
)

func ToStatusString(status int) string {
	switch status {
	case SRV_INIT:
		return "init"
	case SRV_STARTING:
		return "starting"
	case SRV_RUNNING:
		return "running"
	case SRV_STOPPING:
		return "stopping"
	default:
		return fmt.Sprintf("unsupport status - %v", status)
	}
}

type SimpleServer struct {
	status    int32
	C         chan func()
	Interval  time.Duration
	OnTimeout func()
	Timeout   time.Duration

	OnStart func() error
	OnStop  func()
}

func (s *SimpleServer) Start() error {
	if !atomic.CompareAndSwapInt32(&s.status, SRV_INIT, SRV_STARTING) {
		return AlreadyStartedError
	}

	if nil == s.C {
		s.C = make(chan func())
	}

	if 0 == s.Timeout {
		s.Timeout = 20 * time.Second
	}

	if nil != s.OnStart {
		e := s.OnStart()
		if nil != e {
			atomic.StoreInt32(&s.status, SRV_INIT)
			return e
		}
	}

	go s.serve()
	return nil
}

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

func (s *SimpleServer) StatusString() string {
	return ToStatusString(int(atomic.LoadInt32(&s.status)))
}

func (s *SimpleServer) Stop() {
	if atomic.CompareAndSwapInt32(&s.status, SRV_STARTING, SRV_INIT) {
		goto end
	}

	atomic.CompareAndSwapInt32(&s.status, SRV_RUNNING, SRV_STOPPING)
	s.sendExit()

end:
	if nil != s.OnStop {
		s.OnStop()
	}
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

	if 0 == s.Interval {
		for SRV_RUNNING == atomic.LoadInt32(&s.status) {
			f := <-s.C
			s.executeCommand(f)
		}
	} else {
		ticker := time.NewTicker(s.Interval)
		defer ticker.Stop()

		for SRV_RUNNING == atomic.LoadInt32(&s.status) {
			select {
			case f := <-s.C:
				s.executeCommand(f)
			case <-ticker.C:
				s.fireTick()
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

func (s *SimpleServer) fireTick() {
	if nil != s.OnTimeout {
		s.OnTimeout()
	}
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
