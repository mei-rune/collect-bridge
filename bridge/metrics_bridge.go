package main

import (
	"commons"
	"metrics"
	"web"
)

func registerMetrics(srv *web.Server, params map[string]interface{}) error {
	driver, e := metrics.NewMetrics(params)
	if nil != e {
		return e
	}
	registerDrivers(srv, "metric", driver.DriverManager)
	params["drvMgr"].(*commons.DriverManager).Register("metrics", driver)
	return nil
}
