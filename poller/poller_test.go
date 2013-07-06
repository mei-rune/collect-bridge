package poller

import (
	"commons/types"
	"ds"
	"errors"
	"github.com/garyburd/redigo/redis"
	"os"
	"sampling"
	"strings"
	"testing"
	"time"
)

var (
	mo = map[string]interface{}{"name": "dd",
		"type":            "network_device",
		"address":         "127.0.0.1",
		"device_type":     1,
		"services":        2,
		"managed_address": "20.0.8.110"}

	snmp_params = map[string]interface{}{"address": "127.0.0.1",
		"read_community": "public",
		"port":           161,
		"type":           "snmp_param",
		"version":        "v2c"}

	wbem_params = map[string]interface{}{"url": "tcp://192.168.1.9",
		"user":     "user1",
		"password": "password1",
		"type":     "wbem_param"}

	metric_trigger2 = map[string]interface{}{
		"name":       "this is a test trigger",
		"type":       "metric_trigger",
		"metric":     "sys",
		"expression": "@every 1ms"}

	redis_commands2 = map[string]interface{}{
		"type":    "redis_command",
		"name":    "this is a test redis action",
		"command": "SET",
		"arg0":    "abc",
		"arg1":    "$name"}
)

func srvTest(t *testing.T, cb func(client *ds.Client, definitions *types.TableDefinitions)) {
	sampling.SrvTest(t, "../ds/etc/tpt_models.xml", func(client *ds.Client, definitions *types.TableDefinitions) {
		cb(client, definitions)
	})
}

func SetResultToRedis(c redis.Conn, key, value string) error {
	reply, err := c.Do("SET", key, value)
	if nil != err {
		return err
	}
	s, err := redis.String(reply, err)
	if nil != err {
		return err
	}
	if "OK" == s {
		return nil
	}
	return errors.New("result is not equals ok, " + s)
}

func getResultFromRedis(c redis.Conn, key string) (string, error) {
	reply, err := c.Do("GET", key)
	if nil != err {
		return "", err
	}
	s, e := redis.String(reply, err)
	if e == redis.ErrNil {
		return s, nil
	}
	return s, e
}

func TestIntegratedTestPoller(t *testing.T) {
	srvTest(t, func(client *ds.Client, definitions *types.TableDefinitions) {
		t.Log("Please run redis at " + redis_address + " before run unit test.")
		id := ds.CreateItForTest(t, client, "network_device", mo)
		ds.CreateItByParentForTest(t, client, "network_device", id, "wbem_param", wbem_params)
		ds.CreateItByParentForTest(t, client, "network_device", id, "snmp_param", snmp_params)
		mt_id := ds.CreateItByParentForTest(t, client, "network_device", id, "metric_trigger", metric_trigger2)
		ds.CreateItByParentForTest(t, client, "metric_trigger", mt_id, "redis_command", redis_commands2)

		c, err := redis.DialTimeout("tcp", redis_address, 0, 1*time.Second, 1*time.Second)
		if err != nil {
			t.Errorf("[redis] connect to '%s' failed, %v", redis_address, err)
			return
		}
		err = SetResultToRedis(c, "abc", "")
		if nil != err {
			t.Error(err)
			return
		}

		is_test = true
		Runforever()

		if nil == server_test || nil == server_test.jobs || 0 == len(server_test.jobs) {
			t.Error("load trigger failed.")
			return
		}

		hostName, e := os.Hostname()
		if nil != e {
			t.Error(e)
			return
		}

		for i := 0; i < 100; i++ {
			s, e := getResultFromRedis(c, "abc")
			if nil != e {
				t.Error(e)
				return
			}

			if 0 != len(s) {
				if strings.ToLower(hostName) != strings.ToLower(s) {
					t.Error("it is not equals excepted value")
					t.Error("excepted is", hostName)
					t.Error("actual is", s)
				}
				return
			}
			time.Sleep(1 * time.Second)
		}

		t.Error("not wait")

	})
}
