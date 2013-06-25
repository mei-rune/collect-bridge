package metrics

import (
	"ds"
	"github.com/runner-mei/go-restful"
	"time"
)

type server struct {
	caches     *ds.Caches
	dispatcher Dispatcher
}

func newServer(ds_url string, refresh time.Duration) *server {
	return &server{caches: ds.NewCaches(refresh, ds.NewClient(ds_url), map[string]string{"snmp": "snmp_param"})}
}

func (self *server) Get(req *restful.Request, resp *restful.Response) {
	managed_type := req.PathParameter("type")
	if 0 == len(managed_type) {
		return commons.ReturnWithIsRequired("type")
	}
	managed_id := req.PathParameter("id")
	if 0 == len(managed_id) {
		return commons.ReturnWithIsRequired("id")
	}
	metric_name := req.PathParameter("metric_name")
	if 0 == len(metric_name) {
		return commons.ReturnWithIsRequired("metric_name")
	}

}

func (self *server) Put(req *restful.Request, resp *restful.Response) {

}

func (self *server) Create(req *restful.Request, resp *restful.Response) {

}

func (self *server) Delete(req *restful.Request, resp *restful.Response) {

}
