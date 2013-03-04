package poller

import (
	"bytes"
	"commons"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"runtime"
	"time"
)

type Redis struct {
	Address   string
	ch        chan []string
	is_runing bool
}

func (self *Redis) run() {
	defer func() {
		commons.Log.INFO.Print("redis client is exit.")
		close(self.ch)
	}()

	for {
		time.Sleep(1 * time.Second)
		self.runOnce()
	}
}

func (self *Redis) recvCommands() [][]string {
	commands := make([][]string, 0, 30)
	interval := 10 * time.Millisecond
	for self.is_runing {
		select {
		case c := <-self.ch:
			interval = 10 * time.Millisecond
			commands = append(commands, c)
		case <-time.After(interval):
			if 0 != len(commands) {
				return commands
			} else {
				interval = 1 * time.Second
			}
		}
	}
	return nil
}

func (self *Redis) runOnce() {
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
			commons.Log.ERROR.Print(buffer.String())
		}
	}()

	c, err := redis.DialTimeout("tcp", self.Address, 0, 1*time.Second, 1*time.Second)
	if err != nil {
		commons.Log.ERROR.Print("[redis] connect to '%s' failed, %v", self.Address, err)
		return
	}
	for self.is_runing {
		commands := self.recvCommands()
		for _, cmd := range commands {
			switch len(cmd) {
			case 1:
				_, err = c.Do(cmd[0])
			case 2:
				_, err = c.Do(cmd[0], cmd[1])
			case 3:
				_, err = c.Do(cmd[0], cmd[1], cmd[2])
			case 4:
				_, err = c.Do(cmd[0], cmd[1], cmd[2], cmd[3])
			case 5:
				_, err = c.Do(cmd[0], cmd[1], cmd[2], cmd[3], cmd[4])
			}
			if nil != err {
				commons.Log.ERROR.Print("[redis] do command '%s' failed, %v", cmd[0], err)
				return
			}
		}
	}
}

func NewRedis(address string) (chan []string, error) {
	redis := &Redis{Address: address, ch: make(chan []string, 3000), is_runing: true}
	go redis.run()
	return redis.ch, nil
}
