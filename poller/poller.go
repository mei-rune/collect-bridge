package poller

import (
	"commons"
	"ds"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"web"
)

var (
	redisAddress  = flag.String("redis", "127.0.0.1:6379", "the address of redis")
	listenAddress = flag.String("listen", ":7076", "the address of http")
	mdbUrl        = flag.String("ds", "http://127.0.0.1:7071/ds", "the address of ds")
	address       = flag.String("url", "http://127.0.0.1:7070", "the address of bridge")
	timeout       = flag.Int("timeout", 5, "the timeout of http")
)

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

func Runforever() {
	flag.Parse()
	if nil != flag.Args() && 0 != len(flag.Args()) {
		flag.Usage()
		return
	}

	svr := web.NewServer()
	svr.Config.Name = "meijing-poller v1.0"
	svr.Config.Address = *listenAddress
	svr.Get("/", mainHandle)

	client := ds.NewClient(*mdbUrl)
	drvMgr := commons.NewDriverManager()

	redis_channel, err := NewRedis(*redisAddress)
	if nil != err {
		fmt.Println(err)
		return
	}

	res, err := client.FindByWithIncludes("trigger", nil, "action")
	if nil != err {
		fmt.Println("load triggers failed, %v", err)
		return
	}
	ctx := map[string]interface{}{"drvMgr": drvMgr,
		"redis_channel": redis_channel}

	jobs := make([]Job, 0, 100)
	for _, attributes := range res {
		job, e := NewJob(attributes, ctx)
		if nil != e {
			fmt.Printf("create '%v' failed, %v\n", attributes["name"], e)
			continue
		}
		e = job.Start()
		if nil != e {
			fmt.Printf("start '%v' failed, %v\n", attributes["name"], e)
			continue
		}

		fmt.Printf("load '%v' is ok\n", attributes["name"])
		jobs = append(jobs, job)
	}

	svr.Run()
}
