package metrics

import (
	"commons"
	"fmt"
)

type RouteDefinition struct {
	Level      []string          `json:"level"`
	Name       string            `json:"name"`
	Method     string            `json:"method"`
	File       string            `json:"file"`
	Action     map[string]string `json:"action"`
	Match      []Filter          `json:"match"`
	Categories []string          `json:"categories"`
}

type Filter struct {
	Method    string   `json:"method"`
	Arguments []string `json:"arguments"`
}

type Method interface {
	Call(params commons.Map) commons.Result
}

var (
	Methods = map[string]func(params map[string]interface{}) (Method, error){}
)

type RouteSpec struct {
	definition *RouteDefinition
	id, name   string
	matchers   Matchers
	invoke     Method
}

func NewRouteSpec(rd *RouteDefinition) (*RouteSpec, error) {
	rs := &RouteSpec{definition: rd,
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

type Route struct {
	specs []*RouteSpec
}

func (self *Route) registerSpec(rs *RouteSpec) error {
	self.specs = append(self.specs, rs)
	return nil
}

func (self *Route) unregisterSpec(id string) {
	for i, s := range self.specs {
		if nil == s {
			continue
		}

		if s.id == id {
			copy(self.specs[i:], self.specs[i+1:])
			self.specs = self.specs[:len(self.specs)-1]
			break
		}
	}
}

func (self *Route) clear() {
	self.specs = self.specs[0:0]
}

func (self *Route) Invoke(params commons.Map) commons.Result {
	for _, s := range self.specs {
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
