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

func createMockHistoryRule2(t *testing.T, factor string) string {
	id, e := createJson("history_rule", fmt.Sprintf(`{"name":"%s", "expression":"d%s", "metric":"2%s"}`, factor, factor, factor))
	if nil != e {
		t.Error(e.Error())
		return ""
	}
	return id
}
func createMockHistoryRule(t *testing.T, id, factor string) string {
	id, e := createJson("history_rule", fmt.Sprintf(`{"name":"%s", "expression":"d%s", "metric":"2%s", "parent_type":"devices", "parent_id":"%s"}`, factor, factor, factor, id))
	if nil != e {
		t.Error(e.Error())
		return ""
	}
	return id
}

func createMockInterface(t *testing.T, id, factor string) string {
	id, e := createJson("interface", fmt.Sprintf(`{"ifIndex":%s, "ifDescr":"d%s", "ifType":2%s, "ifMtu":3%s, "ifSpeed":4%s, "device_id":"%s"}`, factor, factor, factor, factor, factor, id))
	if nil != e {
		t.Error(e.Error())
		return ""
	}
	return id
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

func DeviceNotExistsById(t *testing.T, id string) {
	_, e := findById("device", id)
	if nil == e {
		t.Errorf("device %s is exists", id)
	}
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
		return fmt.Errorf("delete %s failed, %v", t, e)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("delete %s failed, %v, %v", t, resp.StatusCode, readAll(resp.Body))
	}
	return nil
}
func deleteBy(t string, params map[string]string) error {
	url := "http://127.0.0.1:7071/mdb/" + t + "/query?"
	for k, v := range params {
		url += ("@" + k + "=" + v + "&")
	}
	resp, e := httpDelete(url[:len(url)-1])
	if nil != e {
		return fmt.Errorf("delete %s failed, %v", t, e)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("delete %s failed, %v, %v", t, resp.StatusCode, readAll(resp.Body))
	}
	return nil
}

