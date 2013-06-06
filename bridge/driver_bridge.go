package main

import (
	"commons"
	"commons/as"
	"encoding/json"
	"io/ioutil"
	"web"
)

func notDefined(name string) string {
	return "'" + name + "' is not found."
}

func getStatus(params map[string]interface{}, default_code int) int {
	code, ok := params["code"]
	if !ok {
		return default_code
	}
	i64, e := as.AsInt(code)
	if nil != e {
		return default_code
	}
	return int(i64)
}

func registerBridge(srv *web.Server, drvMgr *commons.DriverManager) {
	registerDrivers(srv, "bridge", drvMgr)
}

func registerDrivers(svr *web.Server, schema string, drvMgr *commons.DriverManager) {
	svr.Get("/"+schema+"/([^/]*)/([^/]*)", func(ctx *web.Context, name, id string) {
		driver, ok := drvMgr.Connect(name)
		if !ok {
			ctx.Abort(404, notDefined(name))
			return
		}

		ctx.Params["id"] = id
		obj := driver.Get(ctx.Params)
		if obj.HasError() {
			ctx.Abort(obj.ErrorCode(), obj.ErrorMessage())
			return
		}
		ctx.Status(200)
		err := json.NewEncoder(ctx).Encode(obj)
		if nil != err {
			ctx.Abort(500, "encode failed, "+err.Error())
		}
	})
	svr.Put("/"+schema+"/([^/]*)/([^/]*)", func(ctx *web.Context, name, id string) {
		driver, ok := drvMgr.Connect(name)
		if !ok {
			ctx.Abort(404, notDefined(name))
			return
		}
		txt, err := ioutil.ReadAll(ctx.Request.Body)
		if nil != err {
			ctx.Abort(400, "read body failed - "+err.Error())
			return
		}

		ctx.Params["id"] = id
		ctx.Params["body"] = string(txt)
		obj := driver.Put(ctx.Params)
		if obj.HasError() {
			ctx.Abort(obj.ErrorCode(), obj.ErrorMessage())
			return
		}
		ctx.Status(200)
		err = json.NewEncoder(ctx).Encode(obj)
		if nil != err {
			ctx.Abort(500, "encode failed, "+err.Error())
		}
	})
	svr.Delete("/"+schema+"/([^/]*)/([^/]*)", func(ctx *web.Context, name, id string) {

		driver, ok := drvMgr.Connect(name)
		if !ok {
			ctx.Abort(404, notDefined(name))
			return
		}
		ctx.Params["id"] = id
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
	})
	svr.Post("/"+schema+"/([^/]*)", func(ctx *web.Context, name string) {
		driver, ok := drvMgr.Connect(name)
		if !ok {
			ctx.Abort(404, notDefined(name))
			return
		}

		txt, err := ioutil.ReadAll(ctx.Request.Body)
		ctx.Params["body"] = string(txt)
		if nil != err {
			ctx.Abort(500, "read body failed - "+err.Error())
			return
		}

		obj := driver.Create(ctx.Params)
		if obj.HasError() {
			ctx.Abort(obj.ErrorCode(), obj.ErrorMessage())
			return
		}
		ctx.Status(201)
		err = json.NewEncoder(ctx).Encode(obj)
		if nil != err {
			ctx.Abort(500, "encode failed, "+err.Error())
		}
	})
}

func registerDriver(svr *web.Server, drvMgr *commons.DriverManager, schema string, drv commons.Driver) {
	drvMgr.Register(schema, drv)
	svr.Get("/"+schema+"/([^/]*)", func(ctx *web.Context, id string) { drvGet(drv, ctx, id) })
	svr.Put("/"+schema+"/([^/]*)", func(ctx *web.Context, id string) { drvPut(drv, ctx, id) })
	svr.Delete("/"+schema+"/([^/]*)", func(ctx *web.Context, id string) { drvDelete(drv, ctx, id) })
	svr.Post("/"+schema, func(ctx *web.Context) { drvCreate(drv, ctx) })
}

func drvGet(driver commons.Driver, ctx *web.Context, id string) {
	ctx.Params["id"] = id

	obj := driver.Get(ctx.Params)
	if obj.HasError() {
		ctx.Abort(obj.ErrorCode(), obj.ErrorMessage())
		return
	}
	ctx.Status(200)
	err := json.NewEncoder(ctx).Encode(obj)
	if nil != err {
		ctx.Abort(500, "encode failed, "+err.Error())
	}
}

func drvPut(driver commons.Driver, ctx *web.Context, id string) {
	ctx.Params["id"] = id

	txt, err := ioutil.ReadAll(ctx.Request.Body)
	if nil != err {
		ctx.Abort(500, "read body failed - "+err.Error())
		return
	}

	ctx.Params["body"] = string(txt)
	obj := driver.Put(ctx.Params)
	if obj.HasError() {
		ctx.Abort(obj.ErrorCode(), obj.ErrorMessage())
		return
	}
	ctx.Status(200)
	err = json.NewEncoder(ctx).Encode(obj)
	if nil != err {
		ctx.Abort(500, "encode failed, "+err.Error())
	}
}

func drvDelete(driver commons.Driver, ctx *web.Context, id string) {
	ctx.Params["id"] = id

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

func drvCreate(driver commons.Driver, ctx *web.Context) {
	txt, err := ioutil.ReadAll(ctx.Request.Body)
	ctx.Params["body"] = string(txt)
	if nil != err {
		ctx.Abort(500, "read body failed - "+err.Error())
		return
	}
	obj := driver.Create(ctx.Params)
	if obj.HasError() {
		ctx.Abort(obj.ErrorCode(), obj.ErrorMessage())
		return
	}
	ctx.Status(201)
	err = json.NewEncoder(ctx).Encode(obj)
	if nil != err {
		ctx.Abort(500, "encode failed, "+err.Error())
	}
}
