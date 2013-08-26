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
	address      = flag.String("sampling.listen", ":7072", "the address of http")
	ds_url       = flag.String("ds.url", "http://127.0.0.1:7071", "the address of http")
	refresh      = flag.Duration("ds.refresh", 60*time.Second, "the duration of refresh")
	snmp_timeout = flag.Duration("snmp.timeout", 60*time.Second, "the timeout duration of snmp")

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
		log.Println("[sampling] serving at '" + *address + "'")
		http.ListenAndServe(*address, srv)
	}
}
