package sampling

import (
	"commons"
	"commons/types"
	ds "data_store"
	"strings"
	"testing"
	"time"
)

func TestInterfaceFlux(t *testing.T) {
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

			link := copyFrom(old_link, map[string]interface{}{"from_device": id1, "to_device": id2})
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
}
