package main

import (
	"commons"
	"commons/netutils"
	"web"
)

func registerICMP(srv *web.Server, drvMgr *commons.DriverManager) error {
	driver := netutils.NewICMPDriver(drvMgr)
	registerDriver(srv, drvMgr, "icmp", driver)
	return nil
}
