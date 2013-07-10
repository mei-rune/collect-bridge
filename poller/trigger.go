package poller

import (
	"bytes"
	"commons"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type triggerFunc func(t time.Time) error

type trigger struct {
	commons.Logger
	id         string
	name       string
	actions    []*actionWrapper
	callback   triggerFunc
	status     int32
	updated_at time.Time

	commands    map[string]func(t *trigger) string
	expression  string
	attachment  string
	description string

	l          sync.Mutex
	last_error error
	start      func(t *trigger) error
	stop       func(t *trigger)
	stats      func(m map[string]interface{})
}

func (self *trigger) Id() string {
	return self.id
}

func (self *trigger) Name() string {
	return self.name
}

func (self *trigger) Version() time.Time {
	return self.updated_at
}

func (self *trigger) Stats() map[string]interface{} {
	res := map[string]interface{}{
		"id":         self.Id(),
		"name":       self.Name(),
		"updated_at": self.updated_at,
		"expression": self.expression,
		"status":     commons.ToStatusString(int(atomic.LoadInt32(&self.status)))}

	if nil != self.stats {
		self.stats(res)
	}

	self.l.Lock()
	defer self.l.Unlock()

	if nil != self.last_error {
		res["error"] = self.last_error.Error()
	}

	if nil != self.actions && 0 != len(self.actions) {
		actions := make([]interface{}, 0, len(self.actions))
		for _, action := range self.actions {
			actions = append(actions, action.Stats())
		}
		res["actions"] = actions
	}

	return res
}

func (self *trigger) Start() error {
	if !atomic.CompareAndSwapInt32(&self.status, commons.SRV_INIT, commons.SRV_STARTING) {
		return nil
	}

	e := self.start(self)
	if nil != e {
		atomic.StoreInt32(&self.status, commons.SRV_INIT)
	}
	return e
}

func (self *trigger) Stop() {
	if atomic.CompareAndSwapInt32(&self.status, commons.SRV_STARTING, commons.SRV_INIT) {
		self.DEBUG.Print("it is starting")
		return
	}

	if !atomic.CompareAndSwapInt32(&self.status, commons.SRV_RUNNING, commons.SRV_STOPPING) {
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

	self.callBefore()

	for _, action := range self.actions {
		action.Run(t, res)
	}

	self.callAfter()
}

func (self *trigger) callBefore() {
	self.l.Lock()
	defer self.l.Unlock()
	for _, action := range self.actions {
		action.RunBefore()
	}
}
func (self *trigger) callAfter() {
	self.l.Lock()
	defer self.l.Unlock()
	for _, action := range self.actions {
		action.RunAfter()
	}
}

const every = "@every "

var (
	ExpressionSyntexError = errors.New("'expression' is error syntex")
	IdIsRequired          = commons.IsRequired("id")
	NameIsRequired        = commons.IsRequired("name")
	CommandIsRequired     = commons.IsRequired("command")
)

func newTrigger(attributes, options, ctx map[string]interface{}, callback triggerFunc) (*trigger, error) {
	id := commons.GetStringWithDefault(attributes, "id", "")
	if "" == id {
		return nil, IdIsRequired
	}
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
	actions := make([]*actionWrapper, 0, 10)
	for _, spec := range action_specs {

		action_id := commons.GetStringWithDefault(spec, "id", "unknow_id")
		action_name := commons.GetStringWithDefault(spec, "name", "unknow_name")

		action, e := newAction(spec, options, ctx)
		if nil != e {
			return nil, errors.New("create action '" + action_id + ":" + action_name + "' failed, " + e.Error())
		}
		actions = append(actions, &actionWrapper{id: action_id, name: action_name, action: action})
	}

	if 0 == len(actions) {
		return nil, errors.New("actions is empty.")
	}

	if strings.HasPrefix(expression, every) {
		interval, err := time.ParseDuration(expression[len(every):])
		if nil != err {
			return nil, errors.New(ExpressionSyntexError.Error() + ", " + err.Error())
		}

		it := &intervalTrigger{c: make(chan *request),
			interval: interval}

		t := &trigger{id: id,
			name:        name,
			expression:  expression,
			attachment:  commons.GetStringWithDefault(attributes, "attachment", ""),
			description: commons.GetStringWithDefault(attributes, "description", ""),
			updated_at:  commons.GetTimeWithDefault(attributes, "updated_at", time.Time{}),
			status:      commons.SRV_INIT,
			callback:    callback,
			actions:     actions,
			stats:       it.onStats,
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

	max_used_duration int64
	begin_fired_at    int64
	end_fired_at      int64
}

func (self *intervalTrigger) start(t *trigger) error {
	go self.run()
	return nil
}

func (self *intervalTrigger) stop(t *trigger) {
	self.send("exit")
}

func (self *intervalTrigger) command(cmd string) error {
	if commons.SRV_RUNNING != atomic.LoadInt32(&self.status) {
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
	if !atomic.CompareAndSwapInt32(&self.status, commons.SRV_STARTING, commons.SRV_RUNNING) {
		self.DEBUG.Print("it is stopping while starting.")
		return
	}

	defer func() {
		atomic.StoreInt32(&self.status, commons.SRV_INIT)

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
			status := atomic.LoadInt32(&self.status)
			if commons.SRV_RUNNING != status {
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

func (self *intervalTrigger) onStats(m map[string]interface{}) {
	m["max_used_duration"] = strconv.FormatInt(atomic.LoadInt64(&self.max_used_duration), 10) + "s"
	m["begin_fired_at"] = time.Unix(atomic.LoadInt64(&self.begin_fired_at), 0).Format(time.RFC3339Nano)
	m["end_fired_at"] = time.Unix(atomic.LoadInt64(&self.end_fired_at), 0).Format(time.RFC3339Nano)
}

func (self *intervalTrigger) timeout(t time.Time) {
	startedAt := t.Unix()

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
			msg := buffer.String()
			self.set_last_error(errors.New(msg))
			self.ERROR.Print(msg)
		}

		now := time.Now().Unix()
		atomic.StoreInt64(&self.end_fired_at, now)
		if self.max_used_duration < now-startedAt {
			atomic.StoreInt64(&self.max_used_duration, now-startedAt)
		}
	}()

	atomic.StoreInt64(&self.begin_fired_at, startedAt)

	self.DEBUG.Printf("timeout %s - %s", self.name, self.interval)
	self.set_last_error(self.callback(t))
}

func (self *intervalTrigger) set_last_error(e error) {
	self.l.Lock()
	defer self.l.Unlock()
	self.last_error = e
}
