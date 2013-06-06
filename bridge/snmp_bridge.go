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

	svr.Get("/snmp/([^/]*)/([^/]*)/get", func(ctx *web.Context, host, oid string) { snmpGet(driver, ctx, host, oid, "get") })
	svr.Put("/snmp/([^/]*)/([^/]*)/put", func(ctx *web.Context, host, oid string) { snmpPut(driver, ctx, host, oid, "set") })

	svr.Get("/snmp/([^/]*)/([^/]*)", func(ctx *web.Context, host, oid string) { snmpGet(driver, ctx, host, oid, "get") })
	svr.Put("/snmp/([^/]*)/([^/]*)", func(ctx *web.Context, host, oid string) { snmpPut(driver, ctx, host, oid, "set") })

	svr.Get("/snmp/([^/]*)/([^/]*)/next", func(ctx *web.Context, host, oid string) { snmpGet(driver, ctx, host, oid, "next") })
	svr.Get("/snmp/([^/]*)/([^/]*)/bulk", func(ctx *web.Context, host, oids string) { snmpGet(driver, ctx, host, oids, "bulk") })
	svr.Get("/snmp/([^/]*)/([^/]*)/table", func(ctx *web.Context, host, oid string) { snmpGet(driver, ctx, host, oid, "table") })
	svr.Delete("/snmp/([^/]*)/reset", func(ctx *web.Context, host string) { snmpReset(driver, ctx, host) })
	svr.Delete("/snmp/reset", func(ctx *web.Context) { snmpReset(driver, ctx, "all") })

	return driver.Start()
}

func snmpReset(driver commons.Driver, ctx *web.Context, host string) {
	ctx.Params["snmp.action"] = "remove_client"
	ctx.Params["snmp.host"] = host
	ctx.Params["id"] = host

	obj := driver.Delete(ctx.Params)
	if obj.HasError() {
		ctx.Abort(obj.ErrorCode(), obj.ErrorMessage())
		return
	}

	ctx.Status(200)
	e := json.NewEncoder(ctx).Encode(obj)
	if nil != e {
		ctx.Abort(500, "encode failed, "+e.Error())
	}
}

func snmpGet(driver commons.Driver, ctx *web.Context, host, oid, action string) {
	ctx.Params["id"] = host
	ctx.Params["snmp.host"] = host
	ctx.Params["snmp.oid"] = oid
	ctx.Params["snmp.action"] = action

	obj := driver.Get(ctx.Params)
	if obj.HasError() {
		ctx.Abort(obj.ErrorCode(), obj.ErrorMessage())
		return
	}

	json.NewEncoder(ctx).Encode(obj)
}

func snmpPut(driver commons.Driver, ctx *web.Context, host, oid, action string) {
	ctx.Params["id"] = host
	ctx.Params["snmp.host"] = host
	ctx.Params["snmp.oid"] = oid

	txt, err := ioutil.ReadAll(ctx.Request.Body)
	ctx.Params["body"] = string(txt)
	if nil != err {
		ctx.Abort(500, "read body failed - "+err.Error())
		return
	}

	obj := driver.Put(ctx.Params)
	if obj.HasError() {
		ctx.Abort(obj.ErrorCode(), obj.ErrorMessage())
		return
	}

	json.NewEncoder(ctx).Encode(obj)
}
