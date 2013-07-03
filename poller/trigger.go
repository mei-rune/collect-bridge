package poller

import (
	"bytes"
	"commons"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

const (
	SRV_INIT     = 0
	SRV_STARTING = 1
	SRV_RUNNING  = 2
	SRV_STOPPING = 3
)

type triggerFunc func(t time.Time)

type trigger struct {
	commons.Logger
	id        string
	name      string
	actions   []ExecuteAction
	callback  triggerFunc
	isRunning int32

	commands    map[string]func(t *trigger) string
	expression  string
	attachment  string
	description string

	start func(t *trigger) error
	stop  func(t *trigger)
}

func (self *trigger) Id() string {
	return self.id
}

func (self *trigger) Name() string {
	return self.name
}

func (self *trigger) Start() error {
	if !atomic.CompareAndSwapInt32(&self.isRunning, SRV_INIT, SRV_STARTING) {
		return nil
	}

	e := self.start(self)
	if nil != e {
		atomic.StoreInt32(&self.isRunning, SRV_INIT)
	}
	return e
}

func (self *trigger) Stop() {
	if atomic.CompareAndSwapInt32(&self.isRunning, SRV_STARTING, SRV_STOPPING) {
		self.DEBUG.Print("it is starting")
		return
	}

	if !atomic.CompareAndSwapInt32(&self.isRunning, SRV_RUNNING, SRV_STOPPING) {
		self.DEBUG.Print("it is not running")
		return
	}

	self.stop(self)
}

func (self *trigger) callActions(t time.Time, res interface{}) {
	if nil == res {
		self.WARN.Print("result of '" + self.name + "' is nil")
		return
	}
	if nil == self.actions || 0 == len(self.actions) {
		self.WARN.Print("actions of '" + self.name + "' is empty")
		return
	}
	for _, action := range self.actions {
		action.Run(t, res)
	}
}

const every = "@every "

var (
	ExpressionSyntexError = errors.New("'expression' is error syntex")
	NameIsRequired        = commons.IsRequired("name")
	CommandIsRequired     = commons.IsRequired("command")
)

func newTrigger(attributes, options, ctx map[string]interface{}, callback triggerFunc) (*trigger, error) {
	name := commons.GetStringWithDefault(attributes, "name", "")
	if "" == name {
		return nil, NameIsRequired
	}

	if nil == callback {
		return nil, commons.IsRequired("callback")
	}

	expression := commons.GetStringWithDefault(attributes, "expression", "")
	if "" == expression {
		return nil, commons.IsRequired("expression")
	}

	action_specs, e := commons.GetObjects(attributes, "$action")
	if nil != e {
		return nil, commons.IsRequired("$action")
	}
	actions := make([]ExecuteAction, 0, 10)
	for _, spec := range action_specs {
		action, e := newAction(spec, options, ctx)
		if nil != e {
			return nil, e
		}
		actions = append(actions, action)
	}

	if strings.HasPrefix(expression, every) {
		interval, err := time.ParseDuration(expression[len(every):])
		if nil != err {
			return nil, errors.New(ExpressionSyntexError.Error() + ", " + err.Error())
		}

		it := &intervalTrigger{c: make(chan *request),
			interval: interval}

		t := &trigger{name: name,
			expression:  expression,
			attachment:  commons.GetStringWithDefault(attributes, "attachment", ""),
			description: commons.GetStringWithDefault(attributes, "description", ""),
			isRunning:   SRV_INIT,
			callback:    callback,
			actions:     actions,
			start:       it.start,
			stop:        it.stop}
		it.trigger = t
		t.InitLoggerWith(ctx, "log.")
		return t, nil
	}
	return nil, ExpressionSyntexError
}

type intervalTrigger struct {
	*trigger
	interval time.Duration
	c        chan *request
}

func (self *intervalTrigger) start(t *trigger) error {
	go self.run()
	return nil
}

func (self *intervalTrigger) stop(t *trigger) {
	self.send("exit")
	// e := self.send("exit")
	// if nil != e {
	// 	self.DEBUG.Printf("stop failed, %v", e)
	// } else {
	// 	self.DEBUG.Print("recved 'exited' signal")
	// }
}

func (self *intervalTrigger) command(cmd string) error {
	if SRV_RUNNING != atomic.LoadInt32(&self.isRunning) {
		return NotStart
	}
	return self.send(cmd)
}

var (
	NotSend  = errors.New("send to trigger failed.")
	NotStart = errors.New("it is not running.")
)

type request struct {
	cmd    string
	c      chan *request
	result string
}

func (self *intervalTrigger) send(cmd string) error {
	req := &request{
		cmd: cmd,
		c:   make(chan *request, 1)}
	defer close(req.c)
	select {
	case self.c <- req:
		select {
		case res := <-req.c:
			if "ok" == res.result {
				return nil
			}
			return errors.New(res.result)
		case <-time.After(5 * time.Second):
			return commons.TimeoutErr
		}
	default: // it is required, becase run() may exited {status != running}
		return NotSend
	}
}

func (self *intervalTrigger) run() {
	if !atomic.CompareAndSwapInt32(&self.isRunning, SRV_STARTING, SRV_RUNNING) {
		self.DEBUG.Print("it is stopping while starting.")
		return
	}

	defer func() {
		atomic.StoreInt32(&self.isRunning, SRV_INIT)

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
			self.ERROR.Print(buffer.String())
		} else {
			self.DEBUG.Print("exited")
		}
	}()

	ticker := time.NewTicker(self.interval)
	defer ticker.Stop()

	is_running := true
	for is_running {
		select {
		case req := <-self.c:
			if "exit" == req.cmd {
				is_running = false
				req.result = "ok"
				func() {
					defer func() {
						if e := recover(); nil != e {
							self.WARN.Printf("[panic] %v", e)
						}
					}()
					req.c <- req
				}()

			} else {
				self.executeCommand(req)
			}
		case t := <-ticker.C:
			status := atomic.LoadInt32(&self.isRunning)
			if SRV_RUNNING != status {
				self.DEBUG.Printf("status is exited, status = %v", status)
				is_running = false
				break
			}
			self.timeout(t)
		}
	}

	self.DEBUG.Print("stopping")
}

func (self *intervalTrigger) executeCommand(req *request) {
	defer func() {
		if e := recover(); nil != e {
			req.result = "[panic]" + fmt.Sprint(e)
			req.c <- req
		}
	}()

	f := self.commands[req.cmd]
	if nil != f {
		req.result = f(self.trigger)
	} else {
		req.result = "[error]no such command"
	}

	req.c <- req
}

func (self *intervalTrigger) timeout(t time.Time) {
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
			self.ERROR.Print(buffer.String())
		}
	}()

	self.DEBUG.Printf("timeout %s - %s", self.name, self.interval)
	self.callback(t)
}
