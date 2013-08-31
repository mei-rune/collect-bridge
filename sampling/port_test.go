package sampling

import (
	"commons"
	"commons/types"
	ds "data_store"
	"testing"
)

func TestInterfaceNative(t *testing.T) {
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, e := client.DeleteBy("network_device", emptyParams)
		if nil != e {
			t.Error(e)
			return
		}
		res := nativeGet(t, sampling_url, "127.0.0.1", "port/1/interface", map[string]string{"snmp.version": "v2c", "snmp.read_community": "public"})
		if res.HasError() {
			t.Error(res.Error())
			return
		}

		if nil == res.InterfaceValue() {
			t.Error("values is nil")
		}

		t.Log(res.InterfaceValue())
		m, err := res.Value().AsObject()
		if nil != err {
			t.Error(err)
			return
		}
		if -1 == commons.GetIntWithDefault(m, "ifType", -1) {
			t.Error("values is error")
		}
	})
}

func TestInterface(t *testing.T) {
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, e := client.DeleteBy("network_device", emptyParams)
		if nil != e {
			t.Error(e)
			return
		}

		id := ds.CreateItForTest(t, client, "network_device", mo)
		ds.CreateItByParentForTest(t, client, "network_device", id, "wbem_param", wbem_params)
		ds.CreateItByParentForTest(t, client, "network_device", id, "snmp_param", snmp_params)
		port_id := ds.CreateItByParentForTest(t, client, "network_device", id, "network_device_port", map[string]interface{}{"name": "1",
			"if_index": 1,
			"if_descr": "aaa",
			"if_type":  1,
			"if_mtu":   23,
			"if_speed": 23})

		res := urlGet(t, sampling_url, "network_device_port", port_id, "interface")
		if res.HasError() {
			t.Error(res.Error())
			return
		}
		if nil == res.InterfaceValue() {
			t.Error("values is nil")
		}

		t.Log(res.InterfaceValue())
		m, err := res.Value().AsObject()
		if nil != err {
			t.Error(err)
			return
		}
		if -1 == commons.GetIntWithDefault(m, "ifType", -1) {
			t.Error("values is error")
		}
	})
}
