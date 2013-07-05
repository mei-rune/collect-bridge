package poller

import (
	"commons"
	"fmt"
	"time"

	"encoding/json"
	"errors"
)

type toJsoner interface {
	ToJson() string
}

type toMapper interface {
	ToMap() map[string]interface{}
}

type valueGetter interface {
	Value() commons.Any
}

type redisAction struct {
	name        string
	description string
	options     commons.Map
	channel     chan<- []string
	command     string
	arguments   []string
}

func toMap(value interface{}) (commons.Map, error) {
	if vg, ok := value.(valueGetter); ok {
		values := vg.Value()
		if nil == values || nil == values.AsInterface() {
			return nil, errors.New("it is nil")
		}
		m, e := values.AsObject()
		if nil != e {
			return nil, e
		}
		return commons.InterfaceMap(m), nil
	} else if m, ok := value.(map[string]interface{}); ok {
		return commons.InterfaceMap(m), nil
	} else if m, ok := value.(map[string]string); ok {
		return commons.StringMap(m), nil
	} else {
		return nil, fmt.Errorf("unsupport type '%T' ", value)
	}
}

func (self *redisAction) Run(t time.Time, value interface{}) error {
	var values commons.Map = nil
	var err error

	commands := make([]string, 0, 1+len(self.arguments))
	commands = append(commands, self.command)
	for _, s := range self.arguments {
		if '$' != s[0] {
			commands = append(commands, s)
			continue
		}

		if "$$" == s {
			if result, ok := value.(toMapper); ok {
				m := result.ToMap()
				if nil == m {
					m = map[string]interface{}{}
				}
				if nil != self.options {
					self.options.CopyTo(m)
				}
				js, e := json.Marshal(m)
				if nil != e {
					return e
				}
				commands = append(commands, string(js))
			} else if js, ok := value.(toJsoner); ok {
				commands = append(commands, js.ToJson())
			} else {
				commands = append(commands, fmt.Sprint(value))
			}
			continue
		}

		if nil != self.options {
			v, e := self.options.GetString(s[1:])
			if nil == e {
				commands = append(commands, v)
				continue
			}
		}

		if nil == values {
			values, err = toMap(value)
			if nil != err {
				return err
			}
		}

		v, e := values.GetString(s[1:])
		if nil != e {
			return e
		}
		commands = append(commands, v)
	}

	self.channel <- commands
	return nil
}

var redis_commands = map[string]bool{"APPEND": true,
	"INCR": true, "DECR": true, "INCRBY": true, "SETBIT": true,
	"LPUSH": true, "RPUSH": true,
	"SADD": true, "SET": true, "GET": true, "MSET": true,
	"HSET": true, "HMSET": true, "PUBLISH": true}

func isRedisCommand(cmd string) bool {
	return true
	//return redis_commands[cmd]
}

func newRedisAction(attributes, options, ctx map[string]interface{}) (ExecuteAction, error) {
	name, e := commons.GetString(attributes, "name")
	if nil != e {
		return nil, NameIsRequired
	}

	command, e := commons.GetString(attributes, "command")
	if nil != e {
		return nil, CommandIsRequired
	}
	if !isRedisCommand(command) {
		return nil, fmt.Errorf("'%s' is not a redis command", command)
	}

	c := ctx["redis_channel"]
	if nil == c {
		return nil, errors.New("'redis_channel' is nil")
	}
	channel, ok := c.(chan<- []string)
	if !ok {
		return nil, errors.New("'redis_channel' is not a chan []stirng")
	}

	return &redisAction{name: name,
		description: commons.GetStringWithDefault(attributes, "description", ""),
		options:     commons.InterfaceMap(options),
		channel:     channel,
		command:     command,
		arguments: newRedisArguments(commons.GetStringWithDefault(attributes, "arg0", ""),
			commons.GetStringWithDefault(attributes, "arg1", ""),
			commons.GetStringWithDefault(attributes, "arg2", ""),
			commons.GetStringWithDefault(attributes, "arg3", ""),
			commons.GetStringWithDefault(attributes, "arg4", ""),
			commons.GetStringWithDefault(attributes, "arg5", ""),
			commons.GetStringWithDefault(attributes, "arg6", ""))}, nil
}

func newRedisArguments(arg0, arg1, arg2, arg3, arg4, arg5, arg6 string) []string {
	if "" == arg0 {
		return []string{}
	} else if "" == arg1 {
		return []string{arg0}
	} else if "" == arg2 {
		return []string{arg0, arg1}
	} else if "" == arg3 {
		return []string{arg0, arg1, arg2}
	} else if "" == arg4 {
		return []string{arg0, arg1, arg2, arg3}
	} else if "" == arg5 {
		return []string{arg0, arg1, arg2, arg3, arg4}
	} else if "" == arg6 {
		return []string{arg0, arg1, arg2, arg3, arg4, arg5}
	} else {
		return []string{arg0, arg1, arg2, arg3, arg4, arg5, arg6}
	}
}
