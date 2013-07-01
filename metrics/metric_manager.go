package metrics

import (
	"commons"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type MetricManager struct {
	js         string
	dispatcher *dispatcher
}

func (self *MetricManager) clear() {
	self.js = ""
	self.dispatcher.clear()
}

func (self *MetricManager) Get(params map[string]string) (interface{}, error) {
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

func (self *MetricManager) Put(params map[string]string) (commons.Result, error) {
	j, ok := params["body"]
	if !ok {
		return nil, commons.BodyNotExists
	}
	if "" == j {
		return nil, commons.BodyIsEmpty
	}

	var definition RouteDefinition
	e := json.Unmarshal([]byte(j), &definition)
	if nil != e {
		return nil, fmt.Errorf("Unmarshal body to route_definitions failed -- %s\n%s", e.Error(), j)
	}

	rs, e := NewRoute(&definition)
	if nil != e {
		return nil, errors.New("parse route definitions failed.\n" + e.Error())
	}

	self.dispatcher.register(rs)

	return commons.Return(true), nil
}

func (self *MetricManager) Create(params map[string]string) (commons.Result, error) {
	j, ok := params["body"]
	if !ok {
		return nil, commons.BodyNotExists
	}
	if "" == j {
		return nil, commons.BodyIsEmpty
	}

	routes_definitions := make([]RouteDefinition, 0)
	e := json.Unmarshal([]byte(j), &routes_definitions)
	if nil != e {
		return nil, fmt.Errorf("Unmarshal body to route_definitions failed -- %s\n%s", e.Error(), j)
	}
	ss := make([]string, 0, 10)
	for _, rd := range routes_definitions {
		rs, e := NewRoute(&rd)
		if nil != e {
			ss = append(ss, e.Error())
		} else {
			self.dispatcher.register(rs)
		}
	}

	if 0 != len(ss) {
		self.clear()
		return nil, errors.New("parse route definitions failed.\n" + strings.Join(ss, "\n"))
	}

	return commons.Return(true), nil
}

func (self *MetricManager) Delete(params map[string]string) (bool, error) {
	id, ok := params["id"]
	if !ok {
		return false, commons.IdNotExists
	}
	self.dispatcher.unregister("", id)
	return true, nil
}
