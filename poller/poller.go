package poller

import (
	"commons"
	ds "data_store"
	"expvar"
	"flag"
	"fmt"
	"github.com/emicklei/go-restful"
	"log"
	"net/http"
	pprof "net/http/pprof"
	"os"
	"time"
)

var (
	redisAddress  = flag.String("redis_address", "127.0.0.1:6379", "the address of redis")
	listenAddress = flag.String("poller.listen", ":7073", "the address of http")
	dsUrl         = flag.String("ds", "http://127.0.0.1:7071", "the address of ds")
	sampling_url  = flag.String("sampling", "http://127.0.0.1:7072", "the address of bridge")
	timeout       = flag.Duration("timeout", 1*time.Minute, "the timeout of http")
	refresh       = flag.Duration("refresh", 5*time.Second, "the refresh interval of cache")
	foreignUrl    = flag.String("foreign.url", "http://127.0.0.1:7074", "the url of foreign db")
	load_cookies  = flag.Bool("load_cookies", true, "load cookies is enabled while value is true")
	not_limit     = flag.Bool("not_limit", false, "not limit")

	trigger_exporter                    = &Exporter{}
	Container        *restful.Container = restful.DefaultContainer
	is_test                             = false
	server_test      *server            = nil
)

func init() {
	expvar.Publish("trigger", trigger_exporter)
}

// Var is an abstract type for all exported variables.
type Exporter struct {
	expvar.Var
}

func (self *Exporter) String() string {
	if nil == self.Var {
		return ""
	}
	return self.Var.String()
}

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

func forward(c chan<- []string) chan<- []string {
	return c
}

func forward2(c chan<- *data_object) chan<- *data_object {
	return c
}

func Runforever() {
	flag.Parse()
	if nil != flag.Args() && 0 != len(flag.Args()) {
		flag.Usage()
		return
	}

	alert_foreign, err := newForeignDb("alerts", commons.NewUrlBuilder(*foreignUrl).Concat("alerts").ToUrl())
	if nil != err {
		fmt.Println("connect to foreign db failed,", err)
		return
	}

	if !is_test {
		defer alert_foreign.Close()
	}

	histories_foreign, err := newForeignDb("histories", commons.NewUrlBuilder(*foreignUrl).Concat("histories").ToUrl())
	if nil != err {
		fmt.Println("connect to foreign db failed,", err)
		return
	}

	if !is_test {
		defer histories_foreign.Close()
	}

	redis_client, err := newRedis(*redisAddress)
	if nil != err {
		fmt.Println("connect to redis failed,", err)
		return
	}

	if !is_test {
		defer redis_client.Close()
	}

	ds_client := ds.NewClient(*dsUrl)
	notification_groups := ds.NewCacheWithIncludes(*refresh, ds_client, "notification_group", "action")

	ctx := map[string]interface{}{"sampling.url": *sampling_url,
		"redis_channel":       forward(redis_client.c),
		"alerts_channel":      forward2(alert_foreign.c),
		"histories_channel":   forward2(histories_foreign.c),
		"notification_groups": notification_groups}

	srv := newServer(*refresh, ds_client, ctx)
	err = srv.Start()
	if nil != err {
		fmt.Println(err)
		return
	}
	defer func() {
		if !is_test {
			srv.Stop()
			trigger_exporter.Var = nil
		}
	}()

	trigger_exporter.Var = srv

	restful.DefaultResponseMimeType = restful.MIME_JSON
	ws := new(restful.WebService)
	if is_test {
		ws.Path("jobs")
		server_test = srv
	}
	ws.Route(ws.GET("/").To(mainHandle))
	ws.Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON) // you can specify this per route as well

	ws.Route(ws.GET("/sync").To(srv.Sync).
		Doc("sync all trigger with db")) // on the response

	ws.Route(ws.GET("/all").To(srv.StatsAll).
		Doc("get info of the all triggers")) // on the response

	ws.Route(ws.GET("/by_id/{id}").To(srv.StatsById).
		Doc("get info of the trigger").
		Param(ws.PathParameter("id", "identifier of the trigger").DataType("string"))) // on the response

	ws.Route(ws.GET("/by_name/{name}").To(srv.StatsByName).
		Doc("get info of the trigger").
		Param(ws.PathParameter("name", "name of the trigger").DataType("string"))) // on the response

	ws.Route(ws.GET("/by_address/{address}").To(srv.StatsByAddress).
		Doc("get info of the trigger").
		Param(ws.PathParameter("address", "address of the trigger").DataType("string"))) // on the response
	Container.Add(ws)

	if restful.DefaultContainer != Container {
		Container.Handle("/debug/vars", http.HandlerFunc(expvarHandler))
		Container.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		Container.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		Container.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		Container.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	}

	if is_test {
		log.Println("[poller-test] serving at '" + *listenAddress + "'")
	} else {
		log.Println("[poller] serving at '" + *listenAddress + "'")
		e := http.ListenAndServe(*listenAddress, nil)
		if nil != e {
			log.Println(e)
		}
	}
}
