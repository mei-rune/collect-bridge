package mdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
)

func atoi(s string) int {
	i, e := strconv.Atoi(s)
	if nil != e {
		panic(e.Error())
	}
	return i
}

func readAll(r io.Reader) string {
	bs, e := ioutil.ReadAll(r)
	if nil != e {
		panic(e.Error())
	}
	return string(bs)
}

func createMockDevice(t *testing.T, factor string) string {
	id, e := createJson("device", fmt.Sprintf(`{"name":"dd%s", "catalog":%s, "services":2%s}`, factor, factor, factor))
	if nil != e {
		t.Error(e.Error())
		return ""
	}
	return id
}

func updateMockDevice(t *testing.T, id, factor string) {
	e := updateJson("device", id, fmt.Sprintf(`{"name":"dd%s", "catalog":%s, "services":2%s}`, factor, factor, factor))
	if nil != e {
		t.Error(e.Error())
	}
}

func getDeviceById(t *testing.T, id string) map[string]interface{} {
	res, e := findById("device", id)
	if nil != e {
		t.Error(e.Error())
		return nil
	}
	return res
}

func getDeviceByName(t *testing.T, factor string) map[string]interface{} {
	res, e := findBy("device", map[string]string{"name": "dd" + factor})
	if nil != e {
		t.Error(e.Error())
		return nil
	}
	if 1 > len(res) {
		return nil
	}
	return res[0]
}

func searchBy(res []map[string]interface{}, f func(r map[string]interface{}) bool) map[string]interface{} {
	for _, r := range res {
		if f(r) {
			return r
		}
	}
	return nil
}

func fetchInt(params map[string]interface{}, key string) int {

	v := params[key]
	if nil == v {
		panic(fmt.Sprintf("value of '"+key+"' is nil in %v", params))
	}
	return int(v.(float64))
}
func validMockDevice(t *testing.T, factor string, drv map[string]interface{}) {
	if "dd"+factor != drv["name"].(string) {
		t.Errorf("excepted name is 'dd%s', actual name is '%v'", factor, drv["name"])
		return
	}
	if atoi(factor) != fetchInt(drv, "catalog") {
		t.Errorf("excepted catalog is '%s', actual catalog is '%v'", factor, drv["catalog"])
		return
	}
	if atoi("2"+factor) != fetchInt(drv, "services") {
		t.Errorf("excepted services is '2%s', actual services is '%v'", factor, drv["services"])
		return
	}
}
func create(t string, body map[string]interface{}) (string, error) {
	msg, e := json.Marshal(body)
	if nil != e {
		return "", fmt.Errorf("create %s failed, %v", t, e)
	}
	return createJson(t, string(msg))
}

func createJson(t, msg string) (string, error) {
	resp, e := http.Post("http://127.0.0.1:7071/mdb/"+t, "application/json", bytes.NewBuffer([]byte(msg)))
	if nil != e {
		return "", fmt.Errorf("create %s failed, %v", t, e)
	}
	if resp.StatusCode != 201 {
		return "", fmt.Errorf("create %s failed, %v, %v", t, resp.StatusCode, readAll(resp.Body))
	}
	result := map[string]interface{}{}
	e = json.Unmarshal([]byte(readAll(resp.Body)), &result)

	if nil != e {
		return "", fmt.Errorf("create %s failed, %v", t, e)
	}
	return result["value"].(string), nil
}

func update(t, id string, body map[string]interface{}) error {
	msg, e := json.Marshal(body)
	if nil != e {
		return fmt.Errorf("update %s failed, %v", t, e)
	}
	return updateJson(t, id, string(msg))
}

