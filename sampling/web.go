package sampling

import (
	"expvar"
	"flag"
	"fmt"
	"github.com/runner-mei/go-restful"
	"log"
	"net/http"
	pprof "net/http/pprof"
	"os"
	"snmp"
	"time"
)

var (
	address      = flag.String("sampling.listen", ":7072", "the address of http")
	ds_url       = flag.String("ds.url", "http://127.0.0.1:7071", "the address of http")
	refresh      = flag.Duration("ds.refresh", 60*time.Second, "the duration of refresh")
	snmp_timeout = flag.Duration("snmp.timeout", 60*time.Second, "the timeout duration of snmp")

	Container    *restful.Container = restful.DefaultContainer
	is_test                         = false
	srv_instance *server            = nil
)

func expvarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
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
		ws.Path("/metrics")
	}
	ws.Route(ws.GET("/").To(mainHandle))

	ws.Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON) // you can specify this per route as well

	ws.Route(ws.GET("/{type}/{id}/{metric-name}").To(srv.Get).
		Doc("get a metric").
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.PathParameter("id", "identifier of the instance").DataType("string")).
		Param(ws.PathParameter("metric-name", "name of the metric").DataType("string"))) // on the response

	// ws.Route(ws.GET("/{type}/{id}/{metric-name}/{metric-params}").To(srv.Get).
	// 	Doc("get a metric").
	// 	Param(ws.PathParameter("type", "type of the instance").DataType("string")).
	// 	Param(ws.PathParameter("id", "identifier of the instance").DataType("string")).
	// 	Param(ws.PathParameter("metric-name", "name of the metric").DataType("string")).
	// 	Param(ws.PathParameter("metric-params", "params of the metric").DataType("string"))) // on the response

	ws.Route(ws.PUT("/{type}/{id}/{metric-name}").To(srv.Put).
		Doc("put a metric").
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.PathParameter("id", "identifier of the instance").DataType("string")).
		Param(ws.PathParameter("metric-name", "name of the metric").DataType("string"))) // on the response

	// ws.Route(ws.PUT("/{type}/{id}/{metric-name}/{metric-params}").To(srv.Put).
	// 	Doc("put a metric").
	// 	Param(ws.PathParameter("type", "type of the instance").DataType("string")).
	// 	Param(ws.PathParameter("id", "identifier of the instance").DataType("string")).
	// 	Param(ws.PathParameter("metric-name", "name of the metric").DataType("string")).
	// 	Param(ws.PathParameter("metric-params", "params of the metric").DataType("string"))) // on the response

	ws.Route(ws.POST("/{type}/{id}/{metric-name}").To(srv.Create).
		Doc("create a metric").
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.PathParameter("id", "identifier of the instance").DataType("string")).
		Param(ws.PathParameter("metric-name", "name of the metric").DataType("string"))) // on the response

	// ws.Route(ws.POST("/{type}/{id}/{metric-name}/{metric-params}").To(srv.Create).
	// 	Doc("create a metric").
	// 	Param(ws.PathParameter("type", "type of the instance").DataType("string")).
	// 	Param(ws.PathParameter("id", "identifier of the instance").DataType("string")).
	// 	Param(ws.PathParameter("metric-name", "name of the metric").DataType("string")).
	// 	Param(ws.PathParameter("metric-params", "params of the metric").DataType("string"))) // on the response

	ws.Route(ws.DELETE("/{type}/{id}/{metric-name}").To(srv.Delete).
		Doc("delete a metric").
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.PathParameter("id", "identifier of the instance").DataType("string")).
		Param(ws.PathParameter("metric-name", "name of the metric").DataType("string"))) // on the response

	// ws.Route(ws.DELETE("/{type}/{id}/{metric-name}/{metric-params}").To(srv.Delete).
	// 	Doc("delete a metric").
	// 	Param(ws.PathParameter("type", "type of the instance").DataType("string")).
	// 	Param(ws.PathParameter("id", "identifier of the instance").DataType("string")).
	// 	Param(ws.PathParameter("metric-name", "name of the metric").DataType("string")).
	// 	Param(ws.PathParameter("metric-params", "params of the metric").DataType("string"))) // on the response

	ws.Route(ws.GET("/{ip}/{metric-name}").To(srv.NativeGet).
		Doc("get a metric").
		Param(ws.PathParameter("ip", "ip of the instance").DataType("string")).
		Param(ws.PathParameter("metric-name", "name of the metric").DataType("string"))) // on the response

	// ws.Route(ws.GET("/{ip}/{metric-name}/{metric-params}").To(srv.NativeGet).
	// 	Doc("get a metric").
	// 	Param(ws.PathParameter("ip", "ip of the instance").DataType("string")).
	// 	Param(ws.PathParameter("metric-name", "name of the metric").DataType("string"))) // on the response

	ws.Route(ws.PUT("/{ip}/{metric-name}").To(srv.NativePut).
		Doc("put a metric").
		Param(ws.PathParameter("ip", "ip of the instance").DataType("string")).
		Param(ws.PathParameter("metric-name", "name of the metric").DataType("string"))) // on the response

	// ws.Route(ws.PUT("/{ip}/{metric-name}/{metric-params}").To(srv.NativePut).
	// 	Doc("put a metric").
	// 	Param(ws.PathParameter("ip", "ip of the instance").DataType("string")).
	// 	Param(ws.PathParameter("metric-name", "name of the metric").DataType("string")).
	// 	Param(ws.PathParameter("metric-params", "params of the metric").DataType("string"))) // on the response

	ws.Route(ws.POST("/{ip}/{metric-name}").To(srv.NativeCreate).
		Doc("create a metric").
		Param(ws.PathParameter("ip", "ip of the instance").DataType("string")).
		Param(ws.PathParameter("metric-name", "name of the metric").DataType("string"))) // on the response

	// ws.Route(ws.POST("/{ip}/{metric-name}/{metric-params}").To(srv.NativeCreate).
	// 	Doc("create a metric").
	// 	Param(ws.PathParameter("ip", "ip of the instance").DataType("string")).
	// 	Param(ws.PathParameter("metric-name", "name of the metric").DataType("string")).
	// 	Param(ws.PathParameter("metric-params", "params of the metric").DataType("string"))) // on the response

	ws.Route(ws.DELETE("/{ip}/{metric-name}").To(srv.NativeDelete).
		Doc("delete a metric").
		Param(ws.PathParameter("ip", "ip of the instance").DataType("string")).
		Param(ws.PathParameter("metric-name", "name of the metric").DataType("string"))) // on the response

	// ws.Route(ws.DELETE("/{ip}/{metric-name}/{metric-params}").To(srv.NativeDelete).
	// 	Doc("delete a metric").
	// 	Param(ws.PathParameter("ip", "ip of the instance").DataType("string")).
	// 	Param(ws.PathParameter("metric-name", "name of the metric").DataType("string")).
	// 	Param(ws.PathParameter("metric-params", "params of the metric").DataType("string"))) // on the response

	Container.Add(ws)

	if restful.DefaultContainer != Container {
		Container.Handle("/debug/vars", http.HandlerFunc(expvarHandler))
		Container.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		Container.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		Container.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		Container.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	}

	if is_test {
		srv_instance = srv
		log.Println("[sampling-test] serving at '" + *address + "'")
	} else {
		log.Println("[sampling] serving at '" + *address + "'")
		http.ListenAndServe(*address, nil)
	}
}
