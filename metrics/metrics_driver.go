package metrics

import (
	"commons"
)

var MetricNotExists = commons.NewRuntimeError(commons.BadRequestCode, "'metric' is required.")

func MetricNotDefined(name string) commons.RuntimeError {
	return commons.NewRuntimeError(commons.NotFoundCode, "'"+name+"' is not defined.")
}

type Metrics struct {
	*commons.DriverManager
	drvMgr *commons.DriverManager
}

func NewMetrics(params map[string]string, drvMgr *commons.DriverManager) (*Metrics, error) {
	metrics := &Metrics{commons.NewDriverManager(), drvMgr}

	for k, f := range commons.METRIC_DRVS {
		drv, err := f(params, drvMgr)
		if nil != err {
			return nil, err
		}
		metrics.Register(k, drv)
	}

	return metrics, nil
}

func (self *Metrics) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	id, ok := params["metric"]
	if !ok {
		return nil, MetricNotExists
	}

	driver, ok := self.Connect(id)
	if !ok {
		return nil, MetricNotDefined(id)
	}
	return driver.Get(params)
}

func (self *Metrics) Put(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	id, ok := params["metric"]
	if !ok {
		return nil, MetricNotExists
	}

	driver, ok := self.Connect(id)
	if !ok {
		return nil, MetricNotDefined(id)
	}
	return driver.Put(params)
}

func (self *Metrics) Create(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	id, ok := params["metric"]
	if !ok {
		return nil, MetricNotExists
	}

	driver, ok := self.Connect(id)
	if !ok {
		return nil, MetricNotDefined(id)
	}
	return driver.Create(params)
}

func (self *Metrics) Delete(params map[string]string) (bool, commons.RuntimeError) {
	id, ok := params["metric"]
	if !ok {
		return false, MetricNotExists
	}

	driver, ok := self.Connect(id)
	if !ok {
		return false, MetricNotDefined(id)
	}
	return driver.Delete(params)
}
