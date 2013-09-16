package sampling

import (
	"bytes"
	"commons"
	"commons/types"
	ds "data_store"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
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

	old_link = map[string]interface{}{"name": "dd",
		"custom_speed_up":   12,
		"custom_speed_down": 12,
		"description":       "",
		"from_device":       0,
		"from_if_index":     1,
		"to_device":         0,
		"to_if_index":       1,
		"link_type":         12,
		"forward":           true,
		"from_based":        true}

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

func copyFrom(from, addition map[string]interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range from {
		res[k] = v
	}

	if nil != addition {
		for k, v := range addition {
			res[k] = v
		}
	}
	return res
}

func createMockSnmpParams(t *testing.T, client *ds.Client, id, body map[string]interface{}) string {
	return ds.CreateItForTest(t, client, "snmp_param", body)
}

func urlGet(t *testing.T, sampling_url, managed_type, managed_id, target string) commons.Result {
	self := &commons.HttpClient{Url: sampling_url}
	url := self.CreateUrl().Concat(managed_type, managed_id, target).ToUrl()
	//t.Log(url)
	return self.InvokeWithObject("GET", url, nil, 200)
}

func nativeGet(t *testing.T, sampling_url, ip, target string, params map[string]string) commons.Result {
	self := &commons.HttpClient{Url: sampling_url}
	url := self.CreateUrl().Concat(ip, target).WithQueries(params, "").ToUrl()
	//t.Log(url)
	return self.InvokeWithObject("GET", url, nil, 200)
}

func batchGet(t *testing.T, url string, requests []*ExchangeRequest) ([]*ExchangeResponse, error) {
	buffer := bytes.NewBuffer(make([]byte, 0, 1000))
	e := json.NewEncoder(buffer).Encode(requests)
	if nil != e {
		return nil, commons.NewApplicationError(http.StatusBadRequest, e.Error())
	}
	req, err := http.NewRequest("POST", url, buffer)
	if err != nil {
		return nil, commons.NewApplicationError(http.StatusBadRequest, err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Connection", "Keep-Alive")
	resp, e := http.DefaultClient.Do(req)
	if nil != e {
		return nil, errors.New("get failed, " + e.Error())
	}

	defer func() {
		if nil != resp.Body {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusAccepted {
		if http.StatusNoContent == resp.StatusCode {
			return nil, nil
		}
		resp_body, e := ioutil.ReadAll(resp.Body)
		if nil != e {
			return nil, e
		}
		if nil == resp_body || 0 == len(resp_body) {
			return nil, commons.NewApplicationError(resp.StatusCode, fmt.Sprintf("%v: error", resp.StatusCode))
		}

		return nil, commons.NewApplicationError(resp.StatusCode, string(resp_body))
	}

	resp_body, e := ioutil.ReadAll(resp.Body)
	if nil != e {
		return nil, e
	}
	if nil == resp_body || 0 == len(resp_body) {
		return nil, commons.NewApplicationError(resp.StatusCode, fmt.Sprintf("%v: error", resp.StatusCode))
	}

	var result []*ExchangeResponse
	decoder := json.NewDecoder(bytes.NewBuffer(resp_body))
	decoder.UseNumber()
	e = decoder.Decode(&result)
	if nil != e {

		fmt.Println(e, ",", string(resp_body))
		return nil, e
	}
	return result, nil
}

func nativePut(t *testing.T, sampling_url, ip, target string, params map[string]string, body interface{}) commons.Result {
	self := &commons.HttpClient{Url: sampling_url}
	url := self.CreateUrl().Concat(ip, target).WithQueries(params, "").ToUrl()
	//t.Log(url)
	return self.InvokeWithObject("PUT", url, body, 200)
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

func TestBatchGetTableBasic(t *testing.T) {
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, err := client.DeleteBy("network_device", emptyParams)
		if nil != err {
			t.Error(err)
			return
		}

		res, e := batchGet(t, sampling_url+"/batch", []*ExchangeRequest{&ExchangeRequest{Address: "127.0.0.1", Action: "GET", Name: "sys", Params: map[string]string{"snmp.version": "v2c", "snmp.read_community": "public"}}})
		if nil != e {
			t.Error(e)
			return
		}
		for i := 0; i < 1000; i++ {
			res, e = batchGet(t, sampling_url+"/batch", []*ExchangeRequest{})
			if nil != e {
				t.Error(e)
				return
			}
			if nil != res && 0 != len(res) {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}

		if nil == res || 0 == len(res) {
			t.Error("not result")
			return
		}

		if nil == res[0].Evalue {
			t.Error("values is nil")
		}

		t.Log(res[0])

		res, e = batchGet(t, sampling_url+"/batch", []*ExchangeRequest{})
		if nil != e {
			t.Error(e)
			return
		}
		if nil != res && 0 != len(res) {
			t.Error("repeated result")
		}
	})
}

func TestBatchGetTableTwo(t *testing.T) {
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, err := client.DeleteBy("network_device", emptyParams)
		if nil != err {
			t.Error(err)
			return
		}

		res, e := batchGet(t, sampling_url+"/batch", []*ExchangeRequest{&ExchangeRequest{Address: "127.0.0.1", Action: "GET", Name: "sys", Params: map[string]string{"snmp.version": "v2c", "snmp.read_community": "public"}},
			&ExchangeRequest{Address: "127.0.0.1", Action: "GET", Name: "sys.name", Params: map[string]string{"snmp.version": "v2c", "snmp.read_community": "public"}}})
		if nil != e {
			t.Error(e)
			return
		}
		var all []*ExchangeResponse
		for i := 0; i < 1000; i++ {
			res, e = batchGet(t, sampling_url+"/batch", []*ExchangeRequest{})
			if nil != e {
				t.Error(e)
				return
			}
			if nil != res && 0 != len(res) {
				all = append(all, res...)
			}

			if 2 == len(all) {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}

		if 2 != len(all) {
			t.Error("not result")
			return
		}

		if nil == all[0].Evalue {
			t.Error("values is nil")
		}

		t.Log(all[0])
		t.Log(all[1])

		res, e = batchGet(t, sampling_url+"/batch", []*ExchangeRequest{})
		if nil != e {
			t.Error(e)
			return
		}
		if nil != res && 0 != len(res) {
			t.Error("repeated result")
		}
	})
}
