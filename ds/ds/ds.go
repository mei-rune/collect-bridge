package main

import (
	"commons/as"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/emicklei/go-restful"
	"io/ioutil"
	"mdb"
	"os"
	"web"
)

var (
	address   = flag.String("http", ":7071", "the address of http")
	directory = flag.String("directory", ".", "the static directory of http")
	cookies   = flag.String("cookies", "", "the static directory of http")
	mgoUrl    = flag.String("mgo", "127.0.0.1", "the address of mongo server")
	mgoDB     = flag.String("db", "test", "the db of mongo server")
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

func mainHandle(req *restful.Request, resp *restful.Response) {
	errFile := "_log_/error.html"
	_, err := os.Stat(errFile)
	if err == nil || os.IsExist(err) {

		http.ServeFile(
			resp.ResponseWriter,
			req.Request,
			errFile)

		return
	}
	resp.WriteString("Hello, World!")
}

func main() {
	flag.Parse()
	
	if nil != flag.Args() && 0 != len(flag.Args()) {
		flag.Usage()
		return
	}

	ws := new(restful.WebService)
	ws.Route(ws.GET("/").To(mainHandle))

	svr.Get("/", mainHandle)
	driver, e := mdb.NewMdbDriver(*mgoUrl, *mgoDB, nil)
	if nil != e {
		fmt.Println(e)
		return
	}

	svr.Get("/mdb/{mdb.type}/{id}", func(req *restful.Request, resp *restful.Response) {
		ctx.Params["mdb.type"] = t
		ctx.Params["id"] = id

		obj, e := driver.FindById(ctx.Params)
		if nil != e {
			ctx.Abort(e.Code(), e.Error())
			return
		}
		ctx.Status(getStatus(obj, 200))
		err := json.NewEncoder(ctx).Encode(obj)
		if nil != err {
			ctx.Abort(500, "encode failed, "+err.Error())
		}
	})

	svr.Get("/mdb/([^/]*)/([^/]*)/children/([^/]*)", func(ctx *web.Context, t, id, child string) {
		ctx.Params["mdb.type"] = child
		ctx.Params["id"] = "by_parent"
		ctx.Params["parent_id"] = id
		ctx.Params["parent_type"] = t

		obj, e := driver.Get(ctx.Params)
		if nil != e {
			ctx.Abort(e.Code(), e.Error())
			return
		}
		ctx.Status(getStatus(obj, 200))
		err := json.NewEncoder(ctx).Encode(obj)
		if nil != err {
			ctx.Abort(500, "encode failed, "+err.Error())
		}
	})
	svr.Get("/mdb/([^/]*)/([^/]*)/parent/([^/]*)", func(ctx *web.Context, t, id, child string) {
		ctx.Params["mdb.type"] = child
		ctx.Params["id"] = "by_child"
		ctx.Params["child_id"] = id
		ctx.Params["child_type"] = t

		obj, e := driver.Get(ctx.Params)
		if nil != e {
			ctx.Abort(e.Code(), e.Error())
			return
		}
		ctx.Status(getStatus(obj, 200))
		err := json.NewEncoder(ctx).Encode(obj)
		if nil != err {
			ctx.Abort(500, "encode failed, "+err.Error())
		}
	})
	svr.Put("/mdb/([^/]*)/([^/]*)", func(ctx *web.Context, t, id string) {
		ctx.Params["mdb.type"] = t
		ctx.Params["id"] = id

		txt, err := ioutil.ReadAll(ctx.Request.Body)
		if nil != err {
			ctx.Abort(500, "read body failed - "+err.Error())
			return
		}

		ctx.Params["body"] = string(txt)
		obj, e := driver.Put(ctx.Params)
		if nil != e {
			ctx.Abort(e.Code(), e.Error())
			return
		}
		ctx.Status(getStatus(obj, 200))
		err = json.NewEncoder(ctx).Encode(obj)
		if nil != err {
			ctx.Abort(500, "encode failed, "+err.Error())
		}
	})
	svr.Delete("/mdb/([^/]*)/([^/]*)", func(ctx *web.Context, t, id string) {
		ctx.Params["mdb.type"] = t
		ctx.Params["id"] = id
		obj, err := driver.Delete(ctx.Params)
		if nil != err {
			ctx.Abort(err.Code(), err.Error())
			return
		}
		ctx.Status(getStatus(obj, 200))
		e = json.NewEncoder(ctx).Encode(obj)
		if nil != e {
			ctx.Abort(500, "encode failed, "+e.Error())
		}
	})
	svr.Post("/mdb/([^/]*)", func(ctx *web.Context, t string) {
		ctx.Params["mdb.type"] = t
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
		err = json.NewEncoder(ctx).Encode(obj)
		if nil != err {
			ctx.Abort(500, "encode failed, "+err.Error())
		}
	})

	restful.Add(ws)
	http.ListenAndServe(":8080", nil)
}
