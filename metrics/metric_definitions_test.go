package metrics

import (
	"encoding/json"
	"reflect"
	"testing"
)

var data_for_tests = []struct {
	json              string
	route_definitions []RouteDefinition
}{
	{`[{"level":["system","12"],"name":"interface","method":"get","file":"test1_route.lua","action":{"action":"table","method":"get","oid":"1.3.6.7","schema":"snmp"},"match":[{"method":"equal","arguments":["ss","equal(ss)"]},{"method":"start_with","arguments":["tt","start_with(tt)"]},{"method":"end_with","arguments":["aa","end_with(aa)"]},{"method":"contains","arguments":["cc","contains(cc)"]},{"method":"match","arguments":["aa","match(aa{0, 3})"]}],"categories":["default","safe"]}]`,
		[]RouteDefinition{RouteDefinition{Level: []string{"system", "12"},
			Name:   "interface",
			Method: "get",
			File:   "test1_route.lua",
			Action: map[string]string{
				"action": "table",
				"schema": "snmp",
				"method": "get",
				"oid":    "1.3.6.7"},
			Match: []Filter{Filter{Method: "equal",
				Arguments: []string{"ss", "equal(ss)"}},
				Filter{Method: "start_with",
					Arguments: []string{"tt", "start_with(tt)"}},
				Filter{Method: "end_with",
					Arguments: []string{"aa", "end_with(aa)"}},
				Filter{Method: "contains",
					Arguments: []string{"cc", "contains(cc)"}},
				Filter{Method: "match",
					Arguments: []string{"aa", "match(aa{0, 3})"}}},
			Categories: []string{"default", "safe"}}}}}

func TestEncode(t *testing.T) {
	for i, data := range data_for_tests {
		bytes, e := json.Marshal(data.route_definitions)
		if nil != e {
			t.Errorf("Marshal %d failed -- %s", i, e.Error())
			continue
		}
		if string(bytes) != data.json {
			t.Errorf("encode %d failed, result is not match", i)
			t.Log("      " + data.json)
			t.Log("      " + string(bytes))
		}
	}
}

func TestDecode(t *testing.T) {
	for i, data := range data_for_tests {
		rds := make([]RouteDefinition, 0)
		e := json.Unmarshal([]byte(data.json), &rds)
		if nil != e {
			t.Errorf("Unmarshal %d failed -- %s", i, e.Error())
			continue
		}
		if !reflect.DeepEqual(data.route_definitions, rds) {
			t.Errorf("encode %d failed, result is not match", i)
			t.Logf("      %v", data.route_definitions)
			t.Logf("      %v", rds)
		}
	}
}
