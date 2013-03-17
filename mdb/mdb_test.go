package mdb

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
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

func createMockMetricRule2(t *testing.T, factor string) string {
	return createJson(t, "metric_rule", fmt.Sprintf(`{"name":"%s", "expression":"d%s", "metric":"2%s"}`, factor, factor, factor))
}
func createMockMetricRule(t *testing.T, id, factor string) string {
	return createJson(t, "metric_rule", fmt.Sprintf(`{"name":"%s", "expression":"d%s", "metric":"2%s", "parent_type":"device", "parent_id":"%s"}`, factor, factor, factor, id))
}

func createMockInterface(t *testing.T, id, factor string) string {
	return createJson(t, "interface", fmt.Sprintf(`{"ifIndex":%s, "ifDescr":"d%s", "ifType":2%s, "ifMtu":3%s, "ifSpeed":4%s, "device_id":"%s"}`, factor, factor, factor, factor, factor, id))
}

func createMockDevice(t *testing.T, factor string) string {
	return createJson(t, "device", fmt.Sprintf(`{"name":"dd%s", "address":"192.168.1.%s", "catalog":%s, "services":2%s, "managed_address":"20.0.8.110"}`, factor, factor, factor, factor))
}

func updateMockDevice(t *testing.T, id, factor string) {
	updateJson(t, "device", id, fmt.Sprintf(`{"name":"dd%s", "catalog":%s, "services":2%s}`, factor, factor, factor))
}

func getDeviceById(t *testing.T, id string) map[string]interface{} {
	return findById(t, "device", id)
}

func DeviceNotExistsById(t *testing.T, id string) {
	if existsById(t, "device", id) {
		t.Error("device '" + id + "' is exists")
		t.FailNow()
	}
}

