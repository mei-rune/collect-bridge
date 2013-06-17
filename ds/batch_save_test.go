package ds

import (
	"commons/types"
	"encoding/json"
	"fmt"
	"testing"
)

var (
	snmp_params = map[string]interface{}{"address": "192.168.1.9",
		"community": "public",
		"port":      161,
		"type":      "snmp_param",
		"version":   "v2c"}
	wbem_params = map[string]interface{}{"url": "tcp://192.168.1.9",
		"user":     "user1",
		"password": "password1",
		"type":     "wbem_param"}

	if1 = map[string]interface{}{
		"name":          "1",
		"ifDescr":       "Software Loopback Interface 1",
		"ifIndex":       1,
		"ifMtu":         1500,
		"ifPhysAddress": "",
		"ifSpeed":       1073741824,
		"ifType":        24}
	if2 = map[string]interface{}{
		"name":          "9",
		"ifDescr":       "RAS Async Adapter",
		"ifIndex":       9,
		"ifMtu":         0,
		"ifPhysAddress": "20:41:53:59:4e:ff",
		"ifSpeed":       0,
		"ifType":        23}
	ip1 = map[string]interface{}{
		"name":         "127.0.0.1",
		"address":      "127.0.0.1",
		"bcastAddress": 1,
		"ifIndex":      1,
		"netmask":      "255.0.0.0",
		"reasmMaxSize": 65535}
	ip2 = map[string]interface{}{
		"name":         "169.254.67.142",
		"address":      "169.254.67.142",
		"bcastAddress": 1,
		"ifIndex":      27,
		"netmask":      "255.255.0.0",
		"reasmMaxSize": 65535}
	ip3 = map[string]interface{}{
		"name":         "192.168.1.9",
		"address":      "192.168.1.9",
		"bcastAddress": 1,
		"ifIndex":      13,
		"netmask":      "255.255.255.0",
		"reasmMaxSize": 65535}

	rule1 = map[string]interface{}{"name": "rule1", "type": "metric_trigger", "expression": "d111", "metric": "2221"}
	rule2 = map[string]interface{}{"name": "rule2", "type": "metric_trigger", "expression": "d222", "metric": "2222"}

	action1 = map[string]interface{}{"name": "action1", "type": "redis_command", "command": "c111", "arg0": "2221"}
	action2 = map[string]interface{}{"name": "action2", "type": "redis_command", "command": "c222", "arg0": "3332"}

	js = map[string]interface{}{
		"$access_param": []map[string]interface{}{snmp_params, wbem_params},
		"$address":      []interface{}{ip1, ip2, ip3},
		"$interface":    []interface{}{if1, if2},
		"$trigger":      []interface{}{rule1, rule2},
		"address":       "192.168.1.9",
		"catalog":       2,
		"description":   "Hardware: Intel64 Family 6 Model 58 Stepping 9 AT/AT COMPATIBLE - Software:Windows Version 6.1 (Build 7601 Multiprocessor Free)",
		"location":      "",
		"name":          "meifakun-PC",
		"oid":           "1.3.6.1.4.1.311.1.1.3.1.1",
		"services":      76}
)

func checkStringField(t *testing.T, actual, excepted map[string]interface{}, name string) {
	if actual[name] != excepted[name] {
		if nil == actual[name] || "" == actual[name] {
			if nil == excepted[name] || "" == excepted[name] {
				return
			}
		}
		t.Errorf("actual[%s] is %v, excepted[%s] is %v", name, actual[name], name, excepted[name])
	}
}
func checkIntField(t *testing.T, actual, excepted map[string]interface{}, name string) {
	if fetchInt(actual, name) != excepted[name] {
		t.Errorf("actual[%s] is %v, excepted[%s] is %v", name, actual[name], name, excepted[name])
	}
}

func checkDevice(t *testing.T, actual, excepted map[string]interface{}) {
	for _, name := range []string{"address", "description", "location", "name", "oid"} {
		checkStringField(t, actual, excepted, name)
	}
	for _, name := range []string{"catalog", "services"} {
		checkIntField(t, actual, excepted, name)
	}
}

func checkInterface(t *testing.T, actual, excepted map[string]interface{}) {
	for _, name := range []string{"ifDescr", "ifPhysAddress"} {
		checkStringField(t, actual, excepted, name)
	}
	for _, name := range []string{"ifIndex", "ifMtu", "ifSpeed", "ifType"} {
		checkIntField(t, actual, excepted, name)
	}
}
func checkAddress(t *testing.T, actual, excepted map[string]interface{}) {
	for _, name := range []string{"address", "netmask"} {
		checkStringField(t, actual, excepted, name)
	}
	for _, name := range []string{"bcastAddress", "ifIndex", "reasmMaxSize"} {
		checkIntField(t, actual, excepted, name)
	}
}

