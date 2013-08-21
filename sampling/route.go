package sampling

import (
	"commons"
	"errors"
	"fmt"
)

type Route struct {
	definition *RouteDefinition
	id, name   string
	matchers   Matchers
	invoke     Method
}

func (self *Route) Invoke(params MContext) commons.Result {
	return self.invoke.Call(params)
}

func newRouteSpec(method, name, descr string, match Matchers, call func(rs *RouteSpec, params map[string]interface{}) (Method, error)) *RouteSpec {
	return &RouteSpec{Method: method,
		Name:        name,
		Description: descr,
		Author:      "mfk",
		License:     "tpt license",
		Level:       []string{"system", "12"},
		Categories:  []string{"default", "safe"},
		Match:       match,
		Call:        call}
}

func newRouteWithSpec(id string, rs *RouteSpec, params map[string]interface{}) (*Route, error) {
	route := &Route{definition: &RouteDefinition{
		Method:      rs.Method,
		Name:        rs.Name,
		Description: rs.Description,
		Author:      rs.Author,
		License:     rs.License,
		Level:       rs.Level,
		Match:       ToFilters(rs.Match),
		Categories:  rs.Categories},
		id:       id,
		name:     rs.Name,
		matchers: rs.Match}

	if nil == rs.Call {
		return nil, errors.New("the call of spec '" + id + "' is nil.")
	}

	m, e := rs.Call(rs, params)
	if nil != e {
		return nil, errors.New("the call of spec '" + id + "' is make failed, " + e.Error())
	}

	if nil == m {
		return nil, errors.New("the result of call of spec '" + id + "' is nil.")
	}

	route.invoke = m
	return route, nil
}

func NewRoute(rd *RouteDefinition) (*Route, error) {
	rs := &Route{definition: rd,
		id:       rd.File,
		name:     rd.Name,
		matchers: NewMatchers()}

	if nil != rd.Match {
		for i, def := range rd.Match {
			matcher, e := NewMatcher(def.Method, def.Arguments)
			if nil != e {
				return nil, fmt.Errorf("Create matcher %d failed, %v", i, e.Error())
			}
			rs.matchers = append(rs.matchers, matcher)
		}
	}

	return rs, nil
}

type Routers struct {
	routes []*Route
}

func (self *Routers) register(rs *Route) error {
	self.routes = append(self.routes, rs)
	return nil
}

func (self *Routers) unregister(id string) {
	for i, s := range self.routes {
		if nil == s {
			continue
		}

		if s.id == id {
			copy(self.routes[i:], self.routes[i+1:])
			self.routes = self.routes[:len(self.routes)-1]
			break
		}
	}
}

func (self *Routers) clear() {
	self.routes = self.routes[0:0]
}

func (self *Routers) Invoke(params MContext) commons.Result {
	for _, s := range self.routes {
		matched, e := s.matchers.Match(params, false)
		if nil != e {
			return commons.ReturnWithInternalError(e.Error())
		}

		if matched {
			//fmt.Println("invoke", s.id)
			res := s.invoke.Call(params)
			if res.ErrorCode() == commons.ContinueCode {
				continue
			}
			return res
		}
	}
	return commons.ReturnWithNotAcceptable("not match")
}

// type DispatchFunc func(params map[string]string) (map[string]interface{}, commons.RuntimeError)

// var emptyResult = make(map[string]interface{})

// type dispatcherBase struct {
// 	SnmpBase
// 	get, set    DispatchFunc
// 	get_methods map[uint]map[string]DispatchFunc
// 	set_methods map[uint]map[string]DispatchFunc
// }

// func splitsystemOid(oid string) (uint, string) {
// 	if !strings.HasPrefix(oid, "1.3.6.1.4.1.") {
// 		return 0, oid
// 	}
// 	oid = oid[12:]
// 	idx := strings.IndexRune(oid, '.')
// 	if -1 == idx {
// 		u, e := strconv.ParseUint(oid, 10, 0)
// 		if nil != e {
// 			panic(e.Error())
// 		}
// 		return uint(u), ""
// 	}

