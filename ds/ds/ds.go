package main

import (
	"commons/as"
	"ds"
	"flag"
	"fmt"
	"github.com/emicklei/go-restful"
	"net/http"
	"os"
	"path"
)

var (
	address    = flag.String("http", ":7071", "the address of http")
	directory  = flag.String("directory", ".", "the static directory of http")
	dbUrl      = flag.String("dburl", "host=127.0.0.1 dbname=test user=postgres password=mfk", "the db url")
	drv        = flag.String("db", "postgres", "the db driver")
	goroutines = flag.Int("connections", 10, "the db connection number")
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
	resp.Write([]byte("Hello, World!"))
}

// http://localhost:8080/static/test.xml
// http://localhost:8080/static/
func staticFromPathParam(req *restful.Request, resp *restful.Response) {
	fmt.Println("get " + path.Join(*directory, req.PathParameter("resource")))
	http.ServeFile(
		resp.ResponseWriter,
		req.Request,
		path.Join(*directory, req.PathParameter("resource")))
}

func main() {
	flag.Parse()

	if nil != flag.Args() && 0 != len(flag.Args()) {
		flag.Usage()
		return
	}
	srv, e := ds.NewServer(*drv, *dbUrl, *goroutines)
	if nil != e {
		fmt.Println(e)
		return
	}
	defer srv.Close()

	ws := new(restful.WebService)
	ws.Route(ws.GET("/").To(mainHandle))
	ws.Route(ws.GET("/static/{resource}").To(staticFromPathParam))

	//ws.Consumes(restful.MIME_XML, restful.MIME_JSON).
	//	Produces(restful.MIME_JSON, restful.MIME_XML) // you can specify this per route as well

	ws.Route(ws.GET("/{type}/{id}").To(srv.FindById).
		Doc("get a object instance").
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.PathParameter("id", "identifier of the instance").DataType("string"))) // on the response

	// ws.Route(ws.GET("/{type}/{id}/parent/{parent-type}/").To(srv.FindById).
	// 	Doc("get  a object instance").
	// 	Param(ws.PathParameter("type", "type of the instance").DataType("string")).
	// 	Param(ws.PathParameter("id", "identifier of the instance").DataType("string")).
	// 	Writes(User{})) // on the response

	// ws.Route(ws.GET("/{type}/{id}/children/{children-type}").To(srv.FindById).
	// 	Doc("get a object instance").
	// 	Param(ws.PathParameter("type", "type of the instance").DataType("string")).
	// 	Param(ws.PathParameter("id", "identifier of the instance").DataType("string")).
	// 	Writes(User{})) // on the response

	ws.Route(ws.POST("/{type}").To(srv.Create).
		Doc("create a object").
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.BodyParameter("object", "representation of a object instance").DataType("main.User"))) // from the request

	ws.Route(ws.PUT("/{type}/{id}").To(srv.UpdateById).
		Doc("update a object").
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.PathParameter("id", "identifier of the instance").DataType("string"))) // from the request

	ws.Route(ws.DELETE("/{type}/{id}").To(srv.DeleteById).
		Doc("delete a object").
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.PathParameter("id", "identifier of the instance").DataType("string")))

	println("[ds] serving files on http://localhost" + *address + "/static from local '" + *directory + "'")
	restful.Add(ws)
	http.ListenAndServe(*address, nil)
}
