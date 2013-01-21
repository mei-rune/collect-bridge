package metrics

import (
	"fmt"
)

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
	metrics map[string]*Metric
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{metrics: make(map[string]*Metric)}
}

func (self *Dispatcher) Clear() {
	self.metrics = make(map[string]*Metric)
}

func (self *Dispatcher) GetMetric(name string) (*Metric, bool) {
	m, ok := self.metrics[name]
	return m, ok
}

func (self *Dispatcher) Metrics() map[string]*Metric {
	return self.metrics
}

func (self *Dispatcher) Register(name string, metric *Metric) {
	self.metrics[name] = metric
}

func (self *Dispatcher) Unregister(name string) {
	delete(self.metrics, name)
}
