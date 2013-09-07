package sampling

import (
	"commons/types"
	ds "data_store"
	"testing"
)

func TestHostMemNative(t *testing.T) {
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, e := client.DeleteBy("network_device", emptyParams)
		if nil != e {
			t.Error(e)
			return
		}
		res := nativeGet(t, sampling_url, "127.0.0.1", "mem", map[string]string{"snmp.version": "v2c", "snmp.read_community": "public"})
		if res.HasError() {
			t.Error(res.Error())
			return
		}

		if nil == res.InterfaceValue() {
			t.Error("values is nil")
		}
	})
}

func TestHostMem(t *testing.T) {
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, e := client.DeleteBy("network_device", emptyParams)
		if nil != e {
			t.Error(e)
			return
		}

		id := ds.CreateItForTest(t, client, "network_device", mo)
		ds.CreateItByParentForTest(t, client, "network_device", id, "wbem_param", wbem_params)
		ds.CreateItByParentForTest(t, client, "network_device", id, "snmp_param", snmp_params)

		res := urlGet(t, sampling_url, "network_device", id, "mem")
		if res.HasError() {
			t.Error(res.Error())
			return
		}
		if nil == res.InterfaceValue() {
			t.Error("values is nil")
		}
	})
}

func TestH3CMemNative(t *testing.T) {
	for _, version := range []string{"v1", "v2c"} {
		SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
			_, e := client.DeleteBy("network_device", emptyParams)
			if nil != e {
				t.Error(e)
				return
			}
			res := nativeGet(t, sampling_url, "121.0.0.4", "mem", map[string]string{"snmp.version": version, "snmp.read_community": "public"})
			if res.HasError() {
				t.Error(res.Error())
				return
			}

			if nil == res.InterfaceValue() {
				t.Error("values is nil")
			}

			t.Log(res.InterfaceValue())
		})
	}
}

func TestH3CMem(t *testing.T) {
	for _, version := range []string{"v1", "v2c"} {
		SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
			_, e := client.DeleteBy("network_device", emptyParams)
			if nil != e {
				t.Error(e)
				return
			}
			_, e = client.DeleteBy("snmp_param", emptyParams)
			if nil != e {
				t.Error(e)
				return
			}

			id := ds.CreateItForTest(t, client, "network_device", copyFrom(mo, map[string]interface{}{"address": "121.0.0.4"}))
			ds.CreateItByParentForTest(t, client, "network_device", id, "wbem_param", wbem_params)
			ds.CreateItByParentForTest(t, client, "network_device", id, "snmp_param", copyFrom(snmp_params, map[string]interface{}{"version": version, "read_community": "public"}))

			res := urlGet(t, sampling_url, "network_device", id, "mem")
			if res.HasError() {
				t.Error(res.Error())
				return
			}
			if nil == res.InterfaceValue() {
				t.Error("values is nil")
			}
			t.Log(res.InterfaceValue())

		})
	}
}
