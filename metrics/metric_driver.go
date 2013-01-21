package metrics

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type MetricDefinitionDriver struct {
	js         string
	dispatcher *Dispatcher
}

func (self *MetricDefinitionDriver) Clear() {
	self.js = ""
	self.dispatcher.Clear()
}

func (self *MetricDefinitionDriver) Register(rs *MetricSpec) error {
	metric, _ := self.dispatcher.GetMetric(rs.name)
	if nil == metric {
		metric = &Metric{metric_get: make([]*MetricSpec, 0),
			metric_put:    make([]*MetricSpec, 0),
			metric_create: make([]*MetricSpec, 0),
			metric_delete: make([]*MetricSpec, 0)}

		self.dispatcher.Register(rs.name, metric)
	}

	switch rs.definition.Method {
	case "get":
		metric.metric_get = append(metric.metric_get, rs)
	case "put":
		metric.metric_put = append(metric.metric_put, rs)
	case "create":
		metric.metric_create = append(metric.metric_create, rs)
	case "delete":
		metric.metric_delete = append(metric.metric_delete, rs)
	default:
		return errors.New("Unsupported method - " + rs.definition.Method)
	}
	return nil
}

func deleteSpecFromSlice(specs []*MetricSpec, id string) []*MetricSpec {
	for i, s := range specs {
		if nil == s {
			continue
		}

		if s.id == id {
			copy(specs[i:], specs[i+1:])
			return specs[:len(specs)-1]
		}
	}

	return specs
}

func (self *MetricDefinitionDriver) Unregister(name, id string) {
	if "" == name {
		for _, metric := range self.dispatcher.Metrics() {
			self.unregisterMetric(metric, id)
		}
	} else {
		metric, _ := self.dispatcher.GetMetric(name)
		if nil == metric {
			return
		}
		self.unregisterMetric(metric, id)
	}
}

func (self *MetricDefinitionDriver) unregisterMetric(metric *Metric, id string) {
	if nil != metric.metric_get {
		metric.metric_get = deleteSpecFromSlice(metric.metric_get, id)
	}

	if nil != metric.metric_put {
		metric.metric_put = deleteSpecFromSlice(metric.metric_put, id)
	}

	if nil != metric.metric_create {
		metric.metric_create = deleteSpecFromSlice(metric.metric_create, id)
	}

	if nil != metric.metric_delete {
		metric.metric_delete = deleteSpecFromSlice(metric.metric_delete, id)
	}
}

func (self *MetricDefinitionDriver) Get(params map[string]string) (interface{}, error) {
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

func (self *MetricDefinitionDriver) Put(params map[string]string) (interface{}, error) {

	j, ok := params["body"]
	if !ok {
		return false, errors.New("'body' is required.")
	}
	if "" == j {
		return false, errors.New("'body' is empty.")
	}

	var definition MetricDefinition
	e := json.Unmarshal([]byte(j), &definition)
	if nil != e {
		return false, fmt.Errorf("Unmarshal body to route_definitions failed -- %s\n%s", e.Error(), j)
	}

	rs, e := NewMetricSpec(&definition)
	if nil != e {
		return nil, errors.New("parse route definitions failed.\n" + e.Error())
	}

	self.Register(rs)

	return "ok", nil
}

func (self *MetricDefinitionDriver) Create(params map[string]string) (bool, error) {
	j, ok := params["body"]
	if !ok {
		return false, errors.New("'body' is required.")
	}
	if "" == j {
		return false, errors.New("'body' is empty.")
	}

	routes_definitions := make([]MetricDefinition, 0)
	e := json.Unmarshal([]byte(j), &routes_definitions)
	if nil != e {
		return false, fmt.Errorf("Unmarshal body to route_definitions failed -- %s\n%s", e.Error(), j)
	}
	ss := make([]string, 0, 10)
	for _, rd := range routes_definitions {
		rs, e := NewMetricSpec(&rd)
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

func (self *MetricDefinitionDriver) Delete(params map[string]string) (bool, error) {
	id, ok := params["id"]
	if !ok {
		return false, errors.New("id is required")
	}
	self.Unregister("", id)
	return true, nil
}
