package main

import (
	"commons"
	"commons/as"
	"encoding/json"
	"io/ioutil"
	"web"
)

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

func registerDrivers(svr *web.Server, drvMgr *commons.DriverManager) {
	svr.Get("/bridge/(.*)/(.*)", func(ctx *web.Context, drvMgr *commons.DriverManager, name, id string) {
		driversGet(ctx, drvMgr, name, id)
	})
	svr.Put("/bridge/(.*)/(.*)", func(ctx *web.Context, drvMgr *commons.DriverManager, name, id string) {
		driversPut(ctx, drvMgr, name, id)
	})
	svr.Delete("/bridge/(.*)/(.*)", func(ctx *web.Context, drvMgr *commons.DriverManager, name, id string) {
		driversDelete(ctx, drvMgr, name, id)
	})
	svr.Post("/bridge/(.*)", func(ctx *web.Context, drvMgr *commons.DriverManager, name string) {
		driversCreate(ctx, drvMgr, name)
	})
}

func driversGet(ctx *web.Context, drvMgr *commons.DriverManager, name, id string) {
	driver, ok := drvMgr.Connect(name)
	if !ok {
		ctx.Abort(404, "'"+name+"' is not found.")
		return
	}

	ctx.Params["id"] = id
	obj, err := driver.Get(ctx.Params)
	if nil != err {
		ctx.Abort(err.Code(), err.Error())
		return
	}
	ctx.Status(getStatus(obj, 200))
	json.NewEncoder(ctx).Encode(obj)
}

func driversPut(ctx *web.Context, drvMgr *commons.DriverManager, name, id string) {
	driver, ok := drvMgr.Connect(name)
	if !ok {
		ctx.Abort(404, "'"+name+"' is not found.")
		return
	}
	txt, err := ioutil.ReadAll(ctx.Request.Body)
	if nil != err {
		ctx.Abort(400, "read body failed - "+err.Error())
		return
	}

	ctx.Params["id"] = id
	ctx.Params["body"] = string(txt)
	obj, e := driver.Put(ctx.Params)
	if nil != e {
		ctx.Abort(e.Code(), e.Error())
		return
	}
	ctx.Status(getStatus(obj, 200))
	json.NewEncoder(ctx).Encode(obj)
}

func driversDelete(ctx *web.Context, drvMgr *commons.DriverManager, name, id string) {
	driver, ok := drvMgr.Connect(name)
	if !ok {
		ctx.Abort(404, "'"+name+"' is not found.")
		return
	}
	ctx.Params["id"] = id
	obj, err := driver.Delete(ctx.Params)
	if nil != err {
		ctx.Abort(err.Code(), err.Error())
		return
	}

	//ctx.Status(getStatus(obj, 200))
	if obj {
		ctx.Status(200)
		ctx.WriteString("OK")
	} else {
		ctx.Abort(500, "FAILED")
	}
}

func driversCreate(ctx *web.Context, drvMgr *commons.DriverManager, name string) {
	driver, ok := drvMgr.Connect(name)
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

	obj, e := driver.Create(ctx.Params)
	if nil != e {
		ctx.Abort(e.Code(), e.Error())
		return
	}

	ctx.Status(getStatus(obj, 201))
	json.NewEncoder(ctx).Encode(obj)
}

func registerDriver(svr *web.Server, drvMgr *commons.DriverManager, schema string, drv commons.Driver) {
	drvMgr.Register(schema, drv)
	svr.Get("/"+schema+"/(.*)", func(ctx *web.Context, id string) { drvGet(drv, ctx, id) })
	svr.Put("/"+schema+"/(.*)", func(ctx *web.Context, id string) { drvPut(drv, ctx, id) })
	svr.Delete("/"+schema+"/(.*)", func(ctx *web.Context, id string) { drvDelete(drv, ctx, id) })
	svr.Post("/"+schema, func(ctx *web.Context) { drvCreate(drv, ctx) })
}

func drvGet(driver commons.Driver, ctx *web.Context, id string) {
	ctx.Params["id"] = id

	obj, err := driver.Get(ctx.Params)
	if nil != err {
		ctx.Abort(err.Code(), err.Error())
		return
	}
	ctx.Status(getStatus(obj, 200))
	json.NewEncoder(ctx).Encode(obj)
}

func drvPut(driver commons.Driver, ctx *web.Context, id string) {
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

func drvDelete(driver commons.Driver, ctx *web.Context, id string) {
	ctx.Params["id"] = id

	obj, err := driver.Delete(ctx.Params)
	if nil != err {
		ctx.Abort(err.Code(), err.Error())
		return
	}

	//ctx.Status(getStatus(obj, 200))
	if obj {
		ctx.Status(200)
		ctx.WriteString("OK")
	} else {
		ctx.Abort(500, "FAILED")
	}
}

func drvCreate(driver commons.Driver, ctx *web.Context) {
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
