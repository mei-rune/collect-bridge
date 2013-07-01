package poller

import (
	"commons"
	"fmt"
	"time"

	"errors"
)

type RedisAction struct {
	Name        string
	Description string
	Command     string
	channel     chan []string
	Arg0        string
	Arg1        string
	Arg2        string
	Arg3        string
	Arg4        string
}

func (self *RedisAction) Run(t time.Time, value interface{}) {
	s := fmt.Sprint(value)
	if "" == self.Arg0 {
		self.channel <- []string{self.Command, s}
	} else if "" == self.Arg1 {
		self.channel <- []string{self.Command, self.Arg0, s}
	} else if "" == self.Arg2 {
		self.channel <- []string{self.Command, self.Arg0, self.Arg1, s}
	} else if "" == self.Arg3 {
		self.channel <- []string{self.Command, self.Arg0, self.Arg1, self.Arg2, s}
	} else if "" == self.Arg4 {
		self.channel <- []string{self.Command, self.Arg0, self.Arg1, self.Arg2, self.Arg3, s}
	} else {
		self.channel <- []string{self.Command, self.Arg0, self.Arg1, self.Arg2, self.Arg3, self.Arg4, s}
	}
}

var redis_commands = map[string]bool{"APPEND": true,
	"INCR": true, "DECR": true, "INCRBY": true, "SETBIT": true,
	"LPUSH": true, "RPUSH": true,
	"SADD": true, "SET": true, "GET": true, "MSET": true,
	"HSET": true, "HMSET": true}

func IsRedisCommand(cmd string) bool {
	return redis_commands[cmd]
}

func NewRedisAction(attributes, ctx map[string]interface{}) (ExecuteAction, error) {
	name, e := commons.GetString(attributes, "name")
	if nil != e {
		return nil, NameIsRequired
	}

	command, e := commons.GetString(attributes, "command")
	if nil != e {
		return nil, CommandIsRequired
	}
	if !IsRedisCommand(command) {
		return nil, fmt.Errorf("'%s' is not a redis command", command)
	}

	c := ctx["redis_channel"]
	if nil == c {
		return nil, errors.New("'redis_channel' is nil")
	}
	channel, ok := c.(chan []string)
	if !ok {
		return nil, errors.New("'redis_channel' is not a chan []stirng")
	}

	return &RedisAction{Name: name,
		Description: commons.GetStringWithDefault(attributes, "description", ""),
		Command:     command,
		channel:     channel,
		Arg0:        commons.GetStringWithDefault(attributes, "arg0", ""),
		Arg1:        commons.GetStringWithDefault(attributes, "arg1", ""),
		Arg2:        commons.GetStringWithDefault(attributes, "arg2", ""),
		Arg3:        commons.GetStringWithDefault(attributes, "arg3", ""),
		Arg4:        commons.GetStringWithDefault(attributes, "arg4", "")}, nil
}
