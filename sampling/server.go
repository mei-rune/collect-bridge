package sampling

import (
	"commons"
	ds "data_store"
	"errors"
	"github.com/runner-mei/go-restful"
	"time"

	"commons/types"
	"net"
	"net/http"
	"testing"
)

var (
	empty_mo    = map[string]interface{}{}
	alias_names = map[string]string{"snmp": "snmp_param"}
)

type server struct {
	caches     *ds.Caches
	mo_cache   *ds.Cache
	dispatcher *dispatcher
}

func newServer(ds_url string, refresh time.Duration, params map[string]interface{}) (*server, error) {
	dispatch, e := newDispatcher(params)
	if nil != e {
		return nil, errors.New("new dispatcher failed, " + e.Error())
	}
	caches := ds.NewCaches(refresh, ds.NewClient(ds_url),
		"*", map[string]string{"snmp": "snmp_param"})
	mo_cache, e := caches.GetCache("managed_object")
	if nil != e {
		return nil, errors.New("get cache 'managed_object' failed, " + e.Error())
	}
	if nil == mo_cache {
		return nil, errors.New("table 'managed_object' is not exists.")
	}

	return &server{caches: caches, mo_cache: mo_cache,
		dispatcher: dispatch}, nil
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

	mo, e := self.mo_cache.Get(managed_id)
	if nil != e {
		self.returnResult(resp, commons.ReturnWithNotFound(managed_type, managed_id))
		return
	}

	params := &context{params: query_params,
		managed_type: managed_type,
		managed_id:   managed_id,
		mo:           mo,
		local:        make(map[string]map[string]interface{}),
		alias:        alias_names,
		pry:          &proxy{dispatcher: self.dispatcher}}

	self.returnResult(resp, invoker(self.dispatcher, metric_name, params))
}

func (self *server) native_invoke(req *restful.Request, resp *restful.Response, invoker invoke_func) {
	ip := req.PathParameter("ip")
	if 0 == len(ip) {
		self.returnResult(resp, commons.ReturnWithIsRequired("ip"))
		return
	}
	metric_name := req.PathParameter("metric_name")
	if 0 == len(metric_name) {
		self.returnResult(resp, commons.ReturnWithIsRequired("metric_name"))
		return
	}

	query_params := make(map[string]string)
	query_params["@address"] = ip
	query_params["metric_name"] = metric_name
	for k, v := range req.Request.URL.Query() {
		query_params[k] = v[len(v)-1]
	}

	params := &context{params: query_params,
		managed_type: "unknow_type",
		managed_id:   "unknow_id",
		mo:           empty_mo,
		local:        make(map[string]map[string]interface{}),
		alias:        alias_names,
		pry:          &proxy{dispatcher: self.dispatcher}}

	self.returnResult(resp, invoker(self.dispatcher, metric_name, params))
}

func (self *server) Get(req *restful.Request, resp *restful.Response) {
	//fmt.Println("Get")
	self.invoke(req, resp, (*dispatcher).Get)
}

func (self *server) Put(req *restful.Request, resp *restful.Response) {
	//fmt.Println("Put")
	self.invoke(req, resp, (*dispatcher).Put)
}

func (self *server) Create(req *restful.Request, resp *restful.Response) {
	//fmt.Println("Create")
	self.invoke(req, resp, (*dispatcher).Create)
}

func (self *server) Delete(req *restful.Request, resp *restful.Response) {
	//fmt.Println("Delete")
	self.invoke(req, resp, (*dispatcher).Delete)
}

func (self *server) NativeGet(req *restful.Request, resp *restful.Response) {
	//fmt.Println("NativeGet")
	self.native_invoke(req, resp, (*dispatcher).Get)
}

func (self *server) NativePut(req *restful.Request, resp *restful.Response) {
	//fmt.Println("NativePut")
	self.native_invoke(req, resp, (*dispatcher).Put)
}

func (self *server) NativeCreate(req *restful.Request, resp *restful.Response) {
	//fmt.Println("NativeCreate")
	self.native_invoke(req, resp, (*dispatcher).Create)
}

func (self *server) NativeDelete(req *restful.Request, resp *restful.Response) {
	//fmt.Println("NativeDelete")
	self.native_invoke(req, resp, (*dispatcher).Delete)
}

func SrvTest(t *testing.T, file string, cb func(client *ds.Client, definitions *types.TableDefinitions)) {
	ds.SrvTest(t, file, func(client *ds.Client, definitions *types.TableDefinitions) {
		is_test = true
		Main()

		listener, e := net.Listen("tcp", *address)
		if nil != e {
			t.Error("start tcp listen failed,", e)
			return
		}
		ch := make(chan string)

		go func() {
			defer func() {
				ch <- "exit"
			}()
			ch <- "ok"
			http.Serve(listener, nil)
		}()

		s := <-ch
		if "ok" != s {
			t.Error("start http listen failed")
			return
		}

		cb(client, definitions)

		if nil != srv_instance {
			srv_instance = nil
		}
		if nil != listener {
			listener.Close()
		}
		<-ch

	})
}
