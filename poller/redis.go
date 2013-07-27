package poller

import (
	"bytes"
	"errors"
	"expvar"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var redis_error = expvar.NewString("redis")

type redis_gateway struct {
	Address string
	c       chan []string
	status  int32
	wait    sync.WaitGroup
}

func (self *redis_gateway) isRunning() bool {
	return 1 == atomic.LoadInt32(&self.status)
}

func (self *redis_gateway) run() {
	defer func() {
		log.Println("redis client is exit.")
		close(self.c)
		self.wait.Done()
	}()

	ticker := time.NewTicker(1 * time.Second)

	go func() {
		defer func() {
			if o := recover(); nil != o {
				log.Println("[panic]", o)
			}
			self.wait.Done()
			ticker.Stop()
		}()

		for self.isRunning() {
			<-ticker.C
			self.c <- nil
		}
	}()

	self.wait.Add(1)

	for self.isRunning() {
		self.runOnce()
	}
}

func (self *redis_gateway) recvCommands(max_size int) [][]string {
	commands := make([][]string, 0, max_size)

	for self.isRunning() {
		cmd := <-self.c
		if nil == cmd || 0 == len(cmd) {
			if 0 != len(commands) {
				return commands
			}
			continue
		}

		commands = append(commands, cmd)
		if max_size < len(commands) {
			return commands
		}
	}
	return commands
}

func (self *redis_gateway) runOnce() {
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
			redis_error.Set(msg)
			log.Println(msg)
		}
	}()

	c, err := redis.DialTimeout("tcp", self.Address, 0, 1*time.Second, 1*time.Second)
	if err != nil {
		msg := fmt.Sprintf("[redis] connect to '%s' failed, %v", self.Address, err)
		redis_error.Set(msg)
		log.Println(msg)
		return
	}

	redis_error.Set("")

	for self.isRunning() {
		commands := self.recvCommands(10)
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
			case 6:
				_, err = c.Do(cmd[0], cmd[1], cmd[2], cmd[3], cmd[4], cmd[5])
			default:
				err = errors.New("argument length is error.")
			}

			if nil != err {
				msg := fmt.Sprintf("[redis] do command '%s' failed, %v", cmd[0], err)
				redis_error.Set(msg)
				log.Println(msg)
				return
			}
		}
	}
}

func (self *redis_gateway) Close() {
	atomic.StoreInt32(&self.status, 0)
	self.wait.Wait()
}

func newRedis(address string) (*redis_gateway, error) {
	client := &redis_gateway{Address: address, c: make(chan []string, 3000), status: 1}
	go client.run()
	client.wait.Add(1)
	return client, nil
}
