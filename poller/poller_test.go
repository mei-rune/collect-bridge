package poller

import (
	"bytes"
	"carrier"
	"commons"
	"commons/types"
	ds "data_store"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sampling"
	"strconv"
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

	metric_trigger_for_cpu = map[string]interface{}{
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
	sampling.SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, url string, definitions *types.TableDefinitions) {
		*dsUrl = client.Url
		*sampling_url = url
		is_test = true
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

func TestIntegratedPoller(t *testing.T) {
	srvTest(t, func(client *ds.Client, definitions *types.TableDefinitions) {
		t.Log("Please run redis at " + *redisAddress + " before run unit test.")
		id := ds.CreateItForTest(t, client, "network_device", mo)
		ds.CreateItByParentForTest(t, client, "network_device", id, "wbem_param", wbem_params)
		ds.CreateItByParentForTest(t, client, "network_device", id, "snmp_param", snmp_params)
		mt_id := ds.CreateItByParentForTest(t, client, "network_device", id, "metric_trigger", metric_trigger2)
		ds.CreateItByParentForTest(t, client, "metric_trigger", mt_id, "redis_command", redis_commands2)

		c, err := redis.DialTimeout("tcp", *redisAddress, 0, 1*time.Second, 1*time.Second)
		if err != nil {
			t.Errorf("[redis] connect to '%s' failed, %v", *redisAddress, err)
			return
		}
		err = SetResultToRedis(c, "abc", "")
		if nil != err {
			t.Error(err)
			return
		}

		*load_cookies = false
		defer func() { *load_cookies = true }()
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

func TestIntegratedAlert(t *testing.T) {
	srvTest(t, func(client *ds.Client, definitions *types.TableDefinitions) {
		id := ds.CreateItForTest(t, client, "network_device", mo)
		ds.CreateItByParentForTest(t, client, "network_device", id, "wbem_param", wbem_params)
		ds.CreateItByParentForTest(t, client, "network_device", id, "snmp_param", snmp_params)
		mt_id := ds.CreateItByParentForTest(t, client, "network_device", id, "metric_trigger", metric_trigger_for_cpu)
		ds.CreateItByParentForTest(t, client, "metric_trigger", mt_id, "alert", map[string]interface{}{
			"id":               "123",
			"name":             "this is a test alert",
			"delay_times":      0,
			"expression_style": "json",
			"expression_code": map[string]interface{}{
				"attribute": "a",
				"operator":  ">=",
				"value":     "0"}})

		count := 0
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count++
		}))
		defer ts.Close()

		is_test = true
		*foreignUrl = ts.URL
		*load_cookies = false
		defer func() { *load_cookies = true }()
		Runforever()

		if nil == server_test || nil == server_test.jobs || 0 == len(server_test.jobs) {
			t.Error("load trigger failed.")
			return
		}

		tr_instance := server_test.jobs[mt_id].(*metricJob)

		for i := 0; i < 100; i++ {
			if nil != tr_instance.last_error {
				t.Error(tr_instance.last_error)
				return
			}

			tr_instance.l.Lock()
			e := tr_instance.actions[0].last_error
			tr_instance.l.Unlock()

			if nil != e {
				if !strings.Contains(e.Error(), "not found") {
					t.Error(e)
				}
				return
			}

			time.Sleep(1 * time.Second)
		}

		t.Error("not wait")
	})
}

func TestIntegratedAlert2(t *testing.T) {
	srvTest(t, func(client *ds.Client, definitions *types.TableDefinitions) {
		id := ds.CreateItForTest(t, client, "network_device", mo)
		ds.CreateItByParentForTest(t, client, "network_device", id, "wbem_param", wbem_params)
		ds.CreateItByParentForTest(t, client, "network_device", id, "snmp_param", snmp_params)
		mt_id := ds.CreateItByParentForTest(t, client, "network_device", id, "metric_trigger", metric_trigger_for_cpu)
		ds.CreateItByParentForTest(t, client, "metric_trigger", mt_id, "alert", map[string]interface{}{
			"id":               "123",
			"name":             "this is a test alert",
			"delay_times":      0,
			"expression_style": "json",
			"expression_code": map[string]interface{}{
				"attribute": "services",
				"operator":  ">=",
				"value":     "0"}})

		count := 0
		var js string
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bs, _ := ioutil.ReadAll(r.Body)
			js = string(bs)
			count++
		}))
		defer ts.Close()

		is_test = true
		*foreignUrl = ts.URL
		*load_cookies = false
		defer func() { *load_cookies = true }()
		Runforever()

		if nil == server_test || nil == server_test.jobs || 0 == len(server_test.jobs) {
			t.Error("load trigger failed.")
			return
		}

		tr_instance := server_test.jobs[mt_id].(*metricJob)

		for i := 0; i < 100; i++ {
			if nil != tr_instance.last_error {
				t.Error(tr_instance.last_error)
				return
			}

			tr_instance.l.Lock()
			e := tr_instance.actions[0].last_error
			tr_instance.l.Unlock()

			if nil != e {
				if !strings.Contains(e.Error(), "not founc") {
					t.Error(e)
				}
				return
			}

			if 0 != len(js) {
				var data []map[string]interface{}

				decoder := json.NewDecoder(bytes.NewBufferString(js))
				decoder.UseNumber()
				e := decoder.Decode(&data)
				if nil != e {
					t.Error(e)
					return
				}
				if 1 != len(data) {
					t.Error("it is not equals excepted value")
					t.Error("value is", js)
				}
				status, _ := data[0]["status"].(json.Number).Int64()
				if 1 != status {
					t.Error("it is not equals excepted value")
					t.Error("value is", js)
				}
				return
			}
			time.Sleep(1 * time.Second)
		}

		t.Error("not wait")
	})
}

