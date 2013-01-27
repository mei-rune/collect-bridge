package main

import (
	"commons"
	"encoding/json"
	"io/ioutil"
	"snmp"
	"time"
	"web"
)

func registerSNMP(svr *web.Server, timeout time.Duration, drvMgr *commons.DriverManager) error {
	driver := snmp.NewSnmpDriver(timeout, drvMgr)
	drvMgr.Register("snmp", driver)

	svr.Get("/snmp/(.*)/(.*)/get", func(ctx *web.Context, host, oid string) { snmpGet(driver, ctx, host, oid, "get") })
	svr.Put("/snmp/(.*)/(.*)/put", func(ctx *web.Context, host, oid string) { snmpPut(driver, ctx, host, oid, "set") })

	svr.Get("/snmp/(.*)/(.*)", func(ctx *web.Context, host, oid string) { snmpGet(driver, ctx, host, oid, "get") })
	svr.Put("/snmp/(.*)/(.*)", func(ctx *web.Context, host, oid string) { snmpPut(driver, ctx, host, oid, "set") })

	svr.Get("/snmp/(.*)/(.*)/next", func(ctx *web.Context, host, oid string) { snmpGet(driver, ctx, host, oid, "next") })
	svr.Get("/snmp/(.*)/(.*)/bulk", func(ctx *web.Context, host, oids string) { snmpGet(driver, ctx, host, oids, "bulk") })
	svr.Get("/snmp/(.*)/(.*)/table", func(ctx *web.Context, host, oid string) { snmpGet(driver, ctx, host, oid, "table") })
	svr.Delete("/snmp/(.*)/reset", func(ctx *web.Context, host string) { snmpReset(driver, ctx, host) })
	svr.Delete("/snmp/reset", func(ctx *web.Context) { snmpReset(driver, ctx, "all") })

	return driver.Start()
}

func snmpReset(driver commons.Driver, ctx *web.Context, host string) {
	ctx.Params["action"] = "remove_client"
	ctx.Params["id"] = host

	ok, err := driver.Delete(ctx.Params)
	if nil != err {
		ctx.Abort(err.Code(), err.Error())
		return
	}

	if ok {
		ctx.WriteString("OK")
	} else {
		ctx.Abort(500, "FAILED")
	}
}

func snmpGet(driver commons.Driver, ctx *web.Context, host, oid, action string) {
	ctx.Params["id"] = host + "/" + oid
	ctx.Params["action"] = action

	obj, e := driver.Get(ctx.Params)
	if nil != e {
		ctx.Abort(e.Code(), e.Error())
		return
	}

	json.NewEncoder(ctx).Encode(obj)
}

func snmpPut(driver commons.Driver, ctx *web.Context, host, oid, action string) {
	ctx.Params["id"] = host + "/" + oid
	ctx.Params["action"] = action
	txt, err := ioutil.ReadAll(ctx.Request.Body)
	ctx.Params["body"] = string(txt)
	if nil != err {
		ctx.Abort(500, "read body failed - "+err.Error())
		return
	}

	obj, e := driver.Put(ctx.Params)
	if nil != e {
		ctx.Abort(e.Code(), e.Error())
		return
	}
	json.NewEncoder(ctx).Encode(obj)
}