func checkSnmpParams(t *testing.T, actual, excepted map[string]interface{}) {
	for _, name := range []string{"address", "community", "type", "version"} {
		checkStringField(t, actual, excepted, name)
	}
	for _, name := range []string{"port"} {
		checkIntField(t, actual, excepted, name)
	}
}

func checkHistoryRule(t *testing.T, actual, excepted map[string]interface{}) {
	for _, name := range []string{"name", "type", "expression", "metric"} {
		checkStringField(t, actual, excepted, name)
	}
}
func checkRedisAction(t *testing.T, actual, excepted map[string]interface{}) {
	for _, name := range []string{"name", "type", "command", "arg0"} {
		checkStringField(t, actual, excepted, name)
	}
}

func TestIntCreateDevice(t *testing.T) {
	srvTest2(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "device", map[string]string{})
		deleteBy(t, client, "interface", map[string]string{})
		deleteBy(t, client, "trigger", map[string]string{})
		deleteBy(t, client, "action", map[string]string{})

		id := create(t, client, "device", js)

		checkDevice(t, findById(t, client, "device", id), js)
		t.Log("find device ok")
		checkInterface(t, findOne(t, client, "interface", map[string]string{"device_id": id, "ifIndex": "1"}), if1)
		checkInterface(t, findOne(t, client, "interface", map[string]string{"device_id": id, "ifIndex": "9"}), if2)

		checkHistoryRule(t, findOne(t, client, "trigger", map[string]string{"parent_type": "device", "parent_id": id, "name": "rule1"}), rule1)
		checkHistoryRule(t, findOne(t, client, "trigger", map[string]string{"parent_type": "device", "parent_id": id, "name": "rule2"}), rule2)

		checkSnmpParams(t, findOne(t, client, "snmp_param", map[string]string{"managed_object_id": id, "address": "192.168.1.9"}), snmp_params)
	})
}

func TestIntQueryByIncludes(t *testing.T) {
	srvTest2(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "device", map[string]string{})
		deleteBy(t, client, "interface", map[string]string{})
		deleteBy(t, client, "trigger", map[string]string{})
		deleteBy(t, client, "action", map[string]string{})

		id := create(t, client, "device", js)
		drv := findByIdWithIncludes(t, client, "device", id, "interface,trigger,snmp_param")

		checkDevice(t, drv, js)
		bs, e := json.MarshalIndent(drv, "", "  ")
		if nil != e {
			t.Error(e)
		} else {
			t.Log(string(bs))
		}

		checkInterface(t, findOneFrom(t, drv, "interface", map[string]string{"ifIndex": "1"}), if1)
		checkInterface(t, findOneFrom(t, drv, "interface", map[string]string{"ifIndex": "9"}), if2)
		if ExistsInChilren(t, drv, "snmp_param", map[string]string{"type": "wbem_params"}) ||
			ExistsInChilren(t, drv, "wbem_param", map[string]string{"type": "wbem_params"}) ||
			ExistsInChilren(t, drv, "access_param", map[string]string{"type": "wbem_params"}) {
			t.Error("wbem_params in result")
		}

		checkHistoryRule(t, findOneFrom(t, drv, "trigger", map[string]string{"parent_type": "device", "parent_id": id, "name": "rule1"}), rule1)
		checkHistoryRule(t, findOneFrom(t, drv, "trigger", map[string]string{"parent_type": "device", "parent_id": id, "name": "rule2"}), rule2)
		checkSnmpParams(t, findOneFrom(t, drv, "snmp_param", map[string]string{"managed_object_id": id, "address": "192.168.1.9"}), snmp_params)
	})
}

