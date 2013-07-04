package poller

import (
	"ds"
	_ "expvar"
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
	refresh       = flag.Duration("refresh", 10*time.Second, "the refresh interval of cache")

	is_test         = false
	jobs_test []Job = nil
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
	results, err := ds_client.FindByWithIncludes("trigger", map[string]string{}, "action")
	if nil != err {
		fmt.Println("load triggers from db failed,", err)
		return
	}

	ctx := map[string]interface{}{"metrics.url": *metrics_url,
		"redis_channel": forward(redis_channel)}

	jobs := make([]Job, 0, 100)
	for _, attributes := range results {
		name := attributes["name"]
		id := attributes["id"]

		job, e := newJob(attributes, ctx)
		if nil != e {
			fmt.Printf("create '%v:%v' failed, %v\n", id, name, e)
			continue
		}
		e = job.Start()
		if nil != e {
			fmt.Printf("start '%v:%v' failed, %v\n", id, name, e)
			continue
		}

		fmt.Printf("load '%v:%v' is ok\n", id, name)
		jobs = append(jobs, job)
	}

	ws := new(restful.WebService)
	if is_test {
		ws.Path("job")
		jobs_test = jobs
	}
	ws.Route(ws.GET("/").To(mainHandle))
	ws.Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML) // you can specify this per route as well

	// ws.Route(ws.PUT("/{id}").To().
	// 	Doc("put a metric").
	// 	Param(ws.PathParameter("type", "type of the instance").DataType("string")).
	// 	Param(ws.PathParameter("id", "identifier of the instance").DataType("string")).
	// 	Param(ws.PathParameter("metric_name", "name of the metric").DataType("string"))) // on the response

	restful.Add(ws)

	if is_test {
		log.Println("[ds-test] serving at '" + *listenAddress + "'")
	} else {
		log.Println("[ds] serving at '" + *listenAddress + "'")
		http.ListenAndServe(*listenAddress, nil)
	}
}
