package metrics

import (
	"commons"
	"commons/types"
	"ds"
	"net"
	"net/http"
	"testing"
)

var (
	snmp_params = map[string]interface{}{"address": "127.0.0.1",
		"community": "public",
		"port":      161,
		"type":      "snmp_param",
		"version":   "v2c"}
	wbem_params = map[string]interface{}{"url": "tcp://192.168.1.9",
		"user":     "user1",
		"password": "password1",
		"type":     "wbem_param"}
)

func createMockSnmpParams(t *testing.T, client *ds.Client, id, body map[string]interface{}) string {
	return ds.CreateItForTest(t, client, "snmp_param", body)
}

func urlGet(t *testing.T, managed_type, managed_id, target string) commons.Result {
	self := &commons.HttpClient{Url: "http://127.0.0.1" + *address}
	url := self.CreateUrl().Concat("metrics", managed_type, managed_id, target)
	t.Log(url)
	return self.Invoke("GET", url.ToUrl(), nil, 200)
}

func srvTest(t *testing.T, file string, cb func(client *ds.Client, definitions *types.TableDefinitions)) {
	ds.SrvTest(t, file, func(client *ds.Client, definitions *types.TableDefinitions) {
		is_test = true
		Main()

		listener, e := net.Listen("tcp", *address)
		if nil != e {
			return
		}
		ch := make(chan string)

		go func() {
			defer func() {
				ch <- "exit"
			}()
			ch <- "ok"
			http.Serve(listener, nil)
		}()

		s := <-ch
		if "ok" != s {
			return
		}

		cb(client, definitions)

		if nil != srv_instance {
			srv_instance = nil
		}
		if nil != listener {
			listener.Close()
		}
		<-ch

	})
}

func TestGetBasic(t *testing.T) {
	srvTest(t, "../ds/etc/mj_models.xml", func(client *ds.Client, definitions *types.TableDefinitions) {
		_, e := client.DeleteBy("device", emptyParams)
		if nil != e {
			t.Error(e)
			return
		}

		id := ds.CreateMockDeviceForTest(t, client, "1")
		ds.CreateItByParentForTest(t, client, "device", id, "snmp_param", snmp_params)

		res := urlGet(t, "device", id, "sys")
		if res.HasError() {
			t.Error(res.Error())
			return
		}
	})
}
