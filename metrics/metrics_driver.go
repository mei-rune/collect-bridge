package metrics

import (
	"commons"
	"time"
)

var MetricNotExists = commons.NewRuntimeError(commons.BadRequestCode, "'metric' is required.")

func MetricNotDefined(name string) commons.RuntimeError {
	return commons.NewRuntimeError(commons.NotFoundCode, "'"+name+"' is not defined.")
}

type Metrics struct {
	*commons.DriverManager
	drvMgr *commons.DriverManager
}

func NewMetrics(ctx map[string]interface{}) (*Metrics, error) {
	metrics := &Metrics{commons.NewDriverManager(), ctx["drvMgr"].(*commons.DriverManager)}
	ctx["metrics"] = metrics.DriverManager

	for k, f := range commons.METRIC_DRVS {
		drv, err := f(ctx)
		if nil != err {
			return nil, err
		}
		metrics.Register(k, drv)
	}

	return metrics, nil
}

func (self *Metrics) Get(params map[string]string) (commons.Result, commons.RuntimeError) {
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

func (self *Metrics) Put(params map[string]string) (commons.Result, commons.RuntimeError) {
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

func (self *Metrics) Create(params map[string]string) (commons.Result, commons.RuntimeError) {
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

func (self *Metrics) Delete(params map[string]string) (commons.Result, commons.RuntimeError) {
	id, ok := params["metric"]
	if !ok {
		return nil, MetricNotExists
	}

	driver, ok := self.Connect(id)
	if !ok {
		return nil, MetricNotDefined(id)
	}
	return driver.Delete(params)
}
