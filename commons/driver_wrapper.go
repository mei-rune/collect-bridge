package commons

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DRV_MESSAGE_REQ_UNKNOW = 0
	DRV_MESSAGE_REQ_GET    = 1
	DRV_MESSAGE_REQ_PUT    = 2
	DRV_MESSAGE_REQ_DELETE = 3
	DRV_MESSAGE_REQ_CREATE = 4
	DRV_MESSAGE_REQ_CALL   = 5
	DRV_MESSAGE_REQ_EXIT   = 6

	DRV_MESSAGE_RET_PANIC  = 9000
	DRV_MESSAGE_RET_NORMAL = 9001
)

type driver_message struct {
	ch chan *driver_message

	command      int
	f            func()
	arguments    map[string]string
	returnResult Result
	returnError  RuntimeError
}

var freeChList = make(chan *driver_message, 1000)

func getCachedCh() (msg *driver_message) {
	select {
	case msg = <-freeChList:
	default:
		msg = &driver_message{ch: make(chan *driver_message),
			command: DRV_MESSAGE_REQ_UNKNOW}
	}
	return
}

func putCachedCh(msg *driver_message) {
	msg.arguments = nil
	msg.returnResult = nil
	msg.returnError = nil
	msg.f = nil
	msg.command = DRV_MESSAGE_REQ_UNKNOW
	select {
	case freeChList <- msg:
	default:
		close(msg.ch)
	}
}

type DriverWrapper struct {
	Logger
	drv         Driver
	Name        string
	initialized sync.Once
	status      int32
	ch          chan *driver_message
	timeout     time.Duration
}

func Wrap(drv Driver, timeout time.Duration) *DriverWrapper {
	return &DriverWrapper{drv: drv, timeout: timeout}
}

func (self *DriverWrapper) Start() (err error) {
	if nil == self.drv {
		return errors.New("'drv' is nil")
	}
	self.initialized.Do(func() {
		self.status = status_inactive
		self.ch = make(chan *driver_message)
		if "" == self.Name {
			self.Name = "DriverWrapper-" + strconv.Itoa(time.Now().Second())
		}

		if !self.LogInitialized() {
			self.InitLoggerWithWriter(os.Stdout, self.Name, log.LstdFlags|log.Lshortfile)
		}

		if "" == self.LogPrefix() {
			self.SetLogPrefix(self.Name)
			self.SetLogFlags(self.LogFlags() | log.Lshortfile)
		}
	})
	if !atomic.CompareAndSwapInt32(&self.status, status_inactive, status_active) {
		return
	}

	go serve_wrapper(self)

	msg := <-self.ch
	if DRV_MESSAGE_RET_PANIC == msg.command {
		err = msg.returnError
		return
	}

	return
}

func (self *DriverWrapper) IsRunning() bool {
	return status_active == atomic.LoadInt32(&self.status)
}

func (self *DriverWrapper) Stop() {
	if !atomic.CompareAndSwapInt32(&self.status, status_active, status_inactive) {
		self.INFO.Printf("It is already exited\r\n")
		return
	}

	msg := getCachedCh()
	var success bool = false
	defer func() {
		if success {
			putCachedCh(msg)
		}
	}()

	msg.command = DRV_MESSAGE_REQ_EXIT
	self.ch <- msg
	select {
	case <-msg.ch:
		success = true
	case <-time.After(5 * time.Minute):
		panic(timeout_message)
	}
	return
}

func serve_wrapper(self *DriverWrapper) {

	isStarted := false
	var exited *driver_message

	defer func() {
		atomic.CompareAndSwapInt32(&self.status, status_active, status_inactive)
		if isStarted {
			onExit(self)
		}
		if err := recover(); nil != err {
			self.INFO.Printf("%v\r\n", err)
		}
		if nil != exited {
			exited.ch <- exited
		}

	}()

	if 0 == self.timeout {
		self.timeout = 5 * time.Second
	}

	if nil == self.ch {
		self.INFO.Print("start svc failed, ch is nil!")
		return
	}

	se := onStart(self)
	if nil != se {
		self.ch <- &driver_message{command: DRV_MESSAGE_RET_NORMAL, returnError: NewRuntimeError(500, se.Error())}
		return
	}

	self.ch <- &driver_message{command: DRV_MESSAGE_RET_NORMAL}
	isStarted = true

	for {
		select {
		case msg, ok := <-self.ch:
			if !ok {
				goto exit
			}
			exited = self.safelyCall(msg)
			if nil != exited {
				goto exit
			}
		case <-time.After(self.timeout):
			onTick(self)
		}
	}
exit:
	self.INFO.Printf("channel is closed or recv an exit driver_message!\r\n")
}