func TestIntQueryByIncludesAll(t *testing.T) {
	srvTest2(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "device", map[string]string{})
		deleteBy(t, client, "interface", map[string]string{})
		deleteBy(t, client, "trigger", map[string]string{})
		deleteBy(t, client, "action", map[string]string{})

		id := create(t, client, "device", js)
		drv := findByIdWithIncludes(t, client, "device", id, "*")

		checkDevice(t, drv, js)
		bs, e := json.MarshalIndent(drv, "", "  ")
		if nil != e {
			t.Error(e)
		} else {
			t.Log(string(bs))
		}

		checkInterface(t, findOneFrom(t, drv, "interface", map[string]string{"ifIndex": "1"}), if1)
		checkInterface(t, findOneFrom(t, drv, "interface", map[string]string{"ifIndex": "9"}), if2)
		if ExistsInChilren(t, drv, "attributes", map[string]string{"type": "wbem_params"}) ||
			ExistsInChilren(t, drv, "attributes", map[string]string{"type": "wbem_params"}) ||
			ExistsInChilren(t, drv, "attributes", map[string]string{"type": "wbem_params"}) {
			t.Error("wbem_params in result")
		}

		checkHistoryRule(t, findOneFrom(t, drv, "trigger", map[string]string{"parent_type": "device", "parent_id": id, "name": "rule1"}), rule1)
		checkHistoryRule(t, findOneFrom(t, drv, "trigger", map[string]string{"parent_type": "device", "parent_id": id, "name": "rule2"}), rule2)
		checkSnmpParams(t, findOneFrom(t, drv, "attributes", map[string]string{"managed_object_id": id, "address": "192.168.1.9"}), snmp_params)
	})
}

func TestIntQueryByParent(t *testing.T) {
	srvTest2(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "device", map[string]string{})
		deleteBy(t, client, "interface", map[string]string{})
		deleteBy(t, client, "trigger", map[string]string{})
		deleteBy(t, client, "action", map[string]string{})

		id := create(t, client, "device", js)
		interfaces := findByParent(t, client, "device", id, "interface", nil)

		d1 := searchBy(interfaces, func(r map[string]interface{}) bool { return fmt.Sprint(r["ifIndex"]) == "1" })
		d2 := searchBy(interfaces, func(r map[string]interface{}) bool { return fmt.Sprint(r["ifIndex"]) == "9" })

		checkInterface(t, d1, if1)
		checkInterface(t, d2, if2)

		triggers := findByParent(t, client, "device", id, "metric_trigger", nil)

		r1 := searchBy(triggers, func(r map[string]interface{}) bool { return fmt.Sprint(r["name"]) == "rule1" })
		r2 := searchBy(triggers, func(r map[string]interface{}) bool { return fmt.Sprint(r["name"]) == "rule2" })
		checkHistoryRule(t, r1, rule1)
		checkHistoryRule(t, r2, rule2)

		tr1 := fmt.Sprint(r1["id"])
		action1["parent_type"] = "metric_trigger"
		action1["parent_id"] = tr1
		action2["parent_type"] = "metric_trigger"
		action2["parent_id"] = tr1

		create(t, client, "redis_command", action1)
		create(t, client, "redis_command", action2)

		actions := findByParent(t, client, "metric_trigger", tr1, "redis_command", nil)
		a1 := searchBy(actions, func(r map[string]interface{}) bool { return fmt.Sprint(r["name"]) == "action1" })
		a2 := searchBy(actions, func(r map[string]interface{}) bool { return fmt.Sprint(r["name"]) == "action2" })

		bs, e := json.MarshalIndent(actions, "", "  ")
		if nil != e {
			t.Error(e)
		} else {
			t.Log(string(bs))
		}

		checkRedisAction(t, a1, action1)
		checkRedisAction(t, a2, action2)
	})
}

func TestIntQueryByChild(t *testing.T) {
	srvTest2(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "device", map[string]string{})
		deleteBy(t, client, "interface", map[string]string{})
		deleteBy(t, client, "trigger", map[string]string{})
		deleteBy(t, client, "action", map[string]string{})

		id := create(t, client, "device", js)

		ifc1 := findOne(t, client, "interface", map[string]string{"device_id": id, "ifIndex": "1"})
		ifc2 := findOne(t, client, "interface", map[string]string{"device_id": id, "ifIndex": "9"})

		checkDevice(t, findByChild(t, client, "device", "interface", fmt.Sprint(ifc1["id"])), js)
		checkDevice(t, findByChild(t, client, "device", "interface", fmt.Sprint(ifc2["id"])), js)

		tr1 := findOne(t, client, "trigger", map[string]string{"parent_type": "device", "parent_id": id, "name": "rule1"})
		tr2 := findOne(t, client, "trigger", map[string]string{"parent_type": "device", "parent_id": id, "name": "rule2"})

		checkDevice(t, findByChild(t, client, "device", "trigger", fmt.Sprint(tr1["id"])), js)
		checkDevice(t, findByChild(t, client, "device", "trigger", fmt.Sprint(tr2["id"])), js)
	})
}
