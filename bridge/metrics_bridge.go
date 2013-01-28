package main

import (
	"commons"
	"errors"
	"strings"
	"web"
)

func registerMetrics(srv *web.Server, params map[string]string, drvMgr *commons.DriverManager) error {
	metrics := commons.NewDriverManager()

	for k, f := range commons.METRIC_DRVS {
		drv := f(params, drvMgr)
		metrics.Register(k, drv)
		drvMgr.Register("metric/"+k, drv)
	}

	registerDrivers(srv, "metric", metrics)
	return nil
}

func startMetrics(drvMgr *commons.DriverManager) error {
	errs := make([]string, 0)
	for k, _ := range commons.METRIC_DRVS {
		e := drvMgr.Start("metric/" + k)
		if nil != e {
			errs = append(errs, e.Error())
		}
	}
	return errors.New("start metrics failed.\n" + strings.Join(errs, "\n"))
}

func stopMetrics(drvMgr *commons.DriverManager) {
	for k, _ := range commons.METRIC_DRVS {
		drvMgr.Stop("metric/" + k)
	}
}
