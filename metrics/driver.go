package metrics

import (
	"commons"
	"commons/types"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"github.com/runner-mei/go-restful"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	_ "runtime/pprof"
	"testing"
	"time"
)

var (
	address = flag.String("http", ":7071", "the address of http")
)

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

func Main() {
	flag.Parse()

	if nil != flag.Args() && 0 != len(flag.Args()) {
		flag.Usage()
		return
	}

	srv, e := NewServer(*drv, *dbUrl, *models_file, *goroutines)
	if nil != e {
		fmt.Println(e)
		return
	}

	defer func() {
		if is_test {
			sinstance = srv
		} else {
			srv.Close()
		}
	}()

	ws := new(restful.WebService)
	ws.Route(ws.GET("/").To(mainHandle))

	ws.Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML) // you can specify this per route as well

	ws.Route(ws.GET("/{type}/{id}/{metric_name}").To(srv.Get).
		Doc("get a metric").
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.PathParameter("id", "identifier of the instance").DataType("string")).
		Param(ws.PathParameter("metric_name", "name of the metric").DataType("string"))) // on the response

	ws.Route(ws.PUT("/{type}/{id}/{metric_name}").To(srv.Put).
		Doc("put a metric").
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.PathParameter("id", "identifier of the instance").DataType("string")).
		Param(ws.PathParameter("metric_name", "name of the metric").DataType("string"))) // on the response

	ws.Route(ws.POST("/{type}/{id}/{metric_name}").To(srv.Create).
		Doc("put a metric").
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.PathParameter("id", "identifier of the instance").DataType("string")).
		Param(ws.PathParameter("metric_name", "name of the metric").DataType("string"))) // on the response

	ws.Route(ws.DELETE("/{type}/{id}/{metric_name}").To(srv.Delete).
		Doc("put a metric").
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.PathParameter("id", "identifier of the instance").DataType("string")).
		Param(ws.PathParameter("metric_name", "name of the metric").DataType("string"))) // on the response

	restful.Add(ws)

	if is_test {
		ws_instance = ws
	} else {
		log.Println("[ds] serving at '" + *address + "'")
		http.ListenAndServe(*address, nil)
	}
}
