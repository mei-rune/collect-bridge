package sampling

import (
	"commons"
	"errors"
	"fmt"
	"strings"
)

type Route struct {
	definition *RouteDefinition
	id, name   string
	matchers   Matchers
	invoke     Method
}

func (self *Route) Match(skipped int, paths []P, params MContext) (bool, error) {
	if nil == self.definition.Paths || 0 == len(self.definition.Paths) {
		if nil != paths && 0 != len(paths) {
			//fmt.Println(self.id, "notMatch, is nil", len(paths), paths)
			return false, nil
		}

		return self.matchers.Match(skipped, nil, params, false)
	}

	if len(paths) != len(self.definition.Paths) {
		//fmt.Println(self.id, "notMatch, count is not match", len(paths), len(self.definition.Paths))
		return false, nil
	}

	for i := 0; i < len(paths); i++ {
		if paths[i][0] != self.definition.Paths[i][0] {
			//fmt.Println(self.id, "notMatch, path is not match", paths[i][0], self.definition.Paths[i][0])
			return false, nil
		}
	}

	path_params := map[string]string{}
	for i, ss := range self.definition.Paths {
		if 0 == len(ss[1]) {
			path_params[ss[0]] = paths[i][1]
		} else {
			path_params[ss[1]] = paths[i][1]
		}
	}
	return self.matchers.Match(skipped, path_params, params, false)
}

func (self *Route) Invoke(paths []P, params MContext) commons.Result {
	if nil != self.definition.Paths && 0 != len(self.definition.Paths) {
		for i, ss := range self.definition.Paths {
			if 0 == len(ss[1]) {
				params.Set(ss[0], paths[i][1])
			} else {
				params.Set(ss[1], paths[i][1])
			}
		}
	}
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

func newRouteSpecWithPaths(method, name, descr string, paths []P, match Matchers, call func(rs *RouteSpec, params map[string]interface{}) (Method, error)) *RouteSpec {
	return &RouteSpec{Method: method,
		Name:        name,
		Description: descr,
		Author:      "mfk",
		License:     "tpt license",
		Level:       []string{"system", "12"},
		Categories:  []string{"default", "safe"},
		Paths:       paths,
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
		Paths:       rs.Paths,
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
	byOid          RouteSetByOid
	routes         []*Route
	default_routes []*Route
}

func (self *Routers) register(rs *Route) error {
	if nil == rs.matchers || 0 == len(rs.matchers) {
		self.default_routes = append(self.default_routes, rs)
		return nil
	}

	match := rs.matchers[0]
	if "start_with" == match.Method &&
		"sys.oid" == match.Attribute {
		if nil == self.byOid {
			self.byOid = RouteSetByOid{}
		}
		self.byOid.register(match.Arguments[0], rs)
		return nil
	}

	self.routes = append(self.routes, rs)
	return nil
}

func (self *Routers) unregister(id string) {
	if nil != self.byOid {
		self.byOid.unregister(id)
	}

	for _, routes := range [][]*Route{self.routes, self.default_routes} {
		if nil == routes || 0 == len(routes) {
			continue
		}

		for i, rs := range routes {
			if nil == rs {
				continue
			}

			if rs.id == id {
				copy(routes[i:], routes[i+1:])
				routes = routes[:len(routes)-1]
				break
			}
		}
	}
}

func (self *Routers) clear() {
	self.byOid = nil
	self.routes = nil
	self.default_routes = nil
}

func (self *Routers) Invoke(paths []P, params MContext) commons.Result {
	if nil != self.byOid {
		oid, e := params.GetString("$sys.oid")
		if nil != e {
			return commons.ReturnWithBadRequest("read 'sys.oid' failed, " + e.Error())
		}
		if 0 == len(oid) {
			return commons.ReturnWithIsRequired("sys.oid")
		}
		route, e := self.byOid.find(oid, paths, params)
		if nil != e {
			return commons.ReturnWithInternalError(e.Error())
		}

		if nil != route {
			return route.Invoke(paths, params)
		}
	}

	if nil != self.routes {
		for _, rs := range self.routes {
			matched, e := rs.Match(0, paths, params)
			if nil != e {
				return commons.ReturnWithInternalError(e.Error())
			}

			if matched {
				return rs.Invoke(paths, params)
			}
		}
	}

	if nil != self.default_routes {
		var res commons.Result
		for _, rs := range self.default_routes {
			if matched, _ := rs.Match(0, paths, params); !matched {
				continue
			}

			res = rs.Invoke(paths, params)
			if res.HasError() {
				continue
			}
			return res
		}
		if nil != res {
			return res
		}
	}
	return commons.ReturnWithNotAcceptable("not match")
}

type RouteSetByOid map[string][]*Route

func normalizeSystemOid(oid string) string {
	if '.' == oid[0] {
		oid = oid[1:]
	}

	if !strings.HasPrefix(oid, "1.3.6.1.4.1.") {
		return "@" + oid
	}

	return oid[12:]
}

func splitOid(oid string) (res []int) {
	for i, c := range oid {
		if '.' == c {
			res = append(res, i)
		}
	}
	res = append(res, len(oid))
	return
}

func (self RouteSetByOid) register(oid string, route *Route) {
	oid = normalizeSystemOid(oid)
	if routes, ok := self[oid]; ok {
		self[oid] = append(routes, route)
	} else {
		self[oid] = []*Route{route}
	}
}

func (self RouteSetByOid) unregister(id string) {
	for k, routes := range self {
		for i, rs := range routes {
			if rs.id == id {
				copy(routes[i:], routes[i+1:])
				self[k] = routes[:len(routes)-1]
				return
			}
		}
	}
}

func (self RouteSetByOid) find(oid string, paths []P, params MContext) (*Route, error) {
	oid = normalizeSystemOid(oid)
	positions := splitOid(oid)
	for i := len(positions) - 1; i >= 0; i-- {
		routes := self[oid[:positions[i]]]
		if nil == routes {
			continue
		}

		for _, rs := range routes {
			matched, e := rs.Match(1, paths, params)
			if nil != e {
				return nil, e
			}
			if matched {
				return rs, nil
			}
		}
	}
	return nil, nil
}
