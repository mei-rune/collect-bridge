package main

import (
	"commons"
	"discovery"
	"time"
	"web"
)

func registerDiscovery(svr *web.Server, timeout time.Duration, drvMgr *commons.DriverManager) error {
	registerDriver(svr, drvMgr, "discovery", discovery.NewDiscoveryDriver(timeout, drvMgr))
	return nil
}
