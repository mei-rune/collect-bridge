package poller

import (
	"commons"
	"errors"
	"sync"
	"time"
)

const (
	CLOSE_REASON_NORMAL   = -1
	CLOSE_REASON_UNKNOW   = 0
	CLOSE_REASON_DISABLED = 1
	CLOSE_REASON_DELETED  = 2
	CLOSE_REASON_MAX      = 2
)

type Job interface {
	Interupt()
	Close(reason int)

	Id() string
	Name() string
	Stats() map[string]interface{}
	Version() time.Time
}

type ValueResult interface {
	ErrorCode() int
	ErrorMessage() string
	HasError() bool
	Error() commons.RuntimeError
	Value() commons.Any
	InterfaceValue() interface{}
	CreatedAt() time.Time
}

var (
	AlreadyStarted    = errors.New("It is already started.")
	AlreadyClosed     = errors.New("It is already closed.")
	AllDisabled       = errors.New("all actions is disabled.")
	ActionsIsEmpty    = errors.New("actions is empty.")
	IdIsRequired      = commons.IsRequired("id")
	NameIsRequired    = commons.IsRequired("name")
	CommandIsRequired = commons.IsRequired("command")
)

type baseJob struct {
	id         string
	name       string
	actions    []*actionWrapper
	updated_at time.Time

	expression  string
	attachment  string
	description string

	l          sync.Mutex
	last_error string
}

func (self *baseJob) Id() string {
	return self.id
}

func (self *baseJob) Name() string {
	return self.name
}

func (self *baseJob) Version() time.Time {
	return self.updated_at
}

func (self *baseJob) Stats() map[string]interface{} {
	res := map[string]interface{}{
		"id":         self.Id(),
		"name":       self.Name(),
		"updated_at": self.updated_at,
		"expression": self.expression}

	self.l.Lock()
	defer self.l.Unlock()

	if 0 != len(self.last_error) {
		res["error"] = self.last_error
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

func (self *baseJob) callActions(t time.Time, res interface{}) {
	if nil == res {
		return
	}
	if nil == self.actions || 0 == len(self.actions) {
		return
	}

	self.callBefore()
	for _, action := range self.actions {
		action.Run(t, res)
	}
	self.callAfter()
}

func (self *baseJob) callBefore() {
	self.l.Lock()
	defer self.l.Unlock()
	for _, action := range self.actions {
		action.RunBefore()
	}
}
func (self *baseJob) callAfter() {
	self.l.Lock()
	defer self.l.Unlock()
	for _, action := range self.actions {
		action.RunAfter()
	}
}

func (self *baseJob) reset(reason int) {
	if CLOSE_REASON_NORMAL == reason {
		return
	}

	for _, action := range self.actions {
		action.reset(reason)
	}
}

func (self *baseJob) set_last_error(e string) {
	self.l.Lock()
	defer self.l.Unlock()
	self.last_error = e
}

func newBase(attributes, options, ctx map[string]interface{}) (*baseJob, error) {
	id := commons.GetStringWithDefault(attributes, "id", "")
	if "" == id {
		return nil, IdIsRequired
	}
	name := commons.GetStringWithDefault(attributes, "name", "")
	if "" == name {
		return nil, NameIsRequired
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
			// TODO:  请尽快重构它。
			reset(action, CLOSE_REASON_DISABLED)
		}
		actions = append(actions, &actionWrapper{id: action_id, name: action_name, enabled: enabled, action: action})
	}

	if 0 == len(actions) {
		return nil, ActionsIsEmpty
	}

	enabled := false
	for _, action := range actions {
		if action.enabled {
			enabled = true
			break
		}
	}

	if !enabled {
		return nil, AllDisabled
	}
	return &baseJob{id: id,
		name:        name,
		expression:  expression,
		attachment:  commons.GetStringWithDefault(attributes, "attachment", ""),
		description: commons.GetStringWithDefault(attributes, "description", ""),
		updated_at:  commons.GetTimeWithDefault(attributes, "updated_at", time.Time{}),
		actions:     actions}, nil
}
