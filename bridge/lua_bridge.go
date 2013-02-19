package main

import (
	"commons"
	"encoding/json"
	"io/ioutil"
	"lua_binding"
	"time"
	"web"
)

func registerLua(svr *web.Server, timeout time.Duration, drvMgr *commons.DriverManager) error {
	driver := lua_binding.NewLuaDriver(timeout, drvMgr)
	drvMgr.Register("lua", driver)
	svr.Get("/lua/([^/]*)/([^/]*)", func(ctx *web.Context, script, id string) { luaGet(driver, ctx, script, id) })
	svr.Put("/lua/([^/]*)/([^/]*)", func(ctx *web.Context, script, id string) { luaPut(driver, ctx, script, id) })
	svr.Delete("/lua/([^/]*)/([^/]*)", func(ctx *web.Context, script, id string) { luaDelete(driver, ctx, script, id) })
	svr.Post("/lua/([^/]*)", func(ctx *web.Context, script string) { luaCreate(driver, ctx, script) })

	return driver.Start()
}

func luaGet(driver commons.Driver, ctx *web.Context, script, id string) {
	ctx.Params["schema"] = script
	ctx.Params["id"] = id

	obj, err := driver.Get(ctx.Params)
	if nil != err {
		ctx.Abort(err.Code(), err.Error())
		return
	}
	ctx.Status(getStatus(obj, 200))
	json.NewEncoder(ctx).Encode(obj)
}

func luaPut(driver commons.Driver, ctx *web.Context, script, id string) {
	ctx.Params["schema"] = script
	ctx.Params["id"] = id

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
	ctx.Status(getStatus(obj, 200))
	json.NewEncoder(ctx).Encode(obj)
}

func luaDelete(driver commons.Driver, ctx *web.Context, script, id string) {
	ctx.Params["schema"] = script
	ctx.Params["id"] = id

	obj, err := driver.Delete(ctx.Params)
	if nil != err {
		ctx.Abort(err.Code(), err.Error())
		return
	}
	if obj {
		ctx.WriteString("OK")
	} else {
		ctx.Abort(500, "FAILED")
	}
}

func luaCreate(driver commons.Driver, ctx *web.Context, script string) {
	ctx.Params["schema"] = script

	txt, err := ioutil.ReadAll(ctx.Request.Body)
	ctx.Params["body"] = string(txt)
	if nil != err {
		ctx.Abort(500, "read body failed - "+err.Error())
		return
	}
	obj, e := driver.Create(ctx.Params)
	if nil != e {
		ctx.Abort(e.Code(), e.Error())
		return
	}

	ctx.Status(getStatus(obj, 201))
	json.NewEncoder(ctx).Encode(obj)
}
