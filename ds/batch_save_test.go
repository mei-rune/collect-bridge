package ds

import (
	"commons/types"
	"encoding/json"
	"fmt"
	"testing"
)

var (
	snmp_params = map[string]interface{}{"read_community": "public",
		"port":    161,
		"type":    "snmp_param",
		"version": "v2c"}
	wbem_params = map[string]interface{}{"url": "tcp://192.168.1.9",
		"user":     "user1",
		"password": "password1",
		"type":     "wbem_param"}

	if1 = map[string]interface{}{
		"name":           "1",
		"if_descr":       "Software Loopback Interface 1",
		"if_index":       1,
		"if_mtu":         1500,
		"if_physAddress": "",
		"if_speed":       1073741824,
		"if_type":        24}
	if2 = map[string]interface{}{
		"name":           "9",
		"if_descr":       "RAS Async Adapter",
		"if_index":       9,
		"if_mtu":         0,
		"if_physAddress": "20:41:53:59:4e:ff",
		"if_speed":       0,
		"if_type":        23}
	ip1 = map[string]interface{}{
		"name":         "127.0.0.1",
		"address":      "127.0.0.1",
		"bcastAddress": 1,
		"if_index":     1,
		"netmask":      "255.0.0.0",
		"reasmMaxSize": 65535}
	ip2 = map[string]interface{}{
		"name":         "169.254.67.142",
		"address":      "169.254.67.142",
		"bcastAddress": 1,
		"if_index":     27,
		"netmask":      "255.255.0.0",
		"reasmMaxSize": 65535}
	ip3 = map[string]interface{}{
		"name":         "192.168.1.9",
		"address":      "192.168.1.9",
		"bcastAddress": 1,
		"if_index":     13,
		"netmask":      "255.255.255.0",
		"reasmMaxSize": 65535}

	rule1 = map[string]interface{}{"name": "rule1", "type": "metric_trigger", "expression": "d111", "metric": "2221"}
	rule2 = map[string]interface{}{"name": "rule2", "type": "metric_trigger", "expression": "d222", "metric": "2222"}

	action1 = map[string]interface{}{"name": "action1", "type": "redis_command", "command": "c111", "arg0": "2221"}
	action2 = map[string]interface{}{"name": "action2", "type": "redis_command", "command": "c222", "arg0": "3332"}

	js = map[string]interface{}{
		"$access_param":        []map[string]interface{}{snmp_params, wbem_params},
		"$network_address":     []interface{}{ip1, ip2, ip3},
		"$network_device_port": map[string]interface{}{"1": if1, "9": if2},
		"$metric_trigger":      []interface{}{rule1, rule2},
		"address":              "192.168.1.9",
		"description":          "Hardware: Intel64 Family 6 Model 58 Stepping 9 AT/AT COMPATIBLE - Software:Windows Version 6.1 (Build 7601 Multiprocessor Free)",
		"location":             "",
		"name":                 "meifakun-PC",
		"device_type":          2,
		"oid":                  "1.3.6.1.4.1.311.1.1.3.1.1",
		"services":             76}

	dev_js = map[string]interface{}{
		"address":     "192.168.1.9",
		"device_type": 2,
		"description": "Hardware: Intel64 Family 6 Model 58 Stepping 9 AT/AT COMPATIBLE - Software:Windows Version 6.1 (Build 7601 Multiprocessor Free)",
		"location":    "",
		"name":        "meifakun-PC",
		"oid":         "1.3.6.1.4.1.311.1.1.3.1.1",
		"services":    76}
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
	for _, name := range []string{"device_type"} {
		checkIntField(t, actual, excepted, name)
	}
}

func checkInterface(t *testing.T, actual, excepted map[string]interface{}) {
	for _, name := range []string{"if_descr", "if_physAddress"} {
		checkStringField(t, actual, excepted, name)
	}
	for _, name := range []string{"if_index", "if_mtu", "if_speed", "if_type"} {
		checkIntField(t, actual, excepted, name)
	}
}
func checkAddress(t *testing.T, actual, excepted map[string]interface{}) {
	for _, name := range []string{"address", "netmask"} {
		checkStringField(t, actual, excepted, name)
	}
	for _, name := range []string{"bcastAddress", "if_index", "reasmMaxSize"} {
		checkIntField(t, actual, excepted, name)
	}
}

