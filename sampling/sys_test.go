package sampling

import (
	"commons"
	"commons/types"
	ds "data_store"
	"snmp"
	"testing"
)

func TestSystemTypeInit(t *testing.T) {
	tp := &systemType{}
	snmp := snmp.NewSnmpDriver(1000, nil)
	tp.Init(map[string]interface{}{"snmp": snmp, "oid2type": "tests/oid2type.conf"})

	if nil == tp.device2id {
		t.Error("device2id is nil")
		return
	}

	for k, v := range map[string]int{"cc": 5,
		"1.2.3": 3,
		"1.2.4": 4,
		"1.2.5": 5,
		"1.2.6": 6} {
		if v != tp.device2id[k] {
			t.Error("test['", k, "'] ", v)
		}

		ctx := MockContext{commons.StringMap(map[string]string{"$sys.oid": k})}
		res := tp.Call(ctx)
		if res.HasError() {
			t.Error(res.ErrorMessage())
			continue
		}

		if vv, e := res.Value().AsInt(); nil != e || v != vv {
			t.Error("test['", k, "'] v(", v, ") != vv(", vv, ")", e)
		}

		ctx = MockContext{commons.StringMap(map[string]string{"$sys.oid": "." + k})}
		res = tp.Call(ctx)
		if res.HasError() {
			t.Error(res.ErrorMessage())
			continue
		}

		if vv, e := res.Value().AsInt(); nil != e || v != vv {
			t.Error("test['", k, "'] v(", v, ") != vv(", vv, ")", e)
		}
	}
}

func TestSystemTypeNative(t *testing.T) {
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, e := client.DeleteBy("network_device", emptyParams)
		if nil != e {
			t.Error(e)
			return
		}

		res := nativeGet(t, sampling_url, "127.0.0.1", "sys.type", map[string]string{"snmp.version": "v2c", "snmp.read_community": "public"})
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