func onStart(self *DriverWrapper) (err error) {

	defer func() {
		if e := recover(); nil != e {
			err = NewPanicError("call onStart failed - ", e)
		}
	}()

	startable, ok := self.drv.(Startable)
	if ok && nil != startable {
		err = startable.Start()
	}
	return
}

func onExit(self *DriverWrapper) {

	defer func() {
		if err := recover(); nil != err {
			self.INFO.Printf("on exit - %v", err)
		}
	}()

	startable, ok := self.drv.(Startable)
	if ok && nil != startable {
		startable.Stop()
	}
}

func onTick(self *DriverWrapper) {

	defer func() {
		if err := recover(); nil != err {
			self.INFO.Printf("on idle - %v", err)
		}
	}()

	idleable, ok := self.drv.(Idleable)
	if ok && nil != idleable {
		idleable.OnIdle()
	}
}

func (self *DriverWrapper) IsAlive() bool {
	return status_active == atomic.LoadInt32(&self.status)
}

func (self *DriverWrapper) safelyCall(msg *driver_message) *driver_message {
	if nil == msg {
		return nil
	}

	defer func() {
		if err := recover(); nil != err {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprint(err))
			buffer.WriteString("\r\n")
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}

			msg.command = DRV_MESSAGE_RET_PANIC
			msg.returnError = NewRuntimeError(500, buffer.String())
		}
	}()

	switch msg.command {
	case DRV_MESSAGE_REQ_EXIT:
		msg.command = DRV_MESSAGE_RET_NORMAL
		msg.returnError = nil
		return msg
	case DRV_MESSAGE_REQ_GET:
		msg.command = DRV_MESSAGE_RET_NORMAL
		msg.returnResult, msg.returnError = self.drv.Get(msg.arguments)
	case DRV_MESSAGE_REQ_PUT:
		msg.command = DRV_MESSAGE_RET_NORMAL
		msg.returnResult, msg.returnError = self.drv.Put(msg.arguments)
	case DRV_MESSAGE_REQ_DELETE:
		msg.command = DRV_MESSAGE_RET_NORMAL
		msg.returnResult, msg.returnError = self.drv.Delete(msg.arguments)
	case DRV_MESSAGE_REQ_CREATE:
		msg.command = DRV_MESSAGE_RET_NORMAL
		msg.returnResult, msg.returnError = self.drv.Create(msg.arguments)
	case DRV_MESSAGE_REQ_CALL:
		msg.command = DRV_MESSAGE_RET_NORMAL
		msg.f()
	default:
		msg.returnError = NewRuntimeError(500, fmt.Sprintf("Unsupported command - %d", msg.command))
		msg.command = DRV_MESSAGE_RET_NORMAL
	}
	msg.ch <- msg
	return nil
}

func (self *DriverWrapper) invoke(cmd int, params map[string]string, f func()) (Result, RuntimeError) {
	if !self.IsAlive() {
		return nil, DieError
	}

	msg := getCachedCh()
	var success bool = false

	defer func() {
		if success {
			putCachedCh(msg)
		} else {
			if nil != msg.ch {
				close(msg.ch)
			}
		}
	}()

	msg.command = cmd
	msg.arguments = params
	msg.f = f
	self.ch <- msg

	select {
	case resp := <-msg.ch:
		success = true
		if DRV_MESSAGE_RET_PANIC == resp.command {
			panic(resp.returnError.Error())
		}
		return resp.returnResult, resp.returnError
	case <-time.After(self.timeout):
		return nil, TimeoutErr
	}
	return nil, nil
}

func (self *DriverWrapper) Get(params map[string]string) (Result, RuntimeError) {
	return self.invoke(DRV_MESSAGE_REQ_GET, params, nil)
}

func (self *DriverWrapper) Put(params map[string]string) (Result, RuntimeError) {
	return self.invoke(DRV_MESSAGE_REQ_PUT, params, nil)
}

func (self *DriverWrapper) Create(params map[string]string) (Result, RuntimeError) {
	return self.invoke(DRV_MESSAGE_REQ_CREATE, params, nil)
}

func (self *DriverWrapper) Delete(params map[string]string) (Result, RuntimeError) {
	return self.invoke(DRV_MESSAGE_REQ_DELETE, params, nil)
}

func (self *DriverWrapper) Call(f func()) {
	self.invoke(DRV_MESSAGE_REQ_DELETE, nil, f)
}
