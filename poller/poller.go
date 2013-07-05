package poller

import (
	"ds"
	"expvar"
	"flag"
	"fmt"
	"github.com/runner-mei/go-restful"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"
)

var (
	redisAddress  = flag.String("redis", "127.0.0.1:6379", "the address of redis")
	listenAddress = flag.String("listen", ":7076", "the address of http")
	dsUrl         = flag.String("ds", "http://127.0.0.1:7071", "the address of ds")
	metrics_url   = flag.String("metrics.url", "http://127.0.0.1:7072", "the address of bridge")
	timeout       = flag.Duration("timeout", 1*time.Minute, "the timeout of http")
	refresh       = flag.Duration("refresh", 5, "the refresh interval of cache")

	is_test             = false
	server_test *server = nil
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

func forward(c chan<- []string) chan<- []string {
	return c
}
func Runforever() {
	flag.Parse()
	if nil != flag.Args() && 0 != len(flag.Args()) {
		flag.Usage()
		return
	}

	redis_channel, err := newRedis(*redisAddress)
	if nil != err {
		fmt.Println("connect to redis failed,", err)
		return
	}

	ds_client := ds.NewClient(*dsUrl)

	ctx := map[string]interface{}{"metrics.url": *metrics_url,
		"redis_channel": forward(redis_channel)}

	srv := newServer(*refresh, ds_client, ctx)
	err = srv.Start()
	if nil != err {
		fmt.Println(err)
		return
	}
	defer func() {
		if !is_test {
			srv.Stop()
		}
	}()

	expvar.Publish("triggers", srv)

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

	restful.Add(ws)

	if is_test {
		log.Println("[poller-test] serving at '" + *listenAddress + "'")
	} else {
		log.Println("[poller] serving at '" + *listenAddress + "'")
		http.ListenAndServe(*listenAddress, nil)
	}
}