func getDeviceByName(t *testing.T, factor string) map[string]interface{} {
	res := findBy(t, "device", map[string]string{"name": "dd" + factor})
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
func validMockDeviceWithId(t *testing.T, factor string, drv map[string]interface{}, id string) {
	if id != drv["_id"].(string) {
		t.Errorf("excepted id is '%s', actual id is '%v'", id, drv["_id"])
		return
	}
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
func create(t *testing.T, target string, body map[string]interface{}) string {
	client := NewClient("http://127.0.0.1:7071/mdb/")
	id, e := client.Create(target, body)
	if nil != e {
		t.Errorf("create %s failed, %v", target, e)
		t.FailNow()
	}
	if nil != client.Warnings {
		t.Error(client.Warnings)
	}
	return id
}

func createJson(t *testing.T, target, msg string) string {
	client := NewClient("http://127.0.0.1:7071/mdb/")
	id, e := client.CreateJson("http://127.0.0.1:7071/mdb/"+target, []byte(msg))
	if nil != e {
		t.Errorf("create %s failed, %v", target, e)
		t.FailNow()
	}
	if nil != client.Warnings {
		t.Error(client.Warnings)
	}
	return id
}

func updateById(t *testing.T, target, id string, body map[string]interface{}) {
	client := NewClient("http://127.0.0.1:7071/mdb/")
	e := client.UpdateById(target, id, body)
	if nil != e {
		t.Errorf("update %s failed, %v", target, e)
		t.FailNow()
	}
	if nil != client.Warnings {
		t.Error(client.Warnings)
	}
}

func updateBy(t *testing.T, target string, params map[string]string, body map[string]interface{}) {
	client := NewClient("http://127.0.0.1:7071/mdb/")
	_, e := client.UpdateBy(target, params, body)
	if nil != e {
		t.Errorf("update %s failed, %v", target, e)
		t.FailNow()
	}
	if nil != client.Warnings {
		t.Error(client.Warnings)
	}
}

// func update(t, id string, body map[string]interface{}) error {
//	msg, e := json.Marshal(body)
//	if nil != e {
//		return fmt.Errorf("update %s failed, %v", t, e)
//	}
//	return updateJson(t, id, string(msg))
// }

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

func updateJson(t *testing.T, target, id, msg string) {

	client := NewClient("http://127.0.0.1:7071/mdb/")
	_, e := client.UpdateJson("http://127.0.0.1:7071/mdb/"+target+"/"+id, []byte(msg))
	if nil != e {
		t.Errorf("update %s failed, %v", target, e)
		t.FailNow()
	}
	if nil != client.Warnings {
		t.Error(client.Warnings)
	}
}

func existsById(t *testing.T, target, id string) bool {
	client := NewClient("http://127.0.0.1:7071/mdb/")
	_, e := client.FindByIdWithIncludes(target, id, "")
	if nil != e {
		if 404 == e.Code() {
			return false
		}
		t.Errorf("find %s failed, %v", target, e)
		t.FailNow()
	}
	if nil != client.Warnings {
		t.Error(client.Warnings)
	}
	return true
}

func findById(t *testing.T, target, id string) map[string]interface{} {
	return findByIdWithIncludes(t, target, id, "")
}

func findByIdWithIncludes(t *testing.T, target, id string, includes string) map[string]interface{} {
	client := NewClient("http://127.0.0.1:7071/mdb/")
	res, e := client.FindByIdWithIncludes(target, id, includes)
	if nil != e {
		t.Errorf("find %s failed, %v", target, e)
		t.FailNow()
	}
	if nil != client.Warnings {
		t.Error(client.Warnings)
	}
	return res
}

func deleteByIdWhileNotExist(t *testing.T, target, id string) {
	client := NewClient("http://127.0.0.1:7071/mdb/")
	e := client.DeleteById(target, id)
	if nil == e {
		t.Errorf("delete not exist %s failed", target)
		t.FailNow()
	}
	if nil != client.Warnings {
		t.Error(client.Warnings)
	}
	if 404 != e.Code() {
		t.Errorf("delete not exist %s failed, %v", target, e)
		t.FailNow()
	}
}

func deleteById(t *testing.T, target, id string) {
	client := NewClient("http://127.0.0.1:7071/mdb/")
	e := client.DeleteById(target, id)
	if nil != e {
		t.Errorf("delete %s failed, %v", target, e)
		t.FailNow()
	}
	if nil != client.Warnings {
		t.Error(client.Warnings)
	}
}

func deleteBy(t *testing.T, target string, params map[string]string) {
	client := NewClient("http://127.0.0.1:7071/mdb/")
	_, e := client.DeleteBy(target, params)
	if nil != e {
		t.Errorf("delete %s failed, %v", target, e)
		t.FailNow()
	}
	if nil != client.Warnings {
		t.Error(client.Warnings)
	}
}

func count(t *testing.T, target string, params map[string]string) int {
	client := NewClient("http://127.0.0.1:7071/mdb/")
	i, e := client.Count(target, params)
	if nil != e {
		t.Errorf("count %s failed, %v", target, e)
		t.FailNow()
	}
	if nil != client.Warnings {
		t.Error(client.Warnings)
	}
	return i
}

func findOne(t *testing.T, target string, params map[string]string) map[string]interface{} {
	return findOneByQueryWithIncludes(t, target, params, "")
}

func findOneByQueryWithIncludes(t *testing.T, target string, params map[string]string, includes string) map[string]interface{} {
	res := findByQueryWithIncludes(t, target, params, includes)
	if 0 == len(res) {
		t.Errorf("find %s failed, result is empty", target)
		t.FailNow()
	}
	return res[0]
}

func ExistsInChilren(t *testing.T, attrs map[string]interface{}, target string, params map[string]string) bool {
	a := attrs["$"+target]
	if nil == a {
		return false
	}
	aa, ok := a.([]interface{})
	if !ok {
		t.Error("$" + target + " is not a []interface{} type")
		t.FailNow()
	}
	for _, r := range aa {
		res, ok := r.(map[string]interface{})
		if !ok {
			t.Error("$" + target + " is not a map[string]interface{} type")
			t.FailNow()
		}
		found := true
		for k, v := range params {
			if v != fmt.Sprint(res[k]) {
				found = false
				break
			}
		}
		if found {
			return true
		}
	}

	return false
}

func findOneFrom(t *testing.T, attrs map[string]interface{}, target string, params map[string]string) map[string]interface{} {
	a := attrs["$"+target]
	if nil == a {
		t.Error("$" + target + " is not found")
		t.FailNow()
	}
	aa, ok := a.([]interface{})
	if !ok {
		t.Error("$" + target + " is not a []interface{} type")
		t.FailNow()
	}
	for _, r := range aa {
		res, ok := r.(map[string]interface{})
		if !ok {
			t.Error("$" + target + " is not a map[string]interface{} type")
			t.FailNow()
		}
		found := true
		for k, v := range params {
			if v != fmt.Sprint(res[k]) {
				found = false
				break
			}
		}
		if found {
			return res
		}
	}

	t.Error("$" + target + " is not found with the speacfic params")
	t.FailNow()
	return nil
}

func findBy(t *testing.T, target string, params map[string]string) []map[string]interface{} {
	return findByQueryWithIncludes(t, target, params, "")
}

func findByQueryWithIncludes(t *testing.T, target string, params map[string]string, includes string) []map[string]interface{} {
	url := "http://127.0.0.1:7071/mdb/" + target + "/query?"
	if "" != includes {
		url += ("includes=" + includes + "&")
	}
	for k, v := range params {
		url += ("@" + k + "=" + v + "&")
	}
	resp, e := http.Get(url[:len(url)-1])
	if nil != e {
		t.Errorf("find %s failed, %v", target, e)
		t.FailNow()
	}
	if resp.StatusCode != 200 {
		t.Errorf("find %s failed, %v, %v", target, resp.StatusCode, readAll(resp.Body))
		t.FailNow()
	}
	body := readAll(resp.Body)
	result := map[string]interface{}{}
	e = json.Unmarshal([]byte(body), &result)

	if nil != e {
		t.Errorf("find %s failed, %v, %v", target, e, body)
		t.FailNow()
	}
	res := result["value"].([]interface{})
	results := make([]map[string]interface{}, 0, 3)
	for _, r := range res {
		results = append(results, r.(map[string]interface{}))
	}
	return results
}
func findByParent(t *testing.T, parent, parent_id, target string, params map[string]string) []map[string]interface{} {
	client := NewClient("http://127.0.0.1:7071/mdb/")
	res, e := client.Children(parent, parent_id, target, params)
	if nil != e {
		t.Errorf("find %s failed, %v", target, e)
		t.FailNow()
	}
	if nil != client.Warnings {
		t.Error(client.Warnings)
	}
	return res
}

func TestDeviceDeleteCascadeAll(t *testing.T) {
	deleteById(t, "device", "all")
	deleteById(t, "interface", "all")

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

	deleteById(t, "device", "all")

	if c := count(t, "interface", map[string]string{}); 0 != c {
		t.Errorf("16 != len(all.interfaces), actual is %d", c)
	}
}

func checkMetricRuleCount(t *testing.T, id1, id2, id3, id4 string, all, d1, d2, d3, d4 int) {
	tName := "metric_rule"
	if c := count(t, tName, map[string]string{}); all != c {
		t.Errorf("%d != len(all.rules), actual is %d", all, c)
	}
	if c := count(t, tName, map[string]string{"parent_type": "device", "parent_id": id1}); d1 != c {
		t.Errorf("%d != len(d1.rules), actual is %d", d1, c)
	}
	if c := count(t, tName, map[string]string{"parent_type": "device", "parent_id": id2}); d2 != c {
		t.Errorf("%d != len(d2.rules), actual is %d", d2, c)
	}
	if c := count(t, tName, map[string]string{"parent_type": "device", "parent_id": id3}); d3 != c {
		t.Errorf("%d != len(d3.rules), actual is %d", d3, c)
	}
	if c := count(t, tName, map[string]string{"parent_type": "device", "parent_id": id4}); d4 != c {
		t.Errorf("%d != len(d4.rules), actual is %d", d4, c)
	}
}
func checkInterfaceCount(t *testing.T, id1, id2, id3, id4 string, all, d1, d2, d3, d4 int) {
	checkCount(t, "device_id", "interface", id1, id2, id3, id4, all, d1, d2, d3, d4)
}
func checkCount(t *testing.T, field, tName, id1, id2, id3, id4 string, all, d1, d2, d3, d4 int) {
	if c := count(t, tName, map[string]string{}); all != c {
		t.Errorf("%d != len(all.interfaces), actual is %d", all, c)
	}
	if c := count(t, tName, map[string]string{field: id1}); d1 != c {
		t.Errorf("%d != len(d1.interfaces), actual is %d", d1, c)
	}
	if c := count(t, tName, map[string]string{field: id2}); d2 != c {
		t.Errorf("%d != len(d2.interfaces), actual is %d", d2, c)
	}
	if c := count(t, tName, map[string]string{field: id3}); d3 != c {
		t.Errorf("%d != len(d3.interfaces), actual is %d", d3, c)
	}
	if c := count(t, tName, map[string]string{field: id4}); d4 != c {
		t.Errorf("%d != len(d4.interfaces), actual is %d", d4, c)
	}
}

func initData(t *testing.T) []string {

	deleteById(t, "device", "all")
	deleteById(t, "interface", "all")
	deleteById(t, "trigger", "all")

	id1 := createMockDevice(t, "1")
	id2 := createMockDevice(t, "2")
	id3 := createMockDevice(t, "3")
	id4 := createMockDevice(t, "4")

	createMockMetricRule2(t, "s")
	createMockInterface(t, id1, "10001")
	createMockInterface(t, id1, "10002")
	createMockInterface(t, id1, "10003")
	createMockInterface(t, id1, "10004")
	createMockMetricRule(t, id1, "10001")
	createMockMetricRule(t, id1, "10002")
	createMockMetricRule(t, id1, "10003")
	createMockMetricRule(t, id1, "10004")

	createMockInterface(t, id2, "20001")
	createMockInterface(t, id2, "20002")
	createMockInterface(t, id2, "20003")
	createMockInterface(t, id2, "20004")
	createMockMetricRule(t, id2, "20001")
	createMockMetricRule(t, id2, "20002")
	createMockMetricRule(t, id2, "20003")
	createMockMetricRule(t, id2, "20004")

	createMockInterface(t, id3, "30001")
	createMockInterface(t, id3, "30002")
	createMockInterface(t, id3, "30003")
	createMockInterface(t, id3, "30004")
	createMockMetricRule(t, id3, "30001")
	createMockMetricRule(t, id3, "30002")
	createMockMetricRule(t, id3, "30003")
	createMockMetricRule(t, id3, "30004")

	createMockInterface(t, id4, "40001")
	createMockInterface(t, id4, "40002")
	createMockInterface(t, id4, "40003")
	createMockInterface(t, id4, "40004")
	createMockMetricRule(t, id4, "40001")
	createMockMetricRule(t, id4, "40002")
	createMockMetricRule(t, id4, "40003")
	createMockMetricRule(t, id4, "40004")
	return []string{id1, id2, id3, id4}
}
func TestDeviceDeleteCascadeByAll(t *testing.T) {

	idlist := initData(t)

	checkInterfaceCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 16, 4, 4, 4, 4)
	checkMetricRuleCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 17, 4, 4, 4, 4)
	deleteById(t, "device", "all")
	checkInterfaceCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 0, 0, 0, 0, 0)
	checkMetricRuleCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 1, 0, 0, 0, 0)
}

