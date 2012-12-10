package main

import (
	"commons"
	"encoding/json"
	"io/ioutil"
	"lua_binding"
	"web"
)

func registerLua(svr *web.Server) {
	driver := new(lua_binding.LuaDriver)
	driver.Start()
	svr.Get("/lua/(.*)", func(ctx *web.Context, script string) { luaGet(driver, ctx, script) })
	svr.Put("/lua/(.*)", func(ctx *web.Context, script string) { luaPut(driver, ctx, script) })
	svr.Delete("/lua/(.*)", func(ctx *web.Context, script string) { luaDelete(driver, ctx, script) })
	svr.Post("/lua/(.*)", func(ctx *web.Context, script string) { luaCreate(driver, ctx, script) })
}

func luaGet(driver commons.Driver, ctx *web.Context, script string) {
	ctx.Params["schema"] = script

	obj, err := driver.Get(ctx.Params)
	if nil != err {
		ctx.Abort(500, err.Error())
		return
	}
	json.NewEncoder(ctx).Encode(obj)
}

func luaPut(driver commons.Driver, ctx *web.Context, script string) {
	ctx.Params["schema"] = script
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

func luaDelete(driver commons.Driver, ctx *web.Context, script string) {
	ctx.Params["schema"] = script

	obj, err := driver.Delete(ctx.Params)
	if nil != err {
		ctx.Abort(500, err.Error())
		return
	}
	if obj {
		ctx.WriteString("true")
	} else {
		ctx.WriteString("false")
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
	obj, err := driver.Create(ctx.Params)
	if nil != err {
		ctx.Abort(500, err.Error())
		return
	}
	if obj {
		ctx.WriteString("true")
	} else {
		ctx.WriteString("false")
	}
}