func checkSnmpParams(t *testing.T, actual, excepted map[string]interface{}) {
	for _, name := range []string{"read_community", "type", "version"} {
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
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "network_device", map[string]string{})
		deleteBy(t, client, "network_device_port", map[string]string{})
		deleteBy(t, client, "trigger", map[string]string{})
		deleteBy(t, client, "action", map[string]string{})

		id := create(t, client, "network_device", js)

		checkDevice(t, findById(t, client, "network_device", id), js)
		t.Log("find device ok")
		checkInterface(t, findOne(t, client, "network_device_port", map[string]string{"device_id": id, "if_index": "1"}), if1)
		checkInterface(t, findOne(t, client, "network_device_port", map[string]string{"device_id": id, "if_index": "9"}), if2)

		checkHistoryRule(t, findOne(t, client, "metric_trigger", map[string]string{"managed_object_id": id, "name": "rule1"}), rule1)
		checkHistoryRule(t, findOne(t, client, "metric_trigger", map[string]string{"managed_object_id": id, "name": "rule2"}), rule2)

		checkSnmpParams(t, findOne(t, client, "snmp_param", map[string]string{"managed_object_id": id, "port": "161"}), snmp_params)
	})
}
func TestIntCreateDeviceByParent(t *testing.T) {
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "network_device", map[string]string{})
		deleteBy(t, client, "network_device_port", map[string]string{})
		deleteBy(t, client, "trigger", map[string]string{})
		deleteBy(t, client, "action", map[string]string{})

		id := create(t, client, "network_device", dev_js)

		// "$access_param": []map[string]interface{}{snmp_params, wbem_params},
		// "$address":      []interface{}{ip1, ip2, ip3},
		// "$interface":    map[string]interface{}{"1": if1, "9": if2},
		// "$trigger":      []interface{}{rule1, rule2},

		createByParent(t, client, "network_device", id, "snmp_param", snmp_params)
		createByParent(t, client, "network_device", id, "wbem_param", wbem_params)
		createByParent(t, client, "network_device", id, "network_address", ip1)
		createByParent(t, client, "network_device", id, "network_address", ip2)
		createByParent(t, client, "network_device", id, "network_address", ip3)

		createByParent(t, client, "network_device", id, "network_device_port", if1)
		createByParent(t, client, "network_device", id, "network_device_port", if2)

		createByParent(t, client, "network_device", id, "metric_trigger", rule1)
		createByParent(t, client, "network_device", id, "metric_trigger", rule2)

		checkDevice(t, findById(t, client, "network_device", id), js)
		t.Log("find device ok")
		checkInterface(t, findOne(t, client, "network_device_port", map[string]string{"device_id": id, "if_index": "1"}), if1)
		checkInterface(t, findOne(t, client, "network_device_port", map[string]string{"device_id": id, "if_index": "9"}), if2)

		checkHistoryRule(t, findOne(t, client, "metric_trigger", map[string]string{"managed_object_id": id, "name": "rule1"}), rule1)
		checkHistoryRule(t, findOne(t, client, "metric_trigger", map[string]string{"managed_object_id": id, "name": "rule2"}), rule2)

		checkSnmpParams(t, findOne(t, client, "snmp_param", map[string]string{"managed_object_id": id, "port": "161"}), snmp_params)
	})
}

func TestIntQueryByIncludes(t *testing.T) {
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "network_device", map[string]string{})
		deleteBy(t, client, "network_device_port", map[string]string{})
		deleteBy(t, client, "trigger", map[string]string{})
		deleteBy(t, client, "action", map[string]string{})

		id := create(t, client, "network_device", js)
		drv := findByIdWithIncludes(t, client, "network_device", id, "network_device_port,metric_trigger,snmp_param")

		checkDevice(t, drv, js)
		bs, e := json.MarshalIndent(drv, "", "  ")
		if nil != e {
			t.Error(e)
		} else {
			t.Log(string(bs))
		}

		checkInterface(t, findOneFrom(t, drv, "network_device_port", map[string]string{"if_index": "1"}), if1)
		checkInterface(t, findOneFrom(t, drv, "network_device_port", map[string]string{"if_index": "9"}), if2)
		if ExistsInChilren(t, drv, "snmp_param", map[string]string{"type": "wbem_params"}) ||
			ExistsInChilren(t, drv, "wbem_param", map[string]string{"type": "wbem_params"}) ||
			ExistsInChilren(t, drv, "access_param", map[string]string{"type": "wbem_params"}) {
			t.Error("wbem_params in result")
		}

		checkHistoryRule(t, findOneFrom(t, drv, "metric_trigger", map[string]string{"managed_object_id": id, "name": "rule1"}), rule1)
		checkHistoryRule(t, findOneFrom(t, drv, "metric_trigger", map[string]string{"managed_object_id": id, "name": "rule2"}), rule2)
		checkSnmpParams(t, findOneFrom(t, drv, "snmp_param", map[string]string{"managed_object_id": id, "port": "161"}), snmp_params)
	})
}

