package poller

import (
	"commons"
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"reflect"
	"testing"
	"time"
)

func checkResult(t *testing.T, c redis.Conn, cmd, key, excepted string) {
	reply, err := c.Do(cmd, key)
	s, err := redis.String(reply, err)
	if nil != err {
		t.Errorf("GET %s failed, %v", key, err)
	} else if excepted != s {
		t.Errorf("check %s failed, actual is %v, excepted is %v", key, reply, excepted)
	}
}

func clearRedis(t *testing.T, c redis.Conn, key string) {
	reply, err := c.Do("DEL", key)
	_, err = redis.Int(reply, err)
	if nil != err {
		t.Errorf("DEL %s failed, %v", key, err)
	}
}

func redisTest(t *testing.T, cb func(client chan []string, c redis.Conn)) {
	redis_client, err := newRedis(*redisAddress)
	if nil != err {
		t.Error(err)
		return
	}

	c, err := redis.DialTimeout("tcp", *redisAddress, 0, 1*time.Second, 1*time.Second)
	if err != nil {
		t.Errorf("[redis] connect to '%s' failed, %v", *redisAddress, err)
		return
	}
	defer c.Close()

	for i := 0; i < 10; i++ {
		clearRedis(t, c, fmt.Sprintf("a%v", i))
	}

	cb(redis_client, c)
}
func TestRedis(t *testing.T) {
	redisTest(t, func(redis_channel chan []string, c redis.Conn) {
		redis_channel <- []string{"SET", "a1", "1223"}
		redis_channel <- []string{"SET", "a2", "1224"}
		redis_channel <- []string{"SET", "a3", "1225"}
		redis_channel <- []string{"SET", "a4", "1226"}
		redis_channel <- []string{"SET", "a5", "1227"}

		time.Sleep(2 * time.Second)

		checkResult(t, c, "GET", "a1", "1223")
		checkResult(t, c, "GET", "a2", "1224")
		checkResult(t, c, "GET", "a3", "1225")
		checkResult(t, c, "GET", "a4", "1226")
		checkResult(t, c, "GET", "a5", "1227")
	})
}

func TestRedisEmpty(t *testing.T) {
	redisTest(t, func(redis_channel chan []string, c redis.Conn) {
		redis_channel <- []string{}
		redis_channel <- []string{"SET", "a1", "1223"}
		redis_channel <- []string{"SET", "a2", "1224"}
		redis_channel <- []string{"SET", "a3", "1225"}
		redis_channel <- []string{"SET", "a4", "1226"}
		redis_channel <- []string{"SET", "a5", "1227"}

		time.Sleep(2 * time.Second)

		checkResult(t, c, "GET", "a1", "1223")
		checkResult(t, c, "GET", "a2", "1224")
		checkResult(t, c, "GET", "a3", "1225")
		checkResult(t, c, "GET", "a4", "1226")
		checkResult(t, c, "GET", "a5", "1227")
	})
}

func TestRedisAction(t *testing.T) {
	ch := make(chan []string, 1)

	action, e := newRedisAction(map[string]interface{}{
		"id":      12,
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
	e = action.Run(time.Now(), result)
	if nil != e {
		t.Error(e)
		return
	}
	var res []string = nil
	select {
	case res = <-ch:
	default:
		t.Error("result is nil")
		return
	}

	m := result.ToMap()
	m["op"] = "option1"
	// m["managed_type"] = "managed_object"
	// m["metric"] = "cpu"
	js, e := json.Marshal(m)
	if nil != e {
		t.Error(e)
		return
	}

	excepted := []string{"SET", "sdfs", string(js), "arg2", "option1", "arg3", "this is a name"}
	if !reflect.DeepEqual(res, excepted) {
		t.Error("excepted is ", excepted)
		t.Error("actual is ", res)
	}
}