func HttpPut(endpoint string, bodyType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("PUT", endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", bodyType)
	return http.DefaultClient.Do(req)
}
func httpDelete(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func updateJson(t, id, msg string) error {
	resp, e := HttpPut("http://127.0.0.1:7071/mdb/"+t+"/"+id, "application/json", bytes.NewBuffer([]byte(msg)))
	if nil != e {
		return fmt.Errorf("update %s failed, %v", t, e)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("update %s failed, %v, %v", t, resp.StatusCode, readAll(resp.Body))
	}
	return nil
}

func findById(t, id string) (map[string]interface{}, error) {
	resp, e := http.Get("http://127.0.0.1:7071/mdb/" + t + "/" + id)
	if nil != e {
		return nil, fmt.Errorf("find %s failed, %v", t, e)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("find %s failed, %v, %v", t, resp.StatusCode, readAll(resp.Body))
	}
	result := map[string]interface{}{}
	e = json.Unmarshal([]byte(readAll(resp.Body)), &result)

	if nil != e {
		return nil, fmt.Errorf("find %s failed, %v", t, e)
	}
	return result["value"].(map[string]interface{}), nil
}
func deleteById(t, id string) error {
	resp, e := httpDelete("http://127.0.0.1:7071/mdb/" + t + "/" + id)
	if nil != e {
		return fmt.Errorf("find %s failed, %v", t, e)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("find %s failed, %v, %v", t, resp.StatusCode, readAll(resp.Body))
	}
	return nil
}

func findBy(t string, params map[string]string) ([]map[string]interface{}, error) {
	url := "http://127.0.0.1:7071/mdb/" + t + "/query?"
	for k, v := range params {
		url += ("@" + k + "=" + v + "&")
	}
	resp, e := http.Get(url[:len(url)-1])
	if nil != e {
		return nil, fmt.Errorf("find %s failed, %v", t, e)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("find %s failed, %v, %v", t, resp.StatusCode, readAll(resp.Body))
	}
	result := map[string]interface{}{}
	e = json.Unmarshal([]byte(readAll(resp.Body)), &result)

	if nil != e {
		return nil, fmt.Errorf("find %s failed, %v", t, e)
	}
	res := result["value"].([]interface{})
	results := make([]map[string]interface{}, 0, 3)
	for _, r := range res {
		results = append(results, r.(map[string]interface{}))
	}
	return results, nil
}

func TestDeviceCreateAndUpdate(t *testing.T) {
	deleteById("device", "all")
	id1 := createMockDevice(t, "1")
	id2 := createMockDevice(t, "2")
	id3 := createMockDevice(t, "3")
	id4 := createMockDevice(t, "4")
	if "" == id1 {
		return
	}

	d1 := getDeviceById(t, id1)
	d2 := getDeviceById(t, id2)
	d3 := getDeviceById(t, id3)
	d4 := getDeviceById(t, id4)

	if nil == d1 {
		return
	}

	validMockDevice(t, "1", d1)
	validMockDevice(t, "2", d2)
	validMockDevice(t, "3", d3)
	validMockDevice(t, "4", d4)

	updateMockDevice(t, id1, "11")
	updateMockDevice(t, id2, "21")
	updateMockDevice(t, id3, "31")
	updateMockDevice(t, id4, "41")

	d1 = getDeviceById(t, id1)
	d2 = getDeviceById(t, id2)
	d3 = getDeviceById(t, id3)
	d4 = getDeviceById(t, id4)

	if nil == d1 {
		return
	}

	validMockDevice(t, "11", d1)
	validMockDevice(t, "21", d2)
	validMockDevice(t, "31", d3)
	validMockDevice(t, "41", d4)

	d1 = getDeviceByName(t, "11")
	d2 = getDeviceByName(t, "21")
	d3 = getDeviceByName(t, "31")
	d4 = getDeviceByName(t, "41")

	if nil == d1 {
		return
	}

	validMockDevice(t, "11", d1)
	validMockDevice(t, "21", d2)
	validMockDevice(t, "31", d3)
	validMockDevice(t, "41", d4)

	deleteById("device", "all")

	d1 = getDeviceByName(t, "11")
	d2 = getDeviceByName(t, "21")
	d3 = getDeviceByName(t, "31")
	d4 = getDeviceByName(t, "41")

	if nil != d1 || nil != d2 || nil != d3 || nil != d4 {
		t.Errorf("remove all failed")
	}
}

func TestDeviceFindBy(t *testing.T) {
	deleteById("device", "all")
	id1 := createMockDevice(t, "1")
	id2 := createMockDevice(t, "2")
	id3 := createMockDevice(t, "3")
	id4 := createMockDevice(t, "4")
	if "" == id1 {
		return
	}
	if "" == id2 {
		return
	}
	if "" == id3 {
		return
	}
	if "" == id4 {
		return
	}

	res, e := findBy("device", map[string]string{"catalog": "[eq]1"})
	if nil != e {
		t.Errorf(e.Error())
		return
	}
	validMockDevice(t, "1", res[0])

	res, e = findBy("device", map[string]string{"catalog": "[lte]1"})
	if nil != e {
		t.Errorf(e.Error())
		return
	}
	validMockDevice(t, "1", res[0])

	res, e = findBy("device", map[string]string{"catalog": "[lte]2"})
	if nil != e {
		t.Errorf(e.Error())
		return
	}
	d1 := searchBy(res, func(r map[string]interface{}) bool { return r["name"] == "dd1" })
	d2 := searchBy(res, func(r map[string]interface{}) bool { return r["name"] == "dd2" })
	if nil == d1 {
		t.Errorf("catalog <=2 failed, result is %v", res)
		return
	}
	validMockDevice(t, "1", d1)
	validMockDevice(t, "2", d2)

	res, e = findBy("device", map[string]string{"catalog": "[lt]2"})
	if nil != e {
		t.Errorf(e.Error())
		return
	}
	validMockDevice(t, "1", res[0])

	res, e = findBy("device", map[string]string{"catalog": "[gt]3"})
	if nil != e {
		t.Errorf(e.Error())
		return
	}
	validMockDevice(t, "4", res[0])
}
