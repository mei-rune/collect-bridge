package mdb

import (
	"testing"
)

var (
	snmp_params = map[string]interface{}{"address": "192.168.1.9",
		"community": "public",
		"port":      161,
		"type":      "snmp_params",
		"version":   "v2c"}

	if1 = map[string]interface{}{
		"ifDescr":       "Software Loopback Interface 1",
		"ifIndex":       1,
		"ifMtu":         1500,
		"ifPhysAddress": "",
		"ifSpeed":       1073741824,
		"ifType":        24}
	if2 = map[string]interface{}{
		"ifDescr":       "RAS Async Adapter",
		"ifIndex":       9,
		"ifMtu":         0,
		"ifPhysAddress": "20:41:53:59:4e:ff",
		"ifSpeed":       0,
		"ifType":        23}
	ip1 = map[string]interface{}{
		"address":      "127.0.0.1",
		"bcastAddress": 1,
		"ifIndex":      1,
		"netmask":      "255.0.0.0",
		"reasmMaxSize": 65535}
	ip2 = map[string]interface{}{
		"address":      "169.254.67.142",
		"bcastAddress": 1,
		"ifIndex":      27,
		"netmask":      "255.255.0.0",
		"reasmMaxSize": 65535}
	ip3 = map[string]interface{}{
		"address":      "192.168.1.9",
		"bcastAddress": 1,
		"ifIndex":      13,
		"netmask":      "255.255.255.0",
		"reasmMaxSize": 65535}

	rule1 = map[string]interface{}{"name": "rule1", "type": "metric_rule", "expression": "d111", "metric": "2221"}
	rule2 = map[string]interface{}{"name": "rule2", "type": "metric_rule", "expression": "d222", "metric": "2222"}

	js = map[string]interface{}{
		"$access_params": []map[string]interface{}{snmp_params},
		"$address":       map[string]interface{}{"127.0.0.1": ip1, "169.254.67.142": ip2, "192.168.1.9": ip3},
		"$interface":     map[string]interface{}{"1": if1, "9": if2},
		"$trigger":       []interface{}{rule1, rule2},
		"address":        "192.168.1.9",
		"catalog":        2,
		"description":    "Hardware: Intel64 Family 6 Model 58 Stepping 9 AT/AT COMPATIBLE - Software:Windows Version 6.1 (Build 7601 Multiprocessor Free)",
		"location":       "",
		"name":           "meifakun-PC",
		"oid":            "1.3.6.1.4.1.311.1.1.3.1.1",
		"services":       76}
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
	//{"name": "rule2", "type": "history_rule", "expression": "d222", "metric": "2222"}
	for _, name := range []string{"name", "type", "expression", "metric"} {
		checkStringField(t, actual, excepted, name)
	}
}

func TestIntCreateDevice(t *testing.T) {

	deleteById(t, "device", "all")
	deleteById(t, "interface", "all")
	deleteById(t, "trigger", "all")

	id := create(t, "device", js)

	checkDevice(t, findById(t, "device", id), js)
	checkInterface(t, findOne(t, "interface", map[string]string{"device_id": id, "ifIndex": "1"}), if1)
	checkInterface(t, findOne(t, "interface", map[string]string{"device_id": id, "ifIndex": "9"}), if2)

	checkHistoryRule(t, findOne(t, "trigger", map[string]string{"parent_type": "device", "parent_id": id, "name": "rule1"}), rule1)
	checkHistoryRule(t, findOne(t, "trigger", map[string]string{"parent_type": "device", "parent_id": id, "name": "rule2"}), rule2)

	checkSnmpParams(t, findOne(t, "snmp_params", map[string]string{"device_id": id, "address": "192.168.1.9"}), snmp_params)
}