func TestDeviceDeleteCascadeByQuery(t *testing.T) {

	idlist := initData(t)
	checkInterfaceCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 16, 4, 4, 4, 4)
	checkMetricRuleCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 17, 4, 4, 4, 4)
	deleteBy(t, "device", map[string]string{"catalog": "[gte]3"})
	checkInterfaceCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 8, 4, 4, 0, 0)
	checkMetricRuleCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 9, 4, 4, 0, 0)
}

func TestDeviceDeleteCascadeById(t *testing.T) {

	idlist := initData(t)

	checkInterfaceCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 16, 4, 4, 4, 4)
	checkMetricRuleCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 17, 4, 4, 4, 4)
	deleteById(t, "device", idlist[0])

	checkInterfaceCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 12, 0, 4, 4, 4)
	checkMetricRuleCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 13, 0, 4, 4, 4)
	deleteById(t, "device", idlist[1])

	checkInterfaceCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 8, 0, 0, 4, 4)
	checkMetricRuleCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 9, 0, 0, 4, 4)
	deleteById(t, "device", idlist[2])

	checkInterfaceCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 4, 0, 0, 0, 4)
	checkMetricRuleCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 5, 0, 0, 0, 4)
	deleteById(t, "device", idlist[3])

	checkInterfaceCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 0, 0, 0, 0, 0)
	checkMetricRuleCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 1, 0, 0, 0, 0)
}

