package sampling

import (
	"commons"
	"fmt"
	"testing"
)

type CallFunc func(params MContext) (interface{}, error)

func (call CallFunc) Call(params MContext) (interface{}, error) {
	return call(params)
}

func TestDefaultRouteWithPath(t *testing.T) {
	c1 := func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
		return CallFunc(func(params MContext) (interface{}, error) {
			cc := params.GetStringWithDefault("arg1", "")
			if "cc" != cc {
				t.Error("arg1 != 'cc', actual is '" + cc + "'")
			}
			return "P1", nil
		}), nil
	}

	c2 := func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
		return CallFunc(func(params MContext) (interface{}, error) {
			cc := params.GetStringWithDefault("arg1", "")
			if "" != cc {
				t.Error("arg1 != '', actual is ", cc)
			}
			return "P2", nil
		}), nil
	}

	c3 := func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
		return CallFunc(func(params MContext) (interface{}, error) {
			cc := params.GetStringWithDefault("arg3", "")
			if "aa" != cc {
				t.Error("arg1 != 'aa', actual is '" + cc + "'")
			}
			return "P3", nil
		}), nil
	}

	c4 := func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
		return CallFunc(func(params MContext) (interface{}, error) {
			cc := params.GetStringWithDefault("arg3", "")
			if "bb" != cc {
				t.Error("arg3 != '', actual is ", cc)
			}
			return "P4", nil
		}), nil
	}

	r1, _ := newRouteWithSpec("id1", newRouteSpecWithPaths("get", "route", "", []P{{"p1", "arg1"}}, nil, c1), nil)
	r2, _ := newRouteWithSpec("id2", newRouteSpecWithPaths("get", "route", "", nil, nil, c2), nil)

	r3, _ := newRouteWithSpec("id3", newRouteSpecWithPaths("get", "route", "", []P{{"p1", "arg3"}}, Match().Equals("arg3", "aa").Build(), c3), nil)
	r4, _ := newRouteWithSpec("id4", newRouteSpecWithPaths("get", "route", "", nil, Match().Equals("arg3", "bb").Build(), c4), nil)

	route := &RouterGroup{}
	route.register(r1)
	route.register(r2)
	route.register(r3)
	route.register(r4)

	res, e := route.Invoke([]P{{"p1", "cc"}}, &MockContext{commons.StringMap(map[string]string{})})
	if nil != e {
		t.Error(e)
	}
	if "P1" != res {
		t.Error("result is not equals 'P1', actaul is ", res)
	}
	res, e = route.Invoke([]P{}, &MockContext{commons.StringMap(map[string]string{})})
	if nil != e {
		t.Error(e)
	}
	if "P2" != res {
		t.Error("result is not equals 'P2', actaul is ", res)
	}

	res, e = route.Invoke([]P{{"p1", "aa"}}, &MockContext{commons.StringMap(map[string]string{})})
	if nil != e {
		t.Error(e)
	}
	if "P3" != res {
		t.Error("result is not equals 'P3', actaul is ", res)
	}

	res, e = route.Invoke([]P{}, &context{query_params: map[string]string{"arg3": "bb"}, local: map[string]map[string]interface{}{}, alias: alias_names})
	if nil != e {
		t.Error(e)
	}
	if "P4" != res {
		t.Error("result is not equals 'P4', actaul is ", res)
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

	route := &RouterGroup{}
	for idx, test := range tests {
		r1, _ := newRouteWithSpec("id"+fmt.Sprint(idx), newRouteSpecWithPaths("get", "route", "", nil, Match().Oid(test.oid).Build(),
			func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
				ret_value := params["value"]
				return CallFunc(func(params MContext) (interface{}, error) {
					return ret_value, nil
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
		res, e := route.Invoke([]P{}, &context{query_params: map[string]string{"sys.oid": test.oid}, local: map[string]map[string]interface{}{}, alias: alias_names})
		if nil != e {
			t.Errorf("tests[%v]: %v", idx, e)
		}

		if test.value != res {
			t.Errorf("tests[%v]: excepted value is %v,  actual is %v", idx, test.value, res)
		}
	}
	for idx, test := range tests2 {
		res, e := route.Invoke([]P{}, &context{query_params: map[string]string{"sys.oid": test.oid}, local: map[string]map[string]interface{}{}, alias: alias_names})
		if nil != e {
			t.Errorf("tests2[%v]: %v", idx, e)
		}

		if test.value != res {
			t.Errorf("tests2[%v]: excepted value is %v,  actual is %v", idx, test.value, res)
		}
	}
}
