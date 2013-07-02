package poller

import (
	"commons"
	"fmt"
	"time"

	"errors"
)

type jsonString interface {
	ToJson() string
}

type valueGetter interface {
	Value() commons.Any
}

type redisAction struct {
	name        string
	description string
	channel     chan<- []string
	command     string
	arguments   []string

	last_error error
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

func (self *redisAction) Run(t time.Time, value interface{}) {
	var values commons.Map = nil

	commands := make([]string, 0, 1+len(self.arguments))
	commands = append(commands, self.command)
	for _, s := range self.arguments {
		if '$' == s[0] {
			if "$$" == s {
				if js, ok := value.(jsonString); ok {
					commands = append(commands, js.ToJson())
				} else {
					commands = append(commands, fmt.Sprint(value))
				}
			} else {
				if nil == values {
					var e error
					values, e = toMap(value)
					if nil != e {
						self.last_error = e
						return
					}
				}

				v, e := values.GetString(s[1:])
				if nil != e {
					self.last_error = errors.New("'" + s[1:] + "' is required, " + e.Error())
					return
				}
				commands = append(commands, v)
			}
		} else {
			commands = append(commands, s)
		}
	}

	self.channel <- commands
	self.last_error = nil
}

var redis_commands = map[string]bool{"APPEND": true,
	"INCR": true, "DECR": true, "INCRBY": true, "SETBIT": true,
	"LPUSH": true, "RPUSH": true,
	"SADD": true, "SET": true, "GET": true, "MSET": true,
	"HSET": true, "HMSET": true}

func isRedisCommand(cmd string) bool {
	return redis_commands[cmd]
}

func newRedisAction(attributes, ctx map[string]interface{}) (ExecuteAction, error) {
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
