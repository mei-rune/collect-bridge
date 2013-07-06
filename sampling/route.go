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

func (self *Route) Invoke(params commons.Map) commons.Result {
	return self.invoke.Call(params)
}

func newRouteSpec(name, descr string, match Matchers, call func(rs *RouteSpec, params map[string]interface{}) (Method, error)) *RouteSpec {
	return &RouteSpec{Method: "get",
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

func (self *Routers) Invoke(params commons.Map) commons.Result {
	for _, s := range self.routes {
		matched, e := s.matchers.Match(params, false)
		if nil != e {
			return commons.ReturnWithInternalError(e.Error())
		}

		if matched {
			res := s.invoke.Call(params)
			if res.ErrorCode() == commons.ContinueCode {
				continue
			}
			return res
		}
	}
	return commons.ReturnWithNotAcceptable("not match")
}
