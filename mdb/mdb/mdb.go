package main

import (
	"commons/as"
	"encoding/json"
	"flag"
	"fmt"
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

func mainHandle(rw *web.Context) {
	errFile := "_log_/error.html"
	_, err := os.Stat(errFile)
	if err == nil || os.IsExist(err) {
		content, _ := ioutil.ReadFile(errFile)
		rw.WriteString(string(content))
		return
	}
	rw.WriteString("Hello, World!")
}

func main() {
	flag.Parse()

	if nil != flag.Args() && 0 != len(flag.Args()) {
		flag.Usage()
		return
	}

	svr := web.NewServer()
	svr.Config.Name = "meijing-mdb v1.0"
	svr.Config.Address = *address
	svr.Config.StaticDirectory = *directory
	svr.Config.CookieSecret = *cookies

	svr.Get("/", mainHandle)
	driver, e := mdb.NewMdbDriver(*mgoUrl, *mgoDB, nil)
	if nil != e {
		fmt.Println(e)
		return
	}

	svr.Get("/mdb/([^/]*)/([^/]*)", func(ctx *web.Context, t, id string) {
		ctx.Params["mdb.type"] = t
		ctx.Params["id"] = id

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

	svr.Run()
}
