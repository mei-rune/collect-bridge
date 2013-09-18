package sampling

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"snmp"
	"time"
)

var (
	redisAddress       = flag.String("redis.address", "127.0.0.1:6379", "the address of redis")
	address            = flag.String("sampling.listen", ":7072", "the address of http")
	ds_url             = flag.String("ds.url", "http://127.0.0.1:7071", "the address of http")
	refresh            = flag.Duration("ds.refresh", 60*time.Second, "the duration of refresh")
	snmp_timeout       = flag.Duration("snmp.timeout", 30*time.Second, "the timeout duration of snmp")
	period_interval    = flag.Duration("period", 1*time.Second, "the tick interval of backgroundWorker")
	flux_expired       = flag.Int64("flux_expired", 60*5, "remove it from scan list while port or link is expired")
	flux_buffer_size   = flag.Int("flux_buffer_size", 30, "the default buffer size of flux")
	icmp_buffer_size   = flag.Int("icmp_buffer_size", 30, "the default buffer size of icmp")
	icmp_poll_interval = flag.Uint64("icmp_poll_interval", 5, "the interval(second) of icmp scan")
	icmp_timeout       = flag.Int64("icmp_timeout", 5, "the timeout (second) of icmp")
	icmp_expired       = flag.Int64("icmp_expired", 60*5, "remove it from scan list while address is expired")

	snmp_test_buffer_size   = flag.Int("snmp_test_buffer_size", 30, "the default buffer size of snmp test")
	snmp_test_poll_interval = flag.Uint64("snmp_test_poll_interval", 5, "the interval(second) of snmp scan")
	snmp_test_timeout       = flag.Int64("snmp_test_timeout", 5, "the timeout (second) of snmp")
	snmp_test_expired       = flag.Int64("snmp_test_expired", 60*5, "remove it from scan list while address is expired")

	dump_request = flag.Bool("dump_request", false, "dump request info to redis for perf")

	is_test              = false
	srv_instance *server = nil
)

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

	if is_test {
		srv_instance = srv
		log.Println("[sampling-test] serving at '" + *address + "'")
	} else {
		go srv.run()
		defer func() { srv.close() }()

		log.Println("[sampling] serving at '" + *address + "'")
		http.ListenAndServe(*address, srv)
	}
}
