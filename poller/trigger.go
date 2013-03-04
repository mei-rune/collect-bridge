package poller

import (
	"bytes"
	"commons"
	"commons/errutils"
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

type TriggerFunc func(t time.Time)

type Trigger struct {
	commons.Logger
	Name      string
	Actions   []ExecuteAction
	Callback  TriggerFunc
	IsRunning int32

	Commands    map[string]func(trigger *Trigger) string
	Expression  string
	Attachment  string
	Description string

	start func(t *Trigger) error
	stop  func(t *Trigger)
}

func (self *Trigger) Start() error {
	if 1 == atomic.LoadInt32(&self.IsRunning) {
		return nil
	}

	return self.start(self)
}

func (self *Trigger) Stop() {
	if 0 == atomic.LoadInt32(&self.IsRunning) {
		return
	}

	self.stop(self)
}

func (self *Trigger) CallActions(t time.Time, res interface{}) {
	if nil == res {
		self.WARN.Print("result of '" + self.Name + "' is nil")
		return
	}
	if nil == self.Actions || 0 == len(self.Actions) {
		self.WARN.Print("actions of '" + self.Name + "' is empty")
		return
	}
	for _, action := range self.Actions {
		action.Run(t, res)
	}
}

const every = "@every "

var (
	ExpressionSyntexError = errutils.BadRequest("'expression' is error syntex")
	NameIsRequired        = errutils.IsRequired("name")
	CommandIsRequired     = errutils.IsRequired("command")
)

func NewTrigger(attributes map[string]interface{}, callback TriggerFunc, ctx map[string]interface{}) (*Trigger, error) {
	name := commons.GetString(attributes, "name", "")
	if "" == name {
		return nil, NameIsRequired
	}

	if nil == callback {
		return nil, errutils.IsRequired("callback")
	}

	expression := commons.GetString(attributes, "expression", "")
	if "" == expression {
		return nil, errutils.IsRequired("expression")
	}

	if strings.HasPrefix(expression, every) {
		interval, err := commons.ParseDuration(expression[len(every):])
		if nil != err {
			return nil, errutils.BadRequest(ExpressionSyntexError.Error() + ", " + err.Error())
		}

		intervalTrigger := &IntervalTrigger{control_ch: make(chan string, 1),
			control_resp_ch: make(chan string),
			interval:        interval}

		trigger := &Trigger{Name: name,
			Expression:  expression,
			Attachment:  commons.GetString(attributes, "attachment", ""),
			Description: commons.GetString(attributes, "description", ""),
			Callback:    callback,
			start: func(t *Trigger) error {
				go intervalTrigger.run()
				return nil
			},
			stop: func(t *Trigger) {
				intervalTrigger.control_ch <- "exit"
				<-intervalTrigger.control_resp_ch
			}}
		intervalTrigger.Trigger = trigger
		return trigger, nil
	}
	return nil, ExpressionSyntexError
}

type IntervalTrigger struct {
	*Trigger
	interval        time.Duration
	control_ch      chan string
	control_resp_ch chan string
}

func (self *IntervalTrigger) run() {
	if !atomic.CompareAndSwapInt32(&self.IsRunning, 0, 1) {
		return
	}

	defer func() {
		atomic.StoreInt32(&self.IsRunning, 0)
		self.control_resp_ch <- "ok"
	}()

	is_running := true
	for is_running {
		select {
		case cmd := <-self.control_ch:
			if "exit" == cmd {
				is_running = true
			} else {
				self.executeCommand(cmd)
			}
		case <-time.After(self.interval):
			self.timeout()
		}
	}
}

func (self *IntervalTrigger) executeCommand(nm string) {
	res := "[error]no such command"
	defer func() {
		if e := recover(); nil != e {
			self.control_resp_ch <- "[error]" + fmt.Sprint(e)
		} else {
			self.control_resp_ch <- res
		}
	}()

	f := self.Commands[nm]
	if nil != f {
		res = f(self.Trigger)
	}
}

func (self *IntervalTrigger) timeout() {
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

	fmt.Printf("timeout %s - %s - %v\n", self.Name, self.interval, self.Callback)
	self.INFO.Printf("timeout %s - %s", self.Name, self.interval)
	self.Callback(time.Now())
}
