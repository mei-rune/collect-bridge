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

type Trigger interface {
	Id() string
	Name() string
	Version() time.Time
	Stats() map[string]interface{}

	Close(reason int)
	CallActions(t time.Time, res interface{})
}

type triggerFunc func(t time.Time) error

type base_trigger struct {
	commons.Logger
	id         string
	name       string
	actions    []*actionWrapper
	callback   triggerFunc
	updated_at time.Time

	expression  string
	attachment  string
	description string

	l          sync.Mutex
	last_error error
}

func (self *base_trigger) Id() string {
	return self.id
}

func (self *base_trigger) Name() string {
	return self.name
}

func (self *base_trigger) Version() time.Time {
	return self.updated_at
}

func (self *base_trigger) Stats() map[string]interface{} {
	res := map[string]interface{}{
		"id":         self.Id(),
		"name":       self.Name(),
		"updated_at": self.updated_at,
		"expression": self.expression}

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

func (self *base_trigger) CallActions(t time.Time, res interface{}) {
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

func (self *base_trigger) callBefore() {
	self.l.Lock()
	defer self.l.Unlock()
	for _, action := range self.actions {
		action.RunBefore()
	}
}
func (self *base_trigger) callAfter() {
	self.l.Lock()
	defer self.l.Unlock()
	for _, action := range self.actions {
		action.RunAfter()
	}
}

func (self *base_trigger) reset(reason int) {
	if CLOSE_REASON_NORMAL == reason {
		return
	}

	for _, action := range self.actions {
		action.reset(reason)
	}
}

const every = "@every "

var (
	ExpressionSyntexError = errors.New("'expression' is error syntex")
	IdIsRequired          = commons.IsRequired("id")
	NameIsRequired        = commons.IsRequired("name")
	CommandIsRequired     = commons.IsRequired("command")
)

func newTrigger(attributes, options, ctx map[string]interface{}, callback triggerFunc) (Trigger, error) {
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
		enabled := commons.GetBoolWithDefault(spec, "enabled", true)

		action, e := newAction(spec, options, ctx)
		if nil != e {
			return nil, errors.New("create action '" + action_id + ":" + action_name + "' failed, " + e.Error())
		}
		if !enabled {
			reset(action, CLOSE_REASON_DISABLED)
		}
		actions = append(actions, &actionWrapper{id: action_id, name: action_name, enabled: enabled, action: action})
	}

	if 0 == len(actions) {
		return nil, errors.New("actions is empty.")
	}

	enabled := false
	for _, action := range actions {
		if action.enabled {
			enabled = true
			break
		}
	}

	if !enabled {
		return nil, errors.New("all actions is disable.")
	}

	if strings.HasPrefix(expression, every) {
		interval, err := time.ParseDuration(expression[len(every):])
		if nil != err {
			return nil, errors.New(ExpressionSyntexError.Error() + ", " + err.Error())
		}

		it := &intervalTrigger{base_trigger: &base_trigger{id: id,
			name:        name,
			expression:  expression,
			attachment:  commons.GetStringWithDefault(attributes, "attachment", ""),
			description: commons.GetStringWithDefault(attributes, "description", ""),
			updated_at:  commons.GetTimeWithDefault(attributes, "updated_at", time.Time{}),
			callback:    callback,
			actions:     actions},
			interval: interval}
		//status:   commons.SRV_INIT}
		it.base_trigger.InitLoggerWith(ctx, "log.")
		err = it.Init()
		if nil != err {
			return nil, err
		}
		return it, nil
	}
	return nil, ExpressionSyntexError
}

type intervalTrigger struct {
	*base_trigger
	interval time.Duration
	c        chan int
	wait     sync.WaitGroup

	//status            int32
	max_used_duration int64
	begin_fired_at    int64
	end_fired_at      int64
}

func (self *intervalTrigger) Init() error {
	self.c = make(chan int)
	go self.run()
	self.wait.Add(1)
	return nil
}

func (self *intervalTrigger) Close(reason int) {
	self.c <- 1
	self.wait.Wait()
	close(self.c)
	self.reset(reason)
}

func (self *intervalTrigger) run() {
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
		} else {
			self.DEBUG.Print("exited")
		}

		self.wait.Done()
	}()

	ticker := time.NewTicker(self.interval)
	defer ticker.Stop()

	is_running := true
	for is_running {
		select {
		case <-self.c:
			is_running = false
		case t := <-ticker.C:
			self.timeout(t)
		}
	}

	self.DEBUG.Print("stopping")
}

func (self *intervalTrigger) Stats() map[string]interface{} {
	m := self.base_trigger.Stats()
	m["max_used_duration"] = strconv.FormatInt(atomic.LoadInt64(&self.max_used_duration), 10) + "s"
	m["begin_fired_at"] = time.Unix(atomic.LoadInt64(&self.begin_fired_at), 0).Format(time.RFC3339Nano)
	m["end_fired_at"] = time.Unix(atomic.LoadInt64(&self.end_fired_at), 0).Format(time.RFC3339Nano)
	//m["status"] = commons.ToStatusString(int(atomic.LoadInt32(&self.status)))
	return m
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
