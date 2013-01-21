package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type RoutesDriver struct {
	js     string
	routes map[string]map[string]RouteSpec
}

func (self *RoutesDriver) Clear() {
	self.js = ""
	self.routes = make(map[string]map[string]RouteSpec)
}

func (self *RoutesDriver) Register(rs *RouteSpec) {
	container, _ := self.routes[rs.metric]
	if nil == container {
		container = make(map[string]RouteSpec)
		self.routes[rs.metric] = container
	}

	container[rs.id] = rs
}

func (self *RoutesDriver) Unregister(rs *RouteSpec) {
	container, _ := self.routes[rs.metric]
	if nil == container {
		return
	}

	delete(container, rs.id)
}

type RouteSpec struct {
	id, metric string
}

type Filter interface {

		return false, errors.New("parse route definitions failed.\n" + strings.Join(ss, "\n"))
}

func NewRouteSpec(rd *RouteDefinition) (*RouteSpec, error) {

}

func (self *RoutesDriver) Get(params map[string]string) (interface{}, error) {
	t, ok := params["id"]
	if !ok {
		t = "definitions"
	}
	switch t {
	case "definitions":
		return self.js, nil
	}
	return nil, errors.New("not implemented")
}

func (self *RoutesDriver) Put(params map[string]string) (interface{}, error) {

	j, ok := params["body"]
	if !ok {
		return false, errors.New("'body' is required.")
	}
	if "" == j {
		return false, errors.New("'body' is empty.")
	}

	var definition RouteDefinition
	e := json.Unmarshal([]byte(j), &definition)
	if nil != e {
		return false, fmt.Errorf("Unmarshal body to route_definitions failed -- %s\n%s", e.Error(), j)
	}

	rs, e := NewRouteSpec(definition)
	if nil != e {
		return nil, errors.New("parse route definitions failed.\n" + e.Error())
	}

	self.Register(rs)

	return "ok", nil
}

func (self *RoutesDriver) Create(params map[string]string) (bool, error) {
	j, ok := params["body"]
	if !ok {
		return false, errors.New("'body' is required.")
	}
	if "" == j {
		return false, errors.New("'body' is empty.")
	}

	routes_definitions := make([]RouteDefinition, 0)
	e := json.Unmarshal([]byte(j), &routes_definitions)
	if nil != e {
		return false, fmt.Errorf("Unmarshal body to route_definitions failed -- %s\n%s", e.Error(), j)
	}
	ss := make([]string, 0, 10)
	for i, rd := range routes_definitions {
		rs, e := NewRouteSpec(rd)
		if nil != e {
			ss = append(ss, e.Error())
		} else {
			self.Register(rs)
		}
	}

	if 0 != len(ss) {
		self.Clear()
		return false, errors.New("parse route definitions failed.\n" + strings.Join(ss, "\n"))
	}

	return true, nil
}

func (self *RoutesDriver) Delete(params map[string]string) (bool, error) {
	id, ok := params["id"]
	if !ok {
		return false, errors.New("id is required")
	}
	for _, c := range self.routes {
		if nil == c {
			continue
		}

		delete(c, id)
	}
	return true, nil
}
