package main

import (
	"commons"
	"encoding/json"
	"io/ioutil"
	"web"
)

func registerDriverBridge(svr *web.Server) {
	svr.Get("/driver/(.*)", func(ctx *web.Context, name string) { driverGet(ctx, name) })
	svr.Put("/driver/(.*)", func(ctx *web.Context, name string) { driverPut(ctx, name) })
	svr.Delete("/driver/(.*)", func(ctx *web.Context, name string) { driverDelete(ctx, name) })
	svr.Post("/driver/(.*)", func(ctx *web.Context, name string) { driverCreate(ctx, name) })
}

func driverGet(ctx *web.Context, name string) {
	driver, ok := commons.Connect(name)
	if !ok {
		ctx.Abort(404, "'"+name+"' is not found.")
		return
	}

	obj, err := driver.Get(ctx.Params)
	if nil != err {
		ctx.Abort(500, err.Error())
		return
	}
	json.NewEncoder(ctx).Encode(obj)
}

func driverPut(ctx *web.Context, name string) {
	driver, ok := commons.Connect(name)
	if !ok {
		ctx.Abort(404, "'"+name+"' is not found.")
		return
	}
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

func driverDelete(ctx *web.Context, name string) {
	driver, ok := commons.Connect(name)
	if !ok {
		ctx.Abort(404, "'"+name+"' is not found.")
		return
	}

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

func driverCreate(ctx *web.Context, name string) {
	driver, ok := commons.Connect(name)
	if !ok {
		ctx.Abort(404, "'"+name+"' is not found.")
		return
	}

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