// 	u, e := strconv.ParseUint(oid[:idx], 10, 0)
// 	if nil != e {
// 		panic(e.Error())
// 	}
// 	return uint(u), oid[idx+1:]
// }

// func (self *dispatcherBase) RegisterGetFunc(oids []string, get DispatchFunc) {
// 	for _, oid := range oids {
// 		main, sub := splitsystemOid(oid)
// 		methods := self.get_methods[main]
// 		if nil == methods {
// 			methods = map[string]DispatchFunc{}
// 			self.get_methods[main] = methods
// 		}
// 		methods[sub] = get
// 	}
// }

// func (self *dispatcherBase) RegisterSetFunc(oids []string, set DispatchFunc) {
// 	for _, oid := range oids {
// 		main, sub := splitsystemOid(oid)
// 		methods := self.set_methods[main]
// 		if nil == methods {
// 			methods = map[string]DispatchFunc{}
// 			self.set_methods[main] = methods
// 		}
// 		methods[sub] = set
// 	}
// }

// func findFunc(oid string, funcs map[uint]map[string]DispatchFunc) DispatchFunc {
// 	main, sub := splitsystemOid(oid)
// 	methods := funcs[main]
// 	if nil == methods {
// 		return nil
// 	}
// 	get := methods[sub]
// 	if nil != get {
// 		return get
// 	}
// 	if "" == sub {
// 		return nil
// 	}
// 	return methods[""]
// }

// func (self *dispatcherBase) FindGetFunc(oid string) DispatchFunc {
// 	return findFunc(oid, self.get_methods)
// }

// func (self *dispatcherBase) FindSetFunc(oid string) DispatchFunc {
// 	return findFunc(oid, self.set_methods)
// }

// func findDefaultFunc(oid string, funcs map[uint]map[string]DispatchFunc) DispatchFunc {
// 	main, sub := splitsystemOid(oid)
// 	methods := funcs[main]
// 	if nil == methods {
// 		return nil
// 	}
// 	if "" == sub {
// 		return nil
// 	}
// 	return methods[""]
// }

// func (self *dispatcherBase) FindDefaultGetFunc(oid string) DispatchFunc {
// 	return findDefaultFunc(oid, self.get_methods)
// }

// func (self *dispatcherBase) FindDefaultSetFunc(oid string) DispatchFunc {
// 	return findDefaultFunc(oid, self.set_methods)
// }

// func (self *dispatcherBase) Init(params map[string]interface{}, drvName string) commons.RuntimeError {
// 	self.get_methods = make(map[uint]map[string]DispatchFunc, 5000)
// 	self.set_methods = make(map[uint]map[string]DispatchFunc, 5000)
// 	return self.SnmpBase.Init(params, drvName)
// }

// func (self *dispatcherBase) invoke(params map[string]string, funcs map[uint]map[string]DispatchFunc) (map[string]interface{}, commons.RuntimeError) {
// 	oid, e := self.GetStringMetric(params, "sys.oid")
// 	if nil != e {
// 		return nil, commons.NewRuntimeError(e.Code(), "get system oid failed, "+e.Error())
// 	}
// 	f := findFunc(oid, funcs)
// 	if nil != f {
// 		res, e := f(params)
// 		if nil == e {
// 			return res, e
// 		}
// 		if commons.ContinueCode != e.Code() {
// 			return res, e
// 		}

// 		f = findDefaultFunc(oid, funcs)
// 		if nil != f {
// 			res, e := f(params)
// 			if nil == e {
// 				return res, e
// 			}
// 			if commons.ContinueCode != e.Code() {
// 				return res, e
// 			}
// 		}
// 	}
// 	if nil != self.get {
// 		return self.get(params)
// 	}
// 	return nil, errutils.NotAcceptable("Unsupported device - " + oid)
// }

// func (self *dispatcherBase) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
// 	return self.invoke(params, self.get_methods)
// }

// func (self *dispatcherBase) Put(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
// 	return self.invoke(params, self.set_methods)
// }
