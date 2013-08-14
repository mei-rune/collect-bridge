package sampling

import (
	"commons"
	"snmp"
	"testing"
	"time"
)

type MockContext struct {
	commons.StringMap
}

func (self MockContext) SetBodyClass(value interface{}) {
}

func (self MockContext) Body() (interface{}, error) {
	return self.GetWithDefault("body", nil), nil
}

func TestSysOid(t *testing.T) {
	snmp := snmp.NewSnmpDriver(10*time.Second, nil)
	e := snmp.Start()
	if nil != e {
		t.Error(e)
		return
	}
	defer snmp.Stop()

	var sys systemOid
	e = sys.Init(map[string]interface{}{"snmp": snmp})
	if nil != e {
		t.Error(e)
		return
	}
	res := sys.Call(MockContext{commons.StringMap(map[string]string{"snmp.version": "v2c",
		"@address":            "127.0.0.1",
		"snmp.port":           "161",
		"snmp.read_community": "public"})})
	if res.HasError() {
		t.Error(res.Error())
		return
	}
	s, e := res.Value().AsString()
	if nil != e {
		t.Error(e)
		return
	}
	if 0 == len(s) {
		t.Error("valus is empty.")
	}
	//t.Error(res.InterfaceValue())
}

func TestSysOidFailed(t *testing.T) {
	snmp := snmp.NewSnmpDriver(10*time.Second, nil)
	e := snmp.Start()
	if nil != e {
		t.Error(e)
		return
	}
	defer snmp.Stop()

	var sys systemOid
	e = sys.Init(map[string]interface{}{"snmp": snmp})
	if nil != e {
		t.Error(e)
		return
	}
	res := sys.Call(MockContext{commons.StringMap(map[string]string{})})
	if !res.HasError() {
		t.Error("excepted error is ''snmp' is required.', actual is nil")
		return
	}
	if res.ErrorMessage() != "'snmp' is required." {
		t.Error(res.Error())
		return
	}
}
