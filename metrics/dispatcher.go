package metrics

import (
	"commons"
	"errors"
	"flag"
	"fmt"
)

var route_debuging = flag.Bool("route.debugging", false, "set max size of pdu")

func NewMetricSpec(rd *MetricDefinition) (*MetricSpec, error) {
	rs := &MetricSpec{definition: rd,
		id:       rd.File,
		name:     rd.Name,
		matchers: NewMatchers()}

	for i, def := range rd.Match {
		matcher, e := NewMatcher(def.Method, def.Arguments)
		if nil != e {
			return nil, fmt.Errorf("Create matcher %d failed, %v", i, e.Error())
		}
		rs.matchers = append(rs.matchers, matcher)
	}

	return rs, nil
}

type Metric struct {
	metric_get    []*MetricSpec
	metric_put    []*MetricSpec
	metric_create []*MetricSpec
	metric_delete []*MetricSpec
}

type MetricSpec struct {
	definition *MetricDefinition
	id, name   string
	matchers   Matchers
}

type Dispatcher struct {
	instances map[string]*Metric
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{instances: make(map[string]*Metric)}
}

func (self *Dispatcher) registerSpec(rs *MetricSpec) error {
	metric, _ := self.instances[rs.name]
	if nil == metric {
		metric = &Metric{metric_get: make([]*MetricSpec, 0),
			metric_put:    make([]*MetricSpec, 0),
			metric_create: make([]*MetricSpec, 0),
			metric_delete: make([]*MetricSpec, 0)}
		self.instances[rs.name] = metric
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

func deleteSpec(metric *Metric, id string) {
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

func (self *Dispatcher) unregisterSpec(name, id string) {
	if "" == name {
		for _, metric := range self.instances {
			deleteSpec(metric, id)
		}
	} else {
		metric, _ := self.instances[name]
		if nil == metric {
			return
		}
		deleteSpec(metric, id)
	}
}

func (self *Dispatcher) clear() {
	self.instances = make(map[string]*Metric)
}

func (self *Dispatcher) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	// id, ok := params["id"]
	// if ok {
	//	return nil, commons.IdNotExists
	// }
	// metric, ok := self.instances[id]
	// if nil == metric || nil == metric.metric_get || 0 == len(metric.metric_get) {
	//	return nil, errutils.NotAcceptable(id)
	// }

	// error := make([]string, 0)
	// for _, spec := range metric.metric_get {
	//	value, e := spec.Invoke(params)
	//	if nil != e {
	//		error = append(error, e.Error())
	//	} else if nil != value {
	//		return value, nil
	//	}
	// }
	// return nil, errutils.InternalError("get failed.\n" + strings.Join(error, "\n"))

	return nil, commons.NotImplemented
}

func (self *Dispatcher) Put(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return nil, commons.NotImplemented
}

func (self *Dispatcher) Create(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return nil, commons.NotImplemented
}

func (self *Dispatcher) Delete(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return nil, commons.NotImplemented
}
