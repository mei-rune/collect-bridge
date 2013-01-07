package main

import (
	"commons"
	"encoding/json"
	"io/ioutil"
	"web"
)

func registerSNMP(svr *web.Server) {
	driver := &SnmpDriver{}
	driver.Init()
	driver.Start()
	svr.Get("/snmp/get/(.*)/(.*)", func(ctx *web.Context, host, oid string) { snmpGet(driver, ctx, host, oid, "get") })
	svr.Put("/snmp/set/(.*)/(.*)", func(ctx *web.Context, host, oid string) { snmpPut(driver, ctx, host, oid, "set") })
	svr.Get("/snmp/next/(.*)/(.*)", func(ctx *web.Context, host, oid string) { snmpGet(driver, ctx, host, oid, "next") })
	svr.Get("/snmp/bulk/(.*)/(.*)", func(ctx *web.Context, host, oids string) { snmpGet(driver, ctx, host, oids, "bulk") })
	svr.Get("/snmp/table/(.*)/(.*)", func(ctx *web.Context, host, oid string) { snmpGet(driver, ctx, host, oid, "table") })
	svr.Delete("/snmp/reset/(.*)", func(ctx *web.Context, host string) { snmpReset(driver, ctx, host) })
	svr.Delete("/snmp/reset", func(ctx *web.Context) { snmpReset(driver, ctx, "") })
}

func snmpReset(driver commons.Driver, ctx *web.Context, host string) {
	ctx.Params["remove_clients"] = "true"
	ctx.Params["client"] = host

	ok, err := driver.Delete(ctx.Params)
	if nil != err {
		ctx.Abort(500, err.Error())
		return
	}

	if ok {
		ctx.WriteString("OK")
	} else {
		ctx.Abort(500, "FAILED")
	}
}

func snmpGet(driver commons.Driver, ctx *web.Context, host, oid, action string) {
	//params := make(map[string]string, len(ctx.Params)*3)
	ctx.Params["host"] = host
	ctx.Params["oid"] = oid
	ctx.Params["action"] = action

	obj, err := driver.Get(ctx.Params)
	if nil != err {
		ctx.Abort(500, err.Error())
		return
	}

	json.NewEncoder(ctx).Encode(obj)
}

func snmpPut(driver commons.Driver, ctx *web.Context, host, oid, action string) {
	//params := make(map[string]string, len(ctx.Params)*3)
	ctx.Params["host"] = host
	ctx.Params["oid"] = oid
	ctx.Params["action"] = action
	txt, err := ioutil.ReadAll(ctx.Request.Body)
	ctx.Params["body"] = string(txt)
	if nil != err {
		ctx.Abort(500, "read body failed - "+err.Error())
		return
	}

	obj, err := driver.Put(ctx.Params)
	if nil != err {
		ctx.Abort(500, err.Error())
		return
	}
	json.NewEncoder(ctx).Encode(obj)
}
