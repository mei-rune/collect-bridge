package sampling

import (
	_ "expvar"
	"flag"
	"fmt"
	"github.com/runner-mei/go-restful"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	_ "runtime/pprof"
	"snmp"
	"time"
)

var (
	address      = flag.String("sampling.listen", ":7072", "the address of http")
	ds_url       = flag.String("ds.url", "http://127.0.0.1:7071", "the address of http")
	refresh      = flag.Duration("ds.refresh", 60*time.Second, "the duration of refresh")
	snmp_timeout = flag.Duration("snmp.timeout", 60*time.Second, "the timeout duration of snmp")

	is_test                           = false
	srv_instance  *server             = nil
	wsrv_instance *restful.WebService = nil
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

	snmp := snmp.NewSnmpDriver(*snmp_timeout, nil)
	e := snmp.Start()
	if nil != e {
		fmt.Println("start snmp failed,", e)
		return
	}

	srv, e := newServer(*ds_url, *refresh, map[string]interface{}{"snmp": snmp})
	if nil != e {
		fmt.Println("init server failed,", e)
		return
	}

	restful.DefaultResponseMimeType = restful.MIME_JSON
	ws := new(restful.WebService)
	if is_test {
		ws.Path("metrics")
	}
	ws.Route(ws.GET("/").To(mainHandle))

	ws.Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON) // you can specify this per route as well

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
		Doc("create a metric").
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.PathParameter("id", "identifier of the instance").DataType("string")).
		Param(ws.PathParameter("metric_name", "name of the metric").DataType("string"))) // on the response

	ws.Route(ws.DELETE("/{type}/{id}/{metric_name}").To(srv.Delete).
		Doc("delete a metric").
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.PathParameter("id", "identifier of the instance").DataType("string")).
		Param(ws.PathParameter("metric_name", "name of the metric").DataType("string"))) // on the response

	ws.Route(ws.GET("/{ip}/{metric_name}").To(srv.NativeGet).
		Doc("get a metric").
		Param(ws.PathParameter("ip", "ip of the instance").DataType("string")).
		Param(ws.PathParameter("metric_name", "name of the metric").DataType("string"))) // on the response

	ws.Route(ws.PUT("/{ip}/{metric_name}").To(srv.NativePut).
		Doc("put a metric").
		Param(ws.PathParameter("ip", "ip of the instance").DataType("string")).
		Param(ws.PathParameter("metric_name", "name of the metric").DataType("string"))) // on the response

	ws.Route(ws.POST("/{ip}/{metric_name}").To(srv.NativeCreate).
		Doc("create a metric").
		Param(ws.PathParameter("ip", "ip of the instance").DataType("string")).
		Param(ws.PathParameter("metric_name", "name of the metric").DataType("string"))) // on the response

	ws.Route(ws.DELETE("/{ip}/{metric_name}").To(srv.NativeDelete).
		Doc("delete a metric").
		Param(ws.PathParameter("ip", "ip of the instance").DataType("string")).
		Param(ws.PathParameter("metric_name", "name of the metric").DataType("string"))) // on the response

	restful.Add(ws)

	if is_test {
		wsrv_instance = ws
		srv_instance = srv
		log.Println("[sampling-test] serving at '" + *address + "'")
	} else {
		log.Println("[sampling] serving at '" + *address + "'")
		http.ListenAndServe(*address, nil)
	}
}
