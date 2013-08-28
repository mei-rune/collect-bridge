package sampling

import (
	"commons"
	"fmt"
	"testing"
)

type CallFunc func(params MContext) commons.Result

func (call CallFunc) Call(params MContext) commons.Result {
	return call(params)
}

func TestDefaultRouteWithPath(t *testing.T) {
	c1 := func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
		return CallFunc(func(params MContext) commons.Result {
			cc := params.GetStringWithDefault("arg1", "")
			if "cc" != cc {
				t.Error("arg1 != 'cc', actual is '" + cc + "'")
			}
			return commons.Return("P1")
		}), nil
	}

	c2 := func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
		return CallFunc(func(params MContext) commons.Result {
			cc := params.GetStringWithDefault("arg1", "")
			if "" != cc {
				t.Error("arg1 != '', actual is ", cc)
			}
			return commons.Return("P2")
		}), nil
	}

	c3 := func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
		return CallFunc(func(params MContext) commons.Result {
			cc := params.GetStringWithDefault("arg3", "")
			if "aa" != cc {
				t.Error("arg1 != 'aa', actual is '" + cc + "'")
			}
			return commons.Return("P3")
		}), nil
	}

	c4 := func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
		return CallFunc(func(params MContext) commons.Result {
			cc := params.GetStringWithDefault("arg3", "")
			if "bb" != cc {
				t.Error("arg3 != '', actual is ", cc)
			}
			return commons.Return("P4")
		}), nil
	}

	r1, _ := newRouteWithSpec("id1", newRouteSpecWithPaths("get", "route", "", []P{{"p1", "arg1"}}, nil, c1), nil)
	r2, _ := newRouteWithSpec("id2", newRouteSpecWithPaths("get", "route", "", nil, nil, c2), nil)

	r3, _ := newRouteWithSpec("id3", newRouteSpecWithPaths("get", "route", "", []P{{"p1", "arg3"}}, Match().Equals("arg3", "aa").Build(), c3), nil)
	r4, _ := newRouteWithSpec("id4", newRouteSpecWithPaths("get", "route", "", nil, Match().Equals("arg3", "bb").Build(), c4), nil)

	route := &Routers{}
	route.register(r1)
	route.register(r2)
	route.register(r3)
	route.register(r4)

	res := route.Invoke([]P{{"p1", "cc"}}, &MockContext{commons.StringMap(map[string]string{})})
	if res.HasError() {
		t.Error(res.ErrorMessage())
	}

	if "P1" != res.InterfaceValue() {
		t.Error("result is not equals 'P1', actaul is ", res.InterfaceValue())
	}

	res = route.Invoke([]P{}, &MockContext{commons.StringMap(map[string]string{})})
	if res.HasError() {
		t.Error(res.ErrorMessage())
	}

	if "P2" != res.InterfaceValue() {
		t.Error("result is not equals 'P2', actaul is ", res.InterfaceValue())
	}

	res = route.Invoke([]P{{"p1", "aa"}}, &MockContext{commons.StringMap(map[string]string{})})
	if res.HasError() {
		t.Error(res.ErrorMessage())
	}

	if "P3" != res.InterfaceValue() {
		t.Error("result is not equals 'P3', actaul is ", res.InterfaceValue())
	}

	res = route.Invoke([]P{}, &context{params: map[string]string{"arg3": "bb"}, local: map[string]map[string]interface{}{}, alias: alias_names})
	if res.HasError() {
		t.Error(res.ErrorMessage())
	}

	if "P4" != res.InterfaceValue() {
		t.Error("result is not equals 'P4', actaul is ", res.InterfaceValue())
	}
}

func TestDefaultRouteWithOid(t *testing.T) {
	tests := []struct{ oid, value string }{{oid: "1.3.6.1.4.1.5655", value: "5655"},
		{oid: "1.3.6.1.4.1.9", value: "9"},
		{oid: "1.3.6.1.4.1.9.1.746", value: "9.1.746"},
		{oid: "1.3.6.1.4.1.9.12.3.1.3", value: "9.12.3.1.3"},
		{oid: "1.12.3.1.3", value: "1.12.3.1.3"},
		{oid: "1.12.3.1", value: "1.12.3.1"}}

	tests2 := []struct{ oid, value string }{{oid: "1.3.6.1.4.1.9.1", value: "9"},
		{oid: "1.3.6.1.4.1.9.1.746.2", value: "9.1.746"}}

	route := &Routers{}
	for idx, test := range tests {
		r1, _ := newRouteWithSpec("id"+fmt.Sprint(idx), newRouteSpecWithPaths("get", "route", "", nil, Match().Oid(test.oid).Build(),
			func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
				ret_value := params["value"]
				return CallFunc(func(params MContext) commons.Result {
					return commons.Return(ret_value)
				}), nil
			}), map[string]interface{}{"value": test.value})

		route.register(r1)
	}

	// for k, v := range route.byOid {
	// 	for _, r := range v {
	// 		fmt.Println(k, "=", r.id)
	// 	}
	// }

	for idx, test := range tests {
		res := route.Invoke([]P{}, &context{params: map[string]string{"sys.oid": test.oid}, local: map[string]map[string]interface{}{}, alias: alias_names})
		if res.HasError() {
			t.Errorf("tests[%v]: %v", idx, res.ErrorMessage())
		}

		if test.value != res.InterfaceValue() {
			t.Errorf("tests[%v]: excepted value is %v,  actual is %v", idx, test.value, res.InterfaceValue())
		}
	}
	for idx, test := range tests2 {
		res := route.Invoke([]P{}, &context{params: map[string]string{"sys.oid": test.oid}, local: map[string]map[string]interface{}{}, alias: alias_names})
		if res.HasError() {
			t.Errorf("tests2[%v]: %v", idx, res.ErrorMessage())
		}

		if test.value != res.InterfaceValue() {
			t.Errorf("tests2[%v]: excepted value is %v,  actual is %v", idx, test.value, res.InterfaceValue())
		}
	}
}