func TestIntegratedHistory(t *testing.T) {
	srvTest(t, func(client *ds.Client, definitions *types.TableDefinitions) {
		id := ds.CreateItForTest(t, client, "network_device", mo)
		ds.CreateItByParentForTest(t, client, "network_device", id, "wbem_param", wbem_params)
		ds.CreateItByParentForTest(t, client, "network_device", id, "snmp_param", snmp_params)
		mt_id := ds.CreateItByParentForTest(t, client, "network_device", id, "metric_trigger", metric_trigger2)
		ds.CreateItByParentForTest(t, client, "metric_trigger", mt_id, "history", map[string]interface{}{
			"id":        "123",
			"name":      "this is a test alert",
			"attribute": "services"})

		// hostName, e := os.Hostname()
		// if nil != e {
		// 	t.Error(e)
		// 	return
		// }

		count := 0
		var js string
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bs, _ := ioutil.ReadAll(r.Body)
			js = string(bs)
			count++
		}))
		defer ts.Close()

		is_test = true
		*foreignUrl = ts.URL
		*load_cookies = false
		defer func() { *load_cookies = true }()
		Runforever()

		if nil == server_test || nil == server_test.jobs || 0 == len(server_test.jobs) {
			t.Error("load trigger failed.")
			return
		}

		tr_instance := server_test.jobs[mt_id].(*metricJob)

		for i := 0; i < 100; i++ {
			if nil != tr_instance.last_error {
				t.Error(tr_instance.last_error)
				return
			}

			tr_instance.l.Lock()
			e := tr_instance.actions[0].last_error
			tr_instance.l.Unlock()

			if nil != e {
				if !strings.Contains(e.Error(), "not founc") {
					t.Error(e)
				}
				return
			}

			if 0 != len(js) {
				if !strings.Contains(strings.ToLower(js), `"value":76`) {
					t.Error("excepted contains", `"value":76"`)
					t.Error("actual is ", js)
				}
				return
			}
			time.Sleep(1 * time.Second)
		}

		t.Error("not wait")
	})
}

func TestIntegratedAlertWithCarrier(t *testing.T) {
	srvTest(t, func(client *ds.Client, definitions *types.TableDefinitions) {
		mo_id := ds.CreateItForTest(t, client, "network_device", mo)
		ds.CreateItByParentForTest(t, client, "network_device", mo_id, "wbem_param", wbem_params)
		ds.CreateItByParentForTest(t, client, "network_device", mo_id, "snmp_param", snmp_params)
		mt_id := ds.CreateItByParentForTest(t, client, "network_device", mo_id, "metric_trigger", metric_trigger_for_cpu)
		rule_id := ds.CreateItByParentForTest(t, client, "metric_trigger", mt_id, "alert", map[string]interface{}{
			"name":             "this is a test alert",
			"delay_times":      0,
			"expression_style": "json",
			"expression_code": map[string]interface{}{
				"attribute": "services",
				"operator":  ">=",
				"value":     "0"}})

		carrier.SrvTest(t, func(db *sql.DB, url string) {
			is_test = true
			*foreignUrl = url
			Runforever()

			if nil == server_test || nil == server_test.jobs || 0 == len(server_test.jobs) {
				t.Error("load trigger failed.")
				return
			}

			tr_instance := server_test.jobs[mt_id].(*metricJob)

			for i := 0; i < 100; i++ {
				if nil != tr_instance.last_error {
					t.Error(tr_instance.last_error)
					return
				}

				tr_instance.l.Lock()
				e := tr_instance.actions[0].last_error
				tr_instance.l.Unlock()

				if nil != e {
					if !strings.Contains(e.Error(), "not founc") {
						t.Error(e)
					}
					return
				}

				entities, e := carrier.SelectAlertCookies(db)
				if nil != e {
					t.Error(e)
					return
				}

				if nil != entities && 0 != len(entities) {
					entity := entities[0]
					rule_id_int, _ := strconv.ParseInt(rule_id, 10, 64)
					mo_id_int, _ := strconv.ParseInt(mo_id, 10, 64)

					if entity.ActionId != rule_id_int {
						t.Error(" entity.ActionId != rule_id, excepted is ", rule_id_int, ", actual is ", entity.ActionId)
					}
					if entity.Status != 1 {
						t.Error(" entity.Status != 1, excepted is ", 1, ", actual is ", entity.Status)
					}
					if entity.ManagedType != "managed_object" {
						t.Error(" entity.ManagedType != mo_type, excepted is managed_object, actual is ", entity.ManagedType)
					}
					if entity.ManagedId != mo_id_int {
						t.Error(" entity.ManagedId != mo_id, excepted is ", mo_id_int, ", actual is ", entity.ManagedId)
					}
					return
				}

				time.Sleep(1 * time.Second)
			}

			t.Error("not wait")
		})
	})
}