func TestIntQueryByIncludesAll(t *testing.T) {
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "network_device", map[string]string{})
		deleteBy(t, client, "network_device_port", map[string]string{})
		deleteBy(t, client, "trigger", map[string]string{})
		deleteBy(t, client, "action", map[string]string{})

		id := create(t, client, "network_device", js)
		drv := findByIdWithIncludes(t, client, "network_device", id, "*")

		checkDevice(t, drv, js)
		bs, e := json.MarshalIndent(drv, "", "  ")
		if nil != e {
			t.Error(e)
		} else {
			t.Log(string(bs))
		}

		checkInterface(t, findOneFrom(t, drv, "network_device_port", map[string]string{"if_index": "1"}), if1)
		checkInterface(t, findOneFrom(t, drv, "network_device_port", map[string]string{"if_index": "9"}), if2)
		if ExistsInChilren(t, drv, "attributes", map[string]string{"type": "wbem_params"}) ||
			ExistsInChilren(t, drv, "attributes", map[string]string{"type": "wbem_params"}) ||
			ExistsInChilren(t, drv, "attributes", map[string]string{"type": "wbem_params"}) {
			t.Error("wbem_params in result")
		}

		checkHistoryRule(t, findOneFrom(t, drv, "metric_trigger", map[string]string{"managed_object_id": id, "name": "rule1"}), rule1)
		checkHistoryRule(t, findOneFrom(t, drv, "metric_trigger", map[string]string{"managed_object_id": id, "name": "rule2"}), rule2)
		checkSnmpParams(t, findOneFrom(t, drv, "attributes", map[string]string{"managed_object_id": id, "port": "161"}), snmp_params)
	})
}

func TestIntQueryByParent(t *testing.T) {
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "network_device", map[string]string{})
		deleteBy(t, client, "network_device_port", map[string]string{})
		deleteBy(t, client, "trigger", map[string]string{})
		deleteBy(t, client, "action", map[string]string{})

		id := create(t, client, "network_device", js)
		interfaces := findByParent(t, client, "network_device", id, "network_device_port", nil)

		d1 := searchBy(interfaces, func(r map[string]interface{}) bool { return fmt.Sprint(r["if_index"]) == "1" })
		d2 := searchBy(interfaces, func(r map[string]interface{}) bool { return fmt.Sprint(r["if_index"]) == "9" })

		checkInterface(t, d1, if1)
		checkInterface(t, d2, if2)

		triggers := findByParent(t, client, "network_device", id, "metric_trigger", nil)

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
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "network_device", map[string]string{})
		deleteBy(t, client, "network_device_port", map[string]string{})
		deleteBy(t, client, "trigger", map[string]string{})
		deleteBy(t, client, "action", map[string]string{})

		id := create(t, client, "network_device", js)

		ifc1 := findOne(t, client, "network_device_port", map[string]string{"device_id": id, "if_index": "1"})
		ifc2 := findOne(t, client, "network_device_port", map[string]string{"device_id": id, "if_index": "9"})

		checkDevice(t, findByChild(t, client, "network_device", "network_device_port", fmt.Sprint(ifc1["id"])), js)
		checkDevice(t, findByChild(t, client, "network_device", "network_device_port", fmt.Sprint(ifc2["id"])), js)

		tr1 := findOne(t, client, "metric_trigger", map[string]string{"managed_object_id": id, "name": "rule1"})
		tr2 := findOne(t, client, "metric_trigger", map[string]string{"managed_object_id": id, "name": "rule2"})

		checkDevice(t, findByChild(t, client, "network_device", "metric_trigger", fmt.Sprint(tr1["id"])), js)
		checkDevice(t, findByChild(t, client, "network_device", "metric_trigger", fmt.Sprint(tr2["id"])), js)
	})
}