func count(t string, params map[string]string) (int, error) {
	url := "http://127.0.0.1:7071/mdb/" + t + "/count?"
	for k, v := range params {
		url += ("@" + k + "=" + v + "&")
	}
	resp, e := http.Get(url[:len(url)-1])
	if nil != e {
		return -1, fmt.Errorf("count %s failed, %v", t, e)
	}
	if resp.StatusCode != 200 {
		return -1, fmt.Errorf("count %s failed, %v, %v", t, resp.StatusCode, readAll(resp.Body))
	}
	result := map[string]interface{}{}
	e = json.Unmarshal([]byte(readAll(resp.Body)), &result)

	if nil != e {
		return -1, fmt.Errorf("count %s failed, %v", t, e)
	}
	res := result["value"].(float64)
	return int(res), nil
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

func TestDeviceDeleteCascadeAll(t *testing.T) {
	e := deleteById("device", "all")
	if nil != e {
		t.Errorf("remove all device failed, " + e.Error())
	}
	e = deleteById("interface", "all")
	if nil != e {
		t.Errorf("remove all interface failed, " + e.Error())
	}

	id1 := createMockDevice(t, "1")
	id2 := createMockDevice(t, "2")
	id3 := createMockDevice(t, "3")
	id4 := createMockDevice(t, "4")
	if "" == id1 {
		return
	}

	createMockInterface(t, id1, "10001")
	createMockInterface(t, id1, "10002")
	createMockInterface(t, id1, "10003")
	createMockInterface(t, id1, "10004")

	createMockInterface(t, id2, "20001")
	createMockInterface(t, id2, "20002")
	createMockInterface(t, id2, "20003")
	createMockInterface(t, id2, "20004")

	createMockInterface(t, id3, "30001")
	createMockInterface(t, id3, "30002")
	createMockInterface(t, id3, "30003")
	createMockInterface(t, id3, "30004")

	createMockInterface(t, id4, "40001")
	createMockInterface(t, id4, "40002")
	createMockInterface(t, id4, "40003")
	createMockInterface(t, id4, "40004")

	e = deleteById("device", "all")
	if nil != e {
		t.Errorf("remove all device failed, " + e.Error())
		return
	}

	if c, err := count("interface", map[string]string{}); 0 != c {
		t.Errorf("16 != len(all.interfaces), actual is %d, %v", c, err)
	}
}

func checkHistoryRuleCount(t *testing.T, id1, id2, id3, id4 string, all, d1, d2, d3, d4 int) {
	tName := "history_rule"
	if c, err := count(tName, map[string]string{}); all != c {
		t.Errorf("%d != len(all.rules), actual is %d, %v", all, c, err)
	}
	if c, err := count(tName, map[string]string{"parent_type": "devices", "parent_id": id1}); d1 != c {
		t.Errorf("%d != len(d1.rules), actual is %d, %v", d1, c, err)
	}
	if c, err := count(tName, map[string]string{"parent_type": "devices", "parent_id": id2}); d2 != c {
		t.Errorf("%d != len(d2.rules), actual is %d, %v", d2, c, err)
	}
	if c, err := count(tName, map[string]string{"parent_type": "devices", "parent_id": id3}); d3 != c {
		t.Errorf("%d != len(d3.rules), actual is %d, %v", d3, c, err)
	}
	if c, err := count(tName, map[string]string{"parent_type": "devices", "parent_id": id4}); d4 != c {
		t.Errorf("%d != len(d4.rules), actual is %d, %v", d4, c, err)
	}
}
func checkInterfaceCount(t *testing.T, id1, id2, id3, id4 string, all, d1, d2, d3, d4 int) {
	checkCount(t, "device_id", "interface", id1, id2, id3, id4, all, d1, d2, d3, d4)
}
func checkCount(t *testing.T, field, tName, id1, id2, id3, id4 string, all, d1, d2, d3, d4 int) {
	if c, err := count(tName, map[string]string{}); all != c {
		t.Errorf("%d != len(all.interfaces), actual is %d, %v", all, c, err)
	}
	if c, err := count(tName, map[string]string{field: id1}); d1 != c {
		t.Errorf("%d != len(d1.interfaces), actual is %d, %v", d1, c, err)
	}
	if c, err := count(tName, map[string]string{field: id2}); d2 != c {
		t.Errorf("%d != len(d2.interfaces), actual is %d, %v", d2, c, err)
	}
	if c, err := count(tName, map[string]string{field: id3}); d3 != c {
		t.Errorf("%d != len(d3.interfaces), actual is %d, %v", d3, c, err)
	}
	if c, err := count(tName, map[string]string{field: id4}); d4 != c {
		t.Errorf("%d != len(d4.interfaces), actual is %d, %v", d4, c, err)
	}
}

func TestDeviceDeleteCascadeByAll(t *testing.T) {
	e := deleteById("device", "all")
	if nil != e {
		t.Errorf("remove all device failed, " + e.Error())
	}
	e = deleteById("interface", "all")
	if nil != e {
		t.Errorf("remove all interface failed, " + e.Error())
	}
	e = deleteById("history_rule", "all")
	if nil != e {
		t.Errorf("remove all interface failed, " + e.Error())
	}

	id1 := createMockDevice(t, "1")
	id2 := createMockDevice(t, "2")
	id3 := createMockDevice(t, "3")
	id4 := createMockDevice(t, "4")
	if "" == id1 {
		return
	}
	createMockHistoryRule2(t, "s")
	createMockInterface(t, id1, "10001")
	createMockInterface(t, id1, "10002")
	createMockInterface(t, id1, "10003")
	createMockInterface(t, id1, "10004")
	createMockHistoryRule(t, id1, "10001")
	createMockHistoryRule(t, id1, "10002")
	createMockHistoryRule(t, id1, "10003")
	createMockHistoryRule(t, id1, "10004")

	createMockInterface(t, id2, "20001")
	createMockInterface(t, id2, "20002")
	createMockInterface(t, id2, "20003")
	createMockInterface(t, id2, "20004")
	createMockHistoryRule(t, id2, "20001")
	createMockHistoryRule(t, id2, "20002")
	createMockHistoryRule(t, id2, "20003")
	createMockHistoryRule(t, id2, "20004")

	createMockInterface(t, id3, "30001")
	createMockInterface(t, id3, "30002")
	createMockInterface(t, id3, "30003")
	createMockInterface(t, id3, "30004")
	createMockHistoryRule(t, id3, "30001")
	createMockHistoryRule(t, id3, "30002")
	createMockHistoryRule(t, id3, "30003")
	createMockHistoryRule(t, id3, "30004")

	createMockInterface(t, id4, "40001")
	createMockInterface(t, id4, "40002")
	createMockInterface(t, id4, "40003")
	createMockInterface(t, id4, "40004")
	createMockHistoryRule(t, id4, "40001")
	createMockHistoryRule(t, id4, "40002")
	createMockHistoryRule(t, id4, "40003")
	createMockHistoryRule(t, id4, "40004")

	checkInterfaceCount(t, id1, id2, id3, id4, 16, 4, 4, 4, 4)
	checkHistoryRuleCount(t, id1, id2, id3, id4, 17, 4, 4, 4, 4)
	e = deleteById("device", "all")
	if nil != e {
		t.Errorf("remove dev1 failed, " + e.Error())
	}
	checkInterfaceCount(t, id1, id2, id3, id4, 0, 0, 0, 0, 0)
	checkHistoryRuleCount(t, id1, id2, id3, id4, 1, 0, 0, 0, 0)
}

func TestDeviceDeleteCascadeByQuery(t *testing.T) {
	e := deleteById("device", "all")
	if nil != e {
		t.Errorf("remove all device failed, " + e.Error())
	}
	e = deleteById("interface", "all")
	if nil != e {
		t.Errorf("remove all interface failed, " + e.Error())
	}
	e = deleteById("trigger", "all")
	if nil != e {
		t.Errorf("remove all interface failed, " + e.Error())
	}

	id1 := createMockDevice(t, "1")
	id2 := createMockDevice(t, "2")
	id3 := createMockDevice(t, "3")
	id4 := createMockDevice(t, "4")
	if "" == id1 {
		return
	}

	createMockHistoryRule2(t, "s")
	createMockInterface(t, id1, "10001")
	createMockInterface(t, id1, "10002")
	createMockInterface(t, id1, "10003")
	createMockInterface(t, id1, "10004")
	createMockHistoryRule(t, id1, "10001")
	createMockHistoryRule(t, id1, "10002")
	createMockHistoryRule(t, id1, "10003")
	createMockHistoryRule(t, id1, "10004")

	createMockInterface(t, id2, "20001")
	createMockInterface(t, id2, "20002")
	createMockInterface(t, id2, "20003")
	createMockInterface(t, id2, "20004")
	createMockHistoryRule(t, id2, "20001")
	createMockHistoryRule(t, id2, "20002")
	createMockHistoryRule(t, id2, "20003")
	createMockHistoryRule(t, id2, "20004")

	createMockInterface(t, id3, "30001")
	createMockInterface(t, id3, "30002")
	createMockInterface(t, id3, "30003")
	createMockInterface(t, id3, "30004")
	createMockHistoryRule(t, id3, "30001")
	createMockHistoryRule(t, id3, "30002")
	createMockHistoryRule(t, id3, "30003")
	createMockHistoryRule(t, id3, "30004")

	createMockInterface(t, id4, "40001")
	createMockInterface(t, id4, "40002")
	createMockInterface(t, id4, "40003")
	createMockInterface(t, id4, "40004")
	createMockHistoryRule(t, id4, "40001")
	createMockHistoryRule(t, id4, "40002")
	createMockHistoryRule(t, id4, "40003")
	createMockHistoryRule(t, id4, "40004")

	checkInterfaceCount(t, id1, id2, id3, id4, 16, 4, 4, 4, 4)
	checkHistoryRuleCount(t, id1, id2, id3, id4, 17, 4, 4, 4, 4)
	e = deleteBy("device", map[string]string{"catalog": "[gte]3"})
	if nil != e {
		t.Errorf("remove query failed, " + e.Error())
	}
	checkInterfaceCount(t, id1, id2, id3, id4, 8, 4, 4, 0, 0)
	checkHistoryRuleCount(t, id1, id2, id3, id4, 9, 4, 4, 0, 0)
}

func TestDeviceDeleteCascadeById(t *testing.T) {
	e := deleteById("device", "all")
	if nil != e {
		t.Errorf("remove all device failed, " + e.Error())
	}
	e = deleteById("interface", "all")
	if nil != e {
		t.Errorf("remove all interface failed, " + e.Error())
	}
	e = deleteById("history_rule", "all")
	if nil != e {
		t.Errorf("remove all interface failed, " + e.Error())
	}

	id1 := createMockDevice(t, "1")
	id2 := createMockDevice(t, "2")
	id3 := createMockDevice(t, "3")
	id4 := createMockDevice(t, "4")
	if "" == id1 {
		return
	}

	createMockHistoryRule2(t, "s")
	createMockInterface(t, id1, "10001")
	createMockInterface(t, id1, "10002")
	createMockInterface(t, id1, "10003")
	createMockInterface(t, id1, "10004")
	createMockHistoryRule(t, id1, "10001")
	createMockHistoryRule(t, id1, "10002")
	createMockHistoryRule(t, id1, "10003")
	createMockHistoryRule(t, id1, "10004")

	createMockInterface(t, id2, "20001")
	createMockInterface(t, id2, "20002")
	createMockInterface(t, id2, "20003")
	createMockInterface(t, id2, "20004")
	createMockHistoryRule(t, id2, "20001")
	createMockHistoryRule(t, id2, "20002")
	createMockHistoryRule(t, id2, "20003")
	createMockHistoryRule(t, id2, "20004")

	createMockInterface(t, id3, "30001")
	createMockInterface(t, id3, "30002")
	createMockInterface(t, id3, "30003")
	createMockInterface(t, id3, "30004")
	createMockHistoryRule(t, id3, "30001")
	createMockHistoryRule(t, id3, "30002")
	createMockHistoryRule(t, id3, "30003")
	createMockHistoryRule(t, id3, "30004")

	createMockInterface(t, id4, "40001")
	createMockInterface(t, id4, "40002")
	createMockInterface(t, id4, "40003")
	createMockInterface(t, id4, "40004")
	createMockHistoryRule(t, id4, "40001")
	createMockHistoryRule(t, id4, "40002")
	createMockHistoryRule(t, id4, "40003")
	createMockHistoryRule(t, id4, "40004")

	checkInterfaceCount(t, id1, id2, id3, id4, 16, 4, 4, 4, 4)
	checkHistoryRuleCount(t, id1, id2, id3, id4, 17, 4, 4, 4, 4)

	e = deleteById("device", id1)
	if nil != e {
		t.Errorf("remove dev1 failed, " + e.Error())
	}
	checkInterfaceCount(t, id1, id2, id3, id4, 12, 0, 4, 4, 4)
	checkHistoryRuleCount(t, id1, id2, id3, id4, 13, 0, 4, 4, 4)
	e = deleteById("device", id2)
	if nil != e {
		t.Errorf("remove dev2 failed, " + e.Error())
	}
	checkInterfaceCount(t, id1, id2, id3, id4, 8, 0, 0, 4, 4)
	checkHistoryRuleCount(t, id1, id2, id3, id4, 9, 0, 0, 4, 4)
	e = deleteById("device", id3)
	if nil != e {
		t.Errorf("remove dev3 failed, " + e.Error())
	}

	checkInterfaceCount(t, id1, id2, id3, id4, 4, 0, 0, 0, 4)
	checkHistoryRuleCount(t, id1, id2, id3, id4, 5, 0, 0, 0, 4)
	e = deleteById("device", id4)
	if nil != e {
		t.Errorf("remove dev4 failed, " + e.Error())
	}

	checkInterfaceCount(t, id1, id2, id3, id4, 0, 0, 0, 0, 0)
	checkHistoryRuleCount(t, id1, id2, id3, id4, 1, 0, 0, 0, 0)
}

func TestDeviceCURD(t *testing.T) {
	e := deleteById("device", "all")
	if nil != e {
		t.Errorf("remove all failed, " + e.Error())
	}

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

	e = deleteById("device", "all")
	if nil != e {
		t.Errorf("remove all failed, " + e.Error())
	}

	d1 = getDeviceByName(t, "11")
	d2 = getDeviceByName(t, "21")
	d3 = getDeviceByName(t, "31")
	d4 = getDeviceByName(t, "41")

	if nil != d1 || nil != d2 || nil != d3 || nil != d4 {
		t.Errorf("remove all failed")
	}
}

func TestDeviceDeleteById(t *testing.T) {
	e := deleteById("device", "all")
	if nil != e {
		t.Errorf("remove all failed, " + e.Error())
	}

	id1 := createMockDevice(t, "1")
	id2 := createMockDevice(t, "2")
	id3 := createMockDevice(t, "3")
	id4 := createMockDevice(t, "4")
	if "" == id1 {
		return
	}

	e = deleteById("device", id1)
	if nil != e {
		t.Errorf("remove id1 failed, " + e.Error())
	}
	e = deleteById("device", id2)
	if nil != e {
		t.Errorf("remove id2 failed, " + e.Error())
	}
	e = deleteById("device", id3)
	if nil != e {
		t.Errorf("remove id3 failed, " + e.Error())
	}
	e = deleteById("device", id4)
	if nil != e {
		t.Errorf("remove id4 failed, " + e.Error())
	}

	DeviceNotExistsById(t, id1)
	DeviceNotExistsById(t, id2)
	DeviceNotExistsById(t, id3)
	DeviceNotExistsById(t, id4)
}

func TestDeviceFindBy(t *testing.T) {
	e := deleteById("device", "all")
	if nil != e {
		t.Errorf("remove all failed, " + e.Error())
	}

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
	if 2 != len(res) {
		t.Errorf("catalog <=2 failed, len(result) is %v", len(res))
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
	res, e = findBy("device", map[string]string{"catalog": "[gte]3"})
	if nil != e {
		t.Errorf(e.Error())
		return
	}
	if 2 != len(res) {
		t.Errorf("catalog <=2 failed, len(result) is %v", len(res))
		return
	}
	d3 := searchBy(res, func(r map[string]interface{}) bool { return r["name"] == "dd3" })
	d4 := searchBy(res, func(r map[string]interface{}) bool { return r["name"] == "dd4" })
	if nil == d3 {
		t.Errorf("catalog <=2 failed, result is %v", res)
		return
	}
	validMockDevice(t, "3", d3)
	validMockDevice(t, "4", d4)

	res, e = findBy("device", map[string]string{"catalog": "[ne]3"})
	if nil != e {
		t.Errorf(e.Error())
		return
	}
	if 3 != len(res) {
		t.Errorf("catalog <=3 failed, len(result) is %v", len(res))
		return
	}
	d1 = searchBy(res, func(r map[string]interface{}) bool { return r["name"] == "dd1" })
	d2 = searchBy(res, func(r map[string]interface{}) bool { return r["name"] == "dd2" })
	d4 = searchBy(res, func(r map[string]interface{}) bool { return r["name"] == "dd4" })
	if nil == d1 {
		t.Errorf("catalog <=2 failed, result is %v", res)
		return
	}
	validMockDevice(t, "1", d1)
	validMockDevice(t, "2", d2)
	validMockDevice(t, "4", d4)
}
