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
	TG_INIT     = 0
	TG_STARTING = 1
	TG_RUNNING  = 2
	TG_STOPPING = 3
)

type triggerFunc func(t time.Time)

type trigger struct {
	commons.Logger
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

func (self *trigger) Start() error {
	if !atomic.CompareAndSwapInt32(&self.isRunning, TG_INIT, TG_STARTING) {
		return nil
	}

	e := self.start(self)
	if nil != e {
		atomic.StoreInt32(&self.isRunning, TG_INIT)
	}
	return e
}

func (self *trigger) Stop() {
	if !atomic.CompareAndSwapInt32(&self.isRunning, TG_STARTING, TG_STOPPING) &&
		!atomic.CompareAndSwapInt32(&self.isRunning, TG_RUNNING, TG_STOPPING) {
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

func newTrigger(attributes map[string]interface{}, callback triggerFunc, ctx map[string]interface{}) (*trigger, error) {
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
		action, e := NewAction(spec, ctx)
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

		it := &intervalTrigger{control_ch: make(chan string),
			control_resp_ch: make(chan string),
			interval:        interval}

		t := &trigger{name: name,
			expression:  expression,
			attachment:  commons.GetStringWithDefault(attributes, "attachment", ""),
			description: commons.GetStringWithDefault(attributes, "description", ""),
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
	interval        time.Duration
	control_ch      chan string
	control_resp_ch chan string
}

func (self *intervalTrigger) start(t *trigger) error {
	go self.run()
	return nil
}

func (self *intervalTrigger) stop(t *trigger) {
	select {
	case self.control_ch <- "exit":
		select {
		case <-self.control_resp_ch:
		}
	}
}

func (self *intervalTrigger) run() {
	if !atomic.CompareAndSwapInt32(&self.isRunning, TG_STARTING, TG_RUNNING) {
		return
	}

	defer func() {
		atomic.StoreInt32(&self.isRunning, TG_INIT)

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

	ticker := time.NewTicker(self.interval)
	defer ticker.Stop()

	is_running := true
	for is_running {
		select {
		case cmd := <-self.control_ch:
			if "exit" == cmd {
				is_running = false
				select {
				case self.control_resp_ch <- "ok":
				}
			} else {
				self.executeCommand(cmd)
			}
		case t := <-ticker.C:
			if TG_RUNNING != atomic.LoadInt32(&self.isRunning) {
				is_running = false
				select {
				case self.control_resp_ch <- "ok":
				}
				break
			}
			self.timeout(t)
		}
	}
}

func (self *intervalTrigger) executeCommand(nm string) {
	res := "[error]no such command"
	defer func() {
		if e := recover(); nil != e {
			self.control_resp_ch <- "[panic]" + fmt.Sprint(e)
		} else {
			self.control_resp_ch <- res
		}
	}()

	f := self.commands[nm]
	if nil != f {
		res = f(self.trigger)
	}
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
