package main

import (
	"commons"
	"metrics"
	"web"
)

func registerMetrics(srv *web.Server, params map[string]string, drvMgr *commons.DriverManager) error {
	driver, e := metrics.NewMetrics(params, drvMgr)
	if nil != e {
		return e
	}
	registerDrivers(srv, "metric", driver.DriverManager)

	drvMgr.Register("metrics", driver)
	return nil
}
