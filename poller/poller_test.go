package poller

import (
	"commons/types"
	"ds"
	"metrics"
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

func srvTest(t *testing.T, cb func(client *ds.Client, definitions *types.TableDefinitions)) {
	metrics.SrvTest(t, "../ds/etc/mj_models.xml", func(client *ds.Client, definitions *types.TableDefinitions) {
		cb(client, definitions)
	})
}

func TestA(t *testing.T) {
	srvTest(t, func(client *ds.Client, definitions *types.TableDefinitions) {
		id := ds.CreateItForTest(t, client, "network_device", mo)
		ds.CreateItByParentForTest(t, client, "network_device", id, "wbem_param", wbem_params)
		ds.CreateItByParentForTest(t, client, "network_device", id, "snmp_param", snmp_params)


		
	})
}
