package main

import (
	"commons"
	"snmp"
	"web"
)

func registerSNMPScanner(srv *web.Server, drvMgr *commons.DriverManager) error {
	driver := snmp.NewPingerDriver(drvMgr)
	registerDriver(srv, drvMgr, "s-scan", driver)
	return nil
}
