package sampling

import (
	"commons"
	"commons/types"
	ds "data_store"
	"strings"
	"testing"
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
)

func createMockSnmpParams(t *testing.T, client *ds.Client, id, body map[string]interface{}) string {
	return ds.CreateItForTest(t, client, "snmp_param", body)
}

func urlGet(t *testing.T, sampling_url, managed_type, managed_id, target string) commons.Result {
	self := &commons.HttpClient{Url: sampling_url}
	url := self.CreateUrl().Concat("metrics", managed_type, managed_id, target).ToUrl()
	t.Log(url)
	//fmt.Println(url)
	return self.Invoke("GET", url, nil, 200)
}

func nativeGet(t *testing.T, sampling_url, ip, target string, params map[string]string) commons.Result {
	self := &commons.HttpClient{Url: sampling_url}
	url := self.CreateUrl().Concat("metrics", ip, target).WithQueries(params, "").ToUrl()
	t.Log(url)
	//fmt.Println(url)
	return self.Invoke("GET", url, nil, 200)
}

func TestGetWithNotFound(t *testing.T) {
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {

		res := urlGet(t, sampling_url, "network_device", "123", "sys.oid")
		if !res.HasError() {
			t.Error("error is nil")
			return
		}

		if !strings.Contains(res.ErrorMessage(), "network_device with id was '123' is not found.") {
			t.Error("excepted contains '", "network_device with id was '123' is not found.", "'")
			t.Error("actual is", res.ErrorMessage())
		}
	})
}

func TestGetWithInvalidId(t *testing.T) {
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {

		res := urlGet(t, sampling_url, "network_device", "a123", "sys.oid")
		if !res.HasError() {
			t.Error("error is nil")
			return
		}

		if !strings.Contains(res.ErrorMessage(), "'id' is not a 'objectId', actual value is 'a123'") {
			t.Error("excepted contains '", "'id' is not a 'objectId', actual value is 'a123'", "'")
			t.Error("actual is", res.ErrorMessage())
		}
	})
}

func TestGetBasic(t *testing.T) {
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, e := client.DeleteBy("network_device", emptyParams)
		if nil != e {
			t.Error(e)
			return
		}

		id := ds.CreateItForTest(t, client, "network_device", mo)
		ds.CreateItByParentForTest(t, client, "network_device", id, "wbem_param", wbem_params)
		ds.CreateItByParentForTest(t, client, "network_device", id, "snmp_param", snmp_params)

		res := urlGet(t, sampling_url, "network_device", id, "sys.oid")
		if res.HasError() {
			t.Error(res.Error())
			return
		}

		if nil == res.InterfaceValue() {
			t.Error("values is nil")
		}
	})
}

func TestGetTableBasic(t *testing.T) {
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, e := client.DeleteBy("network_device", emptyParams)
		if nil != e {
			t.Error(e)
			return
		}

		id := ds.CreateItForTest(t, client, "network_device", mo)
		ds.CreateItByParentForTest(t, client, "network_device", id, "wbem_param", wbem_params)
		ds.CreateItByParentForTest(t, client, "network_device", id, "snmp_param", snmp_params)

		res := urlGet(t, sampling_url, "network_device", id, "sys")
		if res.HasError() {
			t.Error(res.Error())
			return
		}
		if nil == res.InterfaceValue() {
			t.Error("values is nil")
		}
	})
}

func TestNativeGetFailed(t *testing.T) {
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, e := client.DeleteBy("network_device", emptyParams)
		if nil != e {
			t.Error(e)
			return
		}

		res := nativeGet(t, sampling_url, "127.0.0.1", "sys.oid", map[string]string{"snmp.version": "v2c"})
		if !res.HasError() {
			t.Error("errors is nil")
			return
		}
		if "'snmp.read_community' is required." != res.ErrorMessage() {
			t.Error(res.Error())
		}
	})
}

func TestNativeGetFailedWithErrorPort(t *testing.T) {
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, e := client.DeleteBy("network_device", emptyParams)
		if nil != e {
			t.Error(e)
			return
		}

		res := nativeGet(t, sampling_url, "127.0.0.1", "sys.oid", map[string]string{"snmp.version": "v2c", "snmp.read_community": "public", "snmp.port": "3244"})
		if !res.HasError() {
			t.Error("errors is nil")
			return
		}

		if !strings.Contains(res.ErrorMessage(), "127.0.0.1:3244") {
			t.Error(res.Error())
		}
	})
}
func TestNativeGetTableFailed(t *testing.T) {
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, e := client.DeleteBy("network_device", emptyParams)
		if nil != e {
			t.Error(e)
			return
		}

		res := nativeGet(t, sampling_url, "127.0.0.1", "sys", map[string]string{"snmp.version": "v2c"})
		if !res.HasError() {
			t.Error("errors is nil")
			return
		}
		if "'snmp.read_community' is required." != res.ErrorMessage() {
			t.Error(res.Error())
		}
	})
}

func TestNativeGetBasic(t *testing.T) {
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, e := client.DeleteBy("network_device", emptyParams)
		if nil != e {
			t.Error(e)
			return
		}
		res := nativeGet(t, sampling_url, "127.0.0.1", "sys.oid", map[string]string{"snmp.version": "v2c", "snmp.read_community": "public"})
		if res.HasError() {
			t.Error(res.Error())
			return
		}

		if nil == res.InterfaceValue() {
			t.Error("values is nil")
		}
	})
}

func TestNativeGetTableBasic(t *testing.T) {
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, e := client.DeleteBy("network_device", emptyParams)
		if nil != e {
			t.Error(e)
			return
		}

		res := nativeGet(t, sampling_url, "127.0.0.1", "sys", map[string]string{"snmp.version": "v2c", "snmp.read_community": "public"})
		if res.HasError() {
			t.Error(res.Error())
			return
		}
		if nil == res.InterfaceValue() {
			t.Error("values is nil")
		}
	})
}
