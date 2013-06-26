package metrics

import (
	"commons"
	"ds"
	"fmt"
	"github.com/runner-mei/go-restful"
	"time"
)

type server struct {
	caches     *ds.Caches
	dispatcher dispatcher
}

func newServer(ds_url string, refresh time.Duration) *server {
	return &server{caches: ds.NewCaches(refresh, ds.NewClient(ds_url), map[string]string{"snmp": "snmp_param"})}
}

func (self *server) returnResult(resp *restful.Response, res commons.Result) {
	if res.HasError() {
		resp.WriteErrorString(res.ErrorCode(), res.ErrorMessage())
	} else {
		if -1 != res.LastInsertId() {
			resp.WriteHeader(commons.CreatedCode)
		}
		resp.WriteEntity(res)
	}
}

type invoke_func func(self *dispatcher, name string, params commons.Map) commons.Result

func (self *server) invoke(req *restful.Request, resp *restful.Response, invoker invoke_func) {
	fmt.Println("Invoke", req.Request.URL)
	managed_type := req.PathParameter("type")
	if 0 == len(managed_type) {
		self.returnResult(resp, commons.ReturnWithIsRequired("type"))
		return
	}
	managed_id := req.PathParameter("id")
	if 0 == len(managed_id) {
		self.returnResult(resp, commons.ReturnWithIsRequired("id"))
		return
	}
	metric_name := req.PathParameter("metric_name")
	if 0 == len(metric_name) {
		self.returnResult(resp, commons.ReturnWithIsRequired("metric_name"))
		return
	}

	query_params := make(map[string]string)
	query_params["metric_name"] = metric_name
	for k, v := range req.Request.URL.Query() {
		query_params[k] = v[len(v)-1]
	}
	params := commons.Proxy(commons.StringMap(query_params),
		context{managed_type: managed_type,
			managed_id: managed_id,
			caches:     self.caches})

	self.returnResult(resp, invoker(&self.dispatcher, metric_name, params))
}

func (self *server) Get(req *restful.Request, resp *restful.Response) {
	self.invoke(req, resp, (*dispatcher).Get)
}

func (self *server) Put(req *restful.Request, resp *restful.Response) {
	self.invoke(req, resp, (*dispatcher).Put)
}

func (self *server) Create(req *restful.Request, resp *restful.Response) {
	self.invoke(req, resp, (*dispatcher).Create)
}

func (self *server) Delete(req *restful.Request, resp *restful.Response) {
	self.invoke(req, resp, (*dispatcher).Delete)
}
