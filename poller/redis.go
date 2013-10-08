package poller

import (
	"bytes"
	"errors"
	"expvar"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"log"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var redis_error = expvar.NewString("redis_gateway")

type redis_gateway struct {
	Address    string
	c          chan []string
	is_closed  int32
	wait       sync.WaitGroup
	last_error *expvar.String
}

func (self *redis_gateway) isRunning() bool {
	return 0 == atomic.LoadInt32(&self.is_closed)
}

func (self *redis_gateway) Close() {
	if atomic.CompareAndSwapInt32(&self.is_closed, 0, 1) {
		return
	}
	close(self.c)
	self.wait.Wait()
}

func (self *redis_gateway) run() {
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
			self.last_error.Set(msg)
			log.Println(msg)

			if !self.isRunning() {
				os.Exit(-1)
			}
		}

		log.Println("redis client is exit.")
		atomic.StoreInt32(&self.is_closed, 1)
		self.wait.Done()
	}()

	for self.isRunning() {
		self.runOnce()
	}
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
			self.last_error.Set(msg)
			log.Println(msg)
		}
	}()

	c, err := redis.DialTimeout("tcp", self.Address, 0, 1*time.Second, 1*time.Second)
	if err != nil {
		msg := fmt.Sprintf("[redis] connect to '%s' failed, %v", self.Address, err)
		self.last_error.Set(msg)
		log.Println(msg)
		return
	}

	self.last_error.Set("")

	cached_objects := make([][]string, 0, 50)
	is_running := true
	for is_running {
		cmd, ok := <-self.c
		if !ok {
			is_running = false
			break
		}

		if nil == cmd || 0 == len(cmd) {
			break
		}

		objects := append(cached_objects[0:0], cmd)
		one_running := true
		for one_running {
			select {
			case cmd, ok := <-self.c:
				if !ok {
					one_running = false
					is_running = false
					break
				}

				if nil == cmd || 0 == len(cmd) {
					break
				}

				objects = append(objects, cmd)
				if 50 < len(objects) {
					self.execute(objects, c)
					objects = cached_objects[0:0]
				}
			default:
				one_running = false
			}
		}
		if 0 != len(objects) {
			self.execute(objects, c)
		}
	}
}

func (self *redis_gateway) execute(commands [][]string, c redis.Conn) {
	if 0 == len(commands) {
		return
	}
	var err error
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
			self.last_error.Set(msg)
			log.Println(msg)
			return
		}
	}
}

func newRedis(address string) (*redis_gateway, error) {
	client := &redis_gateway{Address: address, c: make(chan []string, 3000), last_error: redis_error}
	go client.run()
	client.wait.Add(1)
	return client, nil
}
