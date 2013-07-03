package poller

import (
	"commons"
	"github.com/garyburd/redigo/redis"
	"reflect"
	"testing"
	"time"
)

var redis_address = *redisAddress //"127.0.0.1:6379"

func checkResult(t *testing.T, c redis.Conn, cmd, key, excepted string) {
	reply, err := c.Do(cmd, key)
	s, err := redis.String(reply, err)
	if nil != err {
		t.Errorf("GET %s failed, %v", key, err)
	} else if excepted != s {
		t.Errorf("check %s failed, actual is %v, excepted is %v", key, reply, excepted)
	}
}

func TestRedis(t *testing.T) {
	redis_channel, err := newRedis(redis_address)
	if nil != err {
		t.Error(err)
		return
	}
	redis_channel <- []string{"SET", "a1", "1223"}
	redis_channel <- []string{"SET", "a2", "1224"}
	redis_channel <- []string{"SET", "a3", "1225"}
	redis_channel <- []string{"SET", "a4", "1226"}
	redis_channel <- []string{"SET", "a5", "1227"}

	time.Sleep(2 * time.Second)

	c, err := redis.DialTimeout("tcp", redis_address, 0, 1*time.Second, 1*time.Second)
	if err != nil {
		t.Errorf("[redis] connect to '%s' failed, %v", redis_address, err)
		return
	}

	checkResult(t, c, "GET", "a1", "1223")
	checkResult(t, c, "GET", "a2", "1224")
	checkResult(t, c, "GET", "a3", "1225")
	checkResult(t, c, "GET", "a4", "1226")
	checkResult(t, c, "GET", "a5", "1227")
}

func TestRedisAction(t *testing.T) {
	ch := make(chan []string, 1)

	action, e := newRedisAction(map[string]interface{}{
		"name":    "this is a test redis action",
		"command": "SET",
		"arg0":    "sdfs",
		"arg1":    "$$",
		"arg2":    "arg2",
		"arg3":    "$op",
		"arg4":    "arg3",
		"arg5":    "$name"},
		map[string]interface{}{"op": "option1"},
		map[string]interface{}{"redis_channel": forward(ch)})

	if nil != e {
		t.Error(e)
		return
	}

	result := commons.Return(map[string]interface{}{"name": "this is a name", "a": "b"})
	action.Run(time.Now(), result)
	if nil != action.(*redisAction).last_error {
		t.Error(action.(*redisAction).last_error)
		return
	}
	var res []string = nil
	select {
	case res = <-ch:
	default:
		t.Error("result is nil")
		return
	}

	excepted := []string{"SET", "sdfs", result.ToJson(), "arg2", "option1", "arg3", "this is a name"}
	if !reflect.DeepEqual(res, excepted) {
		t.Error("excepted is %v", excepted)
		t.Error("actual is %v", res)
	}
}
