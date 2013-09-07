package sampling

import (
	"commons/types"
	ds "data_store"
	"flag"
	"strings"
	"testing"
	"time"
)

var flux_port = flag.String("test.port", "1", "port of test flux")

func TestLinkFlux(t *testing.T) {
	for _, test := range []struct{ device1, snmp1, device2, snmp2 map[string]interface{} }{
		{device1: mo, snmp1: snmp_params,
			device2: copyFrom(mo, map[string]interface{}{"address": "112.23.12.24"}),
			snmp2:   copyFrom(snmp_params, map[string]interface{}{"read_community": "asdfsdfsdfsdf"})}} {
		SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
			_, e := client.DeleteBy("network_device", emptyParams)
			if nil != e {
				t.Error(e)
				return
			}

			id1 := ds.CreateItForTest(t, client, "network_device", test.device1)
			ds.CreateItByParentForTest(t, client, "network_device", id1, "wbem_param", wbem_params)
			ds.CreateItByParentForTest(t, client, "network_device", id1, "snmp_param", test.snmp1)

			id2 := ds.CreateItForTest(t, client, "network_device", test.device2)
			ds.CreateItByParentForTest(t, client, "network_device", id2, "wbem_param", wbem_params)
			ds.CreateItByParentForTest(t, client, "network_device", id2, "snmp_param", test.snmp2)

			link := copyFrom(old_link, map[string]interface{}{"from_device": id1, "to_device": id2, "from_if_index": *flux_port})
			id := ds.CreateItForTest(t, client, "network_link", link)

			res := urlGet(t, sampling_url, "network_link", id, "link_flux")
			if res.HasError() && !strings.Contains(res.ErrorMessage(), "sample is pending.") {
				t.Error(res.Error())
				return
			}
			time.Sleep(1 * time.Second)
			res = urlGet(t, sampling_url, "network_link", id, "link_flux")
			if res.HasError() {
				t.Error(res.Error())
				return
			}

			if nil == res.InterfaceValue() {
				t.Error("values is nil")
			}

			t.Log(res.InterfaceValue())
			_, err := res.Value().AsObject()
			if nil != err {
				t.Error(err)
				return
			}
			// if -1 == commons.GetIntWithDefault(m, "ifType", -1) {
			// 	t.Error("values is error")
			// }
		})
	}
}

func TestInterfaceFluxNative(t *testing.T) {
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, e := client.DeleteBy("network_device", emptyParams)
		if nil != e {
			t.Error(e)
			return
		}
		res := nativeGet(t, sampling_url, "127.0.0.1", "port/"+*flux_port+"/interface_flux", map[string]string{"snmp.version": "v2c", "snmp.read_community": "public"})
		if res.HasError() && !strings.Contains(res.ErrorMessage(), "sample is pending.") {
			t.Error(res.Error())
			return
		}
		time.Sleep(1 * time.Second)
		res = nativeGet(t, sampling_url, "127.0.0.1", "port/"+*flux_port+"/interface_flux", map[string]string{"snmp.version": "v2c", "snmp.read_community": "public"})
		if nil == res.InterfaceValue() {
			t.Error("values is nil")
		}

		t.Log(res.InterfaceValue())
		_, err := res.Value().AsObject()
		if nil != err {
			t.Error(err)
			return
		}
		// if -1 == commons.GetIntWithDefault(m, "ifType", -1) {
		// 	t.Error("values is error")
		// }
	})
}

func TestInterfaceFlux(t *testing.T) {
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
			"if_index": *flux_port,
			"if_descr": "aaa",
			"if_type":  1,
			"if_mtu":   23,
			"if_speed": 23})

		res := urlGet(t, sampling_url, "network_device_port", port_id, "interface_flux")
		if res.HasError() && !strings.Contains(res.ErrorMessage(), "sample is pending.") {
			t.Error(res.Error())
			return
		}
		time.Sleep(1 * time.Second)

		res = urlGet(t, sampling_url, "network_device_port", port_id, "interface_flux")
		if res.HasError() {
			t.Error(res.Error())
			return
		}
		if nil == res.InterfaceValue() {
			t.Error("values is nil")
		}

		t.Log(res.InterfaceValue())
		_, err := res.Value().AsObject()
		if nil != err {
			t.Error(err)
			return
		}
	})
}