func TestDeviceCURD(t *testing.T) {
	deleteById(t, "device", "all")
	deleteById(t, "interface", "all")
	deleteById(t, "trigger", "all")

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

	validMockDeviceWithId(t, "1", d1, id1)
	validMockDeviceWithId(t, "2", d2, id2)
	validMockDeviceWithId(t, "3", d3, id3)
	validMockDeviceWithId(t, "4", d4, id4)

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

	deleteById(t, "device", "all")

	d1 = getDeviceByName(t, "11")
	d2 = getDeviceByName(t, "21")
	d3 = getDeviceByName(t, "31")
	d4 = getDeviceByName(t, "41")

	if nil != d1 || nil != d2 || nil != d3 || nil != d4 {
		t.Errorf("remove all failed")
	}
}

func TestDeviceDeleteById(t *testing.T) {
	deleteById(t, "device", "all")

	id1 := createMockDevice(t, "1")
	id2 := createMockDevice(t, "2")
	id3 := createMockDevice(t, "3")
	id4 := createMockDevice(t, "4")
	if "" == id1 {
		return
	}

	deleteById(t, "device", id1)
	deleteById(t, "device", id2)
	deleteById(t, "device", id3)
	deleteById(t, "device", id4)

	DeviceNotExistsById(t, id1)
	DeviceNotExistsById(t, id2)
	DeviceNotExistsById(t, id3)
	DeviceNotExistsById(t, id4)

	deleteByIdWhileNotExist(t, "device", bson.NewObjectId().Hex())
}

func TestDeviceFindBy(t *testing.T) {
	deleteById(t, "device", "all")

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

	res := findBy(t, "device", map[string]string{"catalog": "[eq]1"})
	validMockDevice(t, "1", res[0])

	res = findBy(t, "device", map[string]string{"catalog": "[lte]1"})
	validMockDevice(t, "1", res[0])

	res = findBy(t, "device", map[string]string{"catalog": "[lte]2"})
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

	res = findBy(t, "device", map[string]string{"catalog": "[lt]2"})
	validMockDevice(t, "1", res[0])

	res = findBy(t, "device", map[string]string{"catalog": "[gt]3"})
	validMockDevice(t, "4", res[0])
	res = findBy(t, "device", map[string]string{"catalog": "[gte]3"})
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

	res = findBy(t, "device", map[string]string{"catalog": "[ne]3"})
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