func getMetric(parentId, metric string) commons.Result {
	client_url := ""
	// if is_test {
	// 	client_url = commons.NewUrlBuilder(*sampling_url).Concat("metrics", "managed_object", fmt.Sprint(parentId), metric).ToUrl()
	// } else {
	client_url = commons.NewUrlBuilder(*sampling_url).Concat("managed_object", fmt.Sprint(parentId), metric).ToUrl()
	//}
	metricClient := &commons.HttpClient{Url: *sampling_url}
	return metricClient.InvokeWith("GET", client_url, nil, 200)
}

func TestIntegratedHistoryWithCarrier(t *testing.T) {
	srvTest(t, func(client *ds.Client, definitions *types.TableDefinitions) {
		mo_id := ds.CreateItForTest(t, client, "network_device", mo)
		ds.CreateItByParentForTest(t, client, "network_device", mo_id, "wbem_param", wbem_params)
		ds.CreateItByParentForTest(t, client, "network_device", mo_id, "snmp_param", snmp_params)
		mt_id := ds.CreateItByParentForTest(t, client, "network_device", mo_id, "metric_trigger", metric_trigger2)
		rule_id := ds.CreateItByParentForTest(t, client, "metric_trigger", mt_id, "history", map[string]interface{}{
			"name":      "this is a test history",
			"attribute": "services"})

		res := getMetric(mo_id, "sys")
		if res.HasError() {
			t.Error("read metric failed,", res.ErrorMessage())
			return
		}
		sys, e := res.Value().AsObject()
		if nil != e {
			t.Error(e)
			return
		}

		value, _ := commons.AsInt(sys["services"])

		carrier.SrvTest(t, func(db *sql.DB, url string) {
			is_test = true
			*foreignUrl = url
			Runforever()

			if nil == server_test || nil == server_test.jobs || 0 == len(server_test.jobs) {
				t.Error("load trigger failed.")
				return
			}

			tr_instance := server_test.jobs[mt_id].(*metricJob)

			for i := 0; i < 100; i++ {
				if nil != tr_instance.last_error {
					t.Error(tr_instance.last_error)
					return
				}

				tr_instance.l.Lock()
				e := tr_instance.actions[0].last_error
				tr_instance.l.Unlock()

				if nil != e {
					if !strings.Contains(e.Error(), "not founc") {
						t.Error(e)
					}
					return
				}

				entities, e := carrier.SelectHistories(db)
				if nil != e {
					t.Error(e)
					return
				}

				if nil != entities && 0 != len(entities) {
					entity := entities[0]
					rule_id_int, _ := strconv.ParseInt(rule_id, 10, 64)
					mo_id_int, _ := strconv.ParseInt(mo_id, 10, 64)

					if entity.ActionId != rule_id_int {
						t.Error(" entity.ActionId != rule_id, excepted is ", rule_id_int, ", actual is ", entity.ActionId)
					}

					if int(entity.CurrentValue) != value {
						t.Error(" entity.CurrentValue != value, excepted is ", value, ", actual is ", entity.CurrentValue)
					}

					if entity.ManagedType != "managed_object" {
						t.Error(" entity.ManagedType != mo_type, excepted is managed_object, actual is ", entity.ManagedType)
					}
					if entity.ManagedId != mo_id_int {
						t.Error(" entity.ManagedId != mo_id, excepted is ", mo_id_int, ", actual is ", entity.ManagedId)
					}
					return
				}
				time.Sleep(1 * time.Second)
			}

			t.Error("not wait")
		})
	})
}
