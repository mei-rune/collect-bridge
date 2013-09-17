package sampling

import (
	"bytes"
	"commons"
	"commons/types"
	ds "data_store"
	"encoding/json"
	"errors"
	"expvar"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/pprof"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var (
	alias_names  = map[string]string{"snmp": "snmp_param"}
	srv_exporter = &exporter{}
)

func init() {
	expvar.Publish("sampling_server", srv_exporter)
}

func type_error(e error) error {
	return TypeError
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

type server struct {
	workers  *backgroundWorkers
	caches   *ds.Caches
	mo_cache *ds.Cache
	perf     *PerfServer

	routes_for_get    map[string]*RouterGroup
	routes_for_put    map[string]*RouterGroup
	routes_for_create map[string]*RouterGroup
	routes_for_delete map[string]*RouterGroup

	route_for_get    map[string]*Route
	route_for_put    map[string]*Route
	route_for_create map[string]*Route
	route_for_delete map[string]*Route

	completions_lock sync.Mutex
	completions      []*ExchangeResponse

	pending_requests   int64
	last_request_size  int
	max_request_size   int
	last_used_duration time.Duration
	max_used_duration  time.Duration
}

func newServer(ds_url string, refresh time.Duration, params map[string]interface{}) (*server, error) {
	client := ds.NewClient(ds_url)
	caches := ds.NewCaches(refresh, client, "*", map[string]string{"snmp": "snmp_param"})
	mo_cache := ds.NewCacheWithIncludes(refresh, client, "managed_object", "*")

	srv := &server{workers: &backgroundWorkers{c: make(chan func()),
		period_interval: *period_interval,
		workers:         nil},
		caches: caches, mo_cache: mo_cache,
		routes_for_get:    make(map[string]*RouterGroup),
		routes_for_put:    make(map[string]*RouterGroup),
		routes_for_create: make(map[string]*RouterGroup),
		routes_for_delete: make(map[string]*RouterGroup),
		route_for_get:     make(map[string]*Route),
		route_for_put:     make(map[string]*Route),
		route_for_create:  make(map[string]*Route),
		route_for_delete:  make(map[string]*Route)}

	if nil == params {
		params = map[string]interface{}{}
	}
	simpled := &simpleWorkers{workers: make(map[string]*worker)}
	wrpped := &wrappedWorkers{backend: simpled}
	params["backgroundWorkers"] = wrpped

	for k, rs := range Methods {
		r, e := newRouteWithSpec(k, rs, params)
		if nil != e {
			return nil, errors.New("init '" + k + "' failed, " + e.Error())
		}

		e = srv.register(r)
		if nil != e {
			return nil, errors.New("register '" + k + "' failed, " + e.Error())
		}
	}

	srv.workers.workers = simpled.workers
	wrpped.backend = srv.workers
	perf, e := NewPerfServer(*redisAddress)
	if nil != e {
		return nil, e
	}
	srv.perf = perf
	srv_exporter.Var = srv
	return srv, nil
}

func (self *server) stats() interface{} {
	return map[string]interface{}{"pending—responses": self.completion_size(),
		"pending_requests":   atomic.LoadInt64(&self.pending_requests),
		"last_request_size":  self.last_request_size,
		"max_request_size":   self.max_request_size,
		"last_used_duration": self.last_used_duration.String(),
		"max_used_duration":  self.max_used_duration.String()}
}

func (self *server) close() {
	self.workers.close()
}

func (self *server) shutdown() {
	self.workers.shutdown()
}

func (self *server) run() {
	self.workers.run()
}

func (self *server) register(rs *Route) error {
	var routes_container map[string]*RouterGroup
	var route_container map[string]*Route

	switch rs.definition.Method {
	case "get", "Get", "GET":
		route_container = self.route_for_get
		routes_container = self.routes_for_get
	case "put", "Put", "PUT":
		route_container = self.route_for_put
		routes_container = self.routes_for_put
	case "create", "Create", "CREATE":
		route_container = self.route_for_create
		routes_container = self.routes_for_create
	case "delete", "Delete", "DELETE":
		route_container = self.route_for_delete
		routes_container = self.routes_for_delete
	default:
		return errors.New("Unsupported method - " + rs.definition.Method)
	}

	if _, ok := route_container[rs.id]; ok {
		return errors.New("route that id is  '" + rs.id + "' is already exists.")
	}
	route_container[rs.id] = rs

	route, _ := routes_container[rs.name]
	if nil == route {
		route = &RouterGroup{}
		routes_container[rs.name] = route
	}

	return route.register(rs)
}

func (self *server) unregister(name, id string) {
	for _, instances := range []map[string]*RouterGroup{self.routes_for_get,
		self.routes_for_put, self.routes_for_create, self.routes_for_delete} {
		if "" == name {
			for _, route := range instances {
				route.unregister(id)
			}
		} else {
			route, _ := instances[name]
			if nil == route {
				return
			}
			route.unregister(id)
		}
	}
}

func (self *server) clear() {
	self.routes_for_get = make(map[string]*RouterGroup)
	self.routes_for_put = make(map[string]*RouterGroup)
	self.routes_for_create = make(map[string]*RouterGroup)
	self.routes_for_delete = make(map[string]*RouterGroup)

	self.route_for_get = make(map[string]*Route)
	self.route_for_put = make(map[string]*Route)
	self.route_for_create = make(map[string]*Route)
	self.route_for_delete = make(map[string]*Route)
}

func notAcceptable(metric_name string) commons.Result {
	return commons.ReturnWithNotAcceptable("'" + metric_name + "' is not acceptable.")
}

func (self *server) returnResult(resp http.ResponseWriter, res commons.Result) {
	if res.HasError() {
		resp.WriteHeader(res.ErrorCode())
		io.WriteString(resp, res.ErrorMessage())
	} else {
		if -1 != res.LastInsertId() {
			resp.WriteHeader(commons.CreatedCode)
		}
		e := json.NewEncoder(resp).Encode(res)
		if nil != e {
			resp.WriteHeader(http.StatusInternalServerError)
			io.WriteString(resp, e.Error())
		}
	}
}

func to2DArray(ss []string) []P {
	params := make([]P, 0, len(ss)/2)
	for i := 0; i < len(ss); i += 2 {
		params = append(params, P{ss[i], ss[i+1]})
	}
	return params
}

type routeObject interface {
	Invoke(paths []P, params MContext) commons.Result
}

func (self *server) route(action, metric_name string) (routeObject, commons.RuntimeError) {
	var route_by_id map[string]*Route
	var routes_by_name map[string]*RouterGroup

	switch action {
	case "GET":
		route_by_id = self.route_for_get
		routes_by_name = self.routes_for_get
	case "PUT":
		route_by_id = self.route_for_put
		routes_by_name = self.routes_for_put
	case "POST":
		route_by_id = self.route_for_create
		routes_by_name = self.routes_for_create
	case "DELETE":
		route_by_id = self.route_for_delete
		routes_by_name = self.routes_for_delete
	default:
		return nil, commons.NewApplicationError(http.StatusMethodNotAllowed,
			"method '"+action+"' is unsupported.")
	}

	if route := route_by_id[metric_name]; nil != route {
		return route, nil
	}

	if routes := routes_by_name[metric_name]; nil != routes {
		return routes, nil
	}

	return nil, commons.NewApplicationError(http.StatusNotAcceptable,
		"'"+metric_name+"' is not acceptable.")
}

func (self *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if nil != r.Body {
			r.Body.Close()
		}
	}()

	defer func() {
		if e := recover(); nil != e {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("[panic]%v", e))
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}

			w.WriteHeader(http.StatusInternalServerError)
			w.Write(buffer.Bytes())
		}
	}()

	if strings.HasPrefix(r.URL.Path, "/debug/") {
		switch r.URL.Path {
		case "/debug/vars":
			expvarHandler(w, r)
		case "/debug/pprof", "/debug/pprof/":
			pprof.Index(w, r)
		case "/debug/pprof/cmdline":
			pprof.Cmdline(w, r)
		case "/debug/pprof/profile":
			pprof.Profile(w, r)
		case "/debug/pprof/symbol", "/debug/pprof/symbol/":
			pprof.Symbol(w, r)
		default:
			if strings.HasPrefix(r.URL.Path, "/debug/pprof/") {
				pprof.Index(w, r)
				return
			}

			http.NotFound(w, r)
		}
		return
	}

	if r.URL.Path == "/batch" || r.URL.Path == "/batch/" {
		if "POST" != r.Method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			io.WriteString(w, "method '"+r.Method+"' is unsupported, must is 'POST'.")
			return
		}

		self.batchExchange(w, r)
		return
	}

	url := r.URL.Path
	if '/' == url[0] {
		url = url[1:]
	}

	paths := strings.Split(url, "/")
	if 2 > len(paths) {
		http.NotFound(w, r)
		return
	}

	var address, managed_type, managed_id, metric_name string
	var is_native bool
	var query_paths []P

	if 0 == len(paths)%2 {
		metric_name = paths[len(paths)-1]
		managed_type = "unknow_type"
		managed_id = "unknow_id"
		is_native = true
		address = paths[0]
		query_paths = to2DArray(paths[1 : len(paths)-1])

	} else {
		metric_name = paths[len(paths)-1]
		managed_type = paths[0]
		managed_id = paths[1]
		is_native = false
		query_paths = to2DArray(paths[2 : len(paths)-1])
	}

	route, e := self.route(r.Method, metric_name)
	if nil != e {
		w.WriteHeader(e.Code())
		io.WriteString(w, e.Error())
		return
	}

	query_params := map[string]string{}
	for k, values := range r.URL.Query() {
		query_params[k] = values[len(values)-1]
	}

	ctx := &context{srv: self,
		alias:        alias_names,
		metric_name:  metric_name,
		is_native:    is_native,
		address:      address,
		managed_type: managed_type,
		managed_id:   managed_id,
		query_paths:  query_paths,
		query_params: query_params,
		body_reader:  r.Body}

	if e := ctx.init(); nil != e {
		if err := e.(commons.RuntimeError); nil != err {
			w.WriteHeader(err.Code())
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		io.WriteString(w, e.Error())
		return
	}

	self.returnResult(w, route.Invoke(ctx.query_paths, ctx))
}

func (self *server) batchExchange(w http.ResponseWriter, r *http.Request) {

	begin_at := time.Now()
	defer func() {
		self.last_used_duration = time.Now().Sub(begin_at)
		if self.last_used_duration > self.max_used_duration {
			self.max_used_duration = self.last_used_duration
		}
	}()

	var requests []*ExchangeRequest
	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	if e := decoder.Decode(&requests); nil != e {
		w.WriteHeader(http.StatusExpectationFailed)
		io.WriteString(w, e.Error())
		return
	}

	if nil != requests && 0 != len(requests) {
		for _, req := range requests {
			go self.doRequest(req)
			atomic.AddInt64(&self.pending_requests, 1)
		}
		self.last_request_size = len(requests)
		if self.last_request_size > self.max_request_size {
			self.max_request_size = self.last_request_size
		}

	} else {
		self.last_request_size = 0
	}

	self.sendCompletion(w)
}

func (self *server) completion_size() int {
	self.completions_lock.Lock()
	defer self.completions_lock.Unlock()
	return len(self.completions)
}
func (self *server) sendCompletion(w http.ResponseWriter) {
	self.completions_lock.Lock()
	defer self.completions_lock.Unlock()

	if nil == self.completions || 0 == len(self.completions) {
		w.WriteHeader(http.StatusNoContent)
		//io.WriteString(w, "NOCONTENT")
		return
	}

	w.WriteHeader(http.StatusAccepted)
	encoder := json.NewEncoder(w)
	if e := encoder.Encode(self.completions); nil != e {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, e.Error())
		//fmt.Println("----------", self.completions[0])
	} else {
		//fmt.Println(self.completions[0])
		if 256 < cap(self.completions) {
			self.completions = nil
		} else {
			self.completions = self.completions[0:0]
		}
	}
}

func (self *server) doRequest(r *ExchangeRequest) {
	var resp *ExchangeResponse
	var res commons.Result
	var ctx *context

	begin_at := time.Now()
	defer func() {
		end_at := time.Now()
		self.perf.C <- []string{"LPUSH", "perf-" + r.ChannelName, `{"begin_at":"` + begin_at.Format(time.RFC3339Nano) + `","end_at":"` + end_at.Format(time.RFC3339Nano) + `","elapsed":"` + end_at.Sub(begin_at).String() + `","elapsed(ns)":` + strconv.FormatInt(int64(end_at.Sub(begin_at)), 10) + `, "error":"` + resp.ErrorMessage() + `"}`}
		self.perf.C <- []string{"LTRIM", "perf-" + r.ChannelName, "0", "100"}

		atomic.AddInt64(&self.pending_requests, -1)
	}()

	route, err := self.route(r.Action, r.Name)
	if nil != err {
		resp = &ExchangeResponse{Id: r.Id, EcreatedAt: time.Now(), ChannelName: r.ChannelName,
			Eerror: &commons.ApplicationError{Ecode: err.Code(), Emessage: err.Error()}}
		goto failed
	}

	ctx = &context{srv: self,
		alias:            alias_names,
		metric_name:      r.Name,
		is_native:        0 != len(r.Address),
		address:          r.Address,
		managed_type:     r.ManagedType,
		managed_id:       r.ManagedId,
		query_paths:      r.Paths,
		query_params:     r.Params,
		body_unmarshaled: true,
		body_instance:    r.Body}

	if e := ctx.init(); nil != e {
		if err := e.(commons.RuntimeError); nil != err {
			resp = &ExchangeResponse{Id: r.Id, EcreatedAt: time.Now(), ChannelName: r.ChannelName,
				Eerror: &commons.ApplicationError{Ecode: err.Code(), Emessage: err.Error()}}
		} else {
			resp = &ExchangeResponse{Id: r.Id, EcreatedAt: time.Now(), ChannelName: r.ChannelName,
				Eerror: &commons.ApplicationError{Ecode: http.StatusInternalServerError, Emessage: e.Error()}}
		}
		goto failed
	}

	res = route.Invoke(ctx.query_paths, ctx)
	if res.HasError() {
		resp = &ExchangeResponse{Id: r.Id, EcreatedAt: res.CreatedAt(), ChannelName: r.ChannelName,
			Eerror: &commons.ApplicationError{Ecode: res.ErrorCode(), Emessage: res.ErrorMessage()}}
	} else {
		resp = &ExchangeResponse{Id: r.Id, EcreatedAt: res.CreatedAt(), ChannelName: r.ChannelName,
			Evalue: res.InterfaceValue()}
	}

failed:
	self.completions_lock.Lock()
	defer self.completions_lock.Unlock()
	if nil == resp {
		resp = &ExchangeResponse{Id: r.Id, EcreatedAt: time.Now(), ChannelName: r.ChannelName,
			Eerror: &commons.ApplicationError{Ecode: http.StatusInternalServerError, Emessage: "unknow error."}}
	}

	self.completions = append(self.completions, resp)
}

func invoke(route_by_id map[string]*Route, routes_by_name map[string]*RouterGroup,
	metric_name string, paths []P, params MContext) commons.Result {
	route := route_by_id[metric_name]
	if nil != route {
		return route.Invoke(paths, params)
	}

	routes := routes_by_name[metric_name]
	if nil == routes {
		return notAcceptable(metric_name)
	}

	return routes.Invoke(paths, params)
}

func (self *server) InvokeGet(metric_name string, paths []P, params MContext) commons.Result {
	return invoke(self.route_for_get, self.routes_for_get, metric_name, paths, params)
}

func (self *server) InvokePut(metric_name string, paths []P, params MContext) commons.Result {
	return invoke(self.route_for_put, self.routes_for_put, metric_name, paths, params)
}

func (self *server) InvokeCreate(metric_name string, paths []P, params MContext) commons.Result {
	return invoke(self.route_for_create, self.routes_for_create, metric_name, paths, params)
}

func (self *server) InvokeDelete(metric_name string, paths []P, params MContext) commons.Result {
	return invoke(self.route_for_delete, self.routes_for_delete, metric_name, paths, params)
}
func (self *server) GetMOCahce(id string) (map[string]interface{}, error) {
	return self.mo_cache.Get(id)
}

func (self *server) Get(metric_name string, paths []P, params MContext) (interface{}, error) {
	res := self.InvokeGet(metric_name, paths, params)
	if res.HasError() {
		return nil, res.Error()
	}
	return res.InterfaceValue(), nil
}

func (self *server) GetBool(metric_name string, paths []P, params MContext) (bool, error) {
	res := self.InvokeGet(metric_name, paths, params)
	if res.HasError() {
		return false, res.Error()
	}
	b, e := res.Value().AsBool()
	if nil != e {
		return false, type_error(e)
	}
	return b, nil
}

func (self *server) GetInt(metric_name string, paths []P, params MContext) (int, error) {
	res := self.InvokeGet(metric_name, paths, params)
	if res.HasError() {
		return 0, res.Error()
	}
	i, e := res.Value().AsInt()
	if nil != e {
		return 0, type_error(e)
	}
	return i, nil
}

func (self *server) GetInt32(metric_name string, paths []P, params MContext) (int32, error) {
	res := self.InvokeGet(metric_name, paths, params)
	if res.HasError() {
		return 0, res.Error()
	}
	i32, e := res.Value().AsInt32()
	if nil != e {
		return 0, type_error(e)
	}
	return i32, nil
}

func (self *server) GetInt64(metric_name string, paths []P, params MContext) (int64, error) {
	res := self.InvokeGet(metric_name, paths, params)
	if res.HasError() {
		return 0, res.Error()
	}
	i64, e := res.Value().AsInt64()
	if nil != e {
		return 0, type_error(e)
	}
	return i64, nil
}

func (self *server) GetUint(metric_name string, paths []P, params MContext) (uint, error) {
	res := self.InvokeGet(metric_name, paths, params)
	if res.HasError() {
		return 0, res.Error()
	}
	u, e := res.Value().AsUint()
	if nil != e {
		return 0, type_error(e)
	}
	return u, nil
}

func (self *server) GetUint32(metric_name string, paths []P, params MContext) (uint32, error) {
	res := self.InvokeGet(metric_name, paths, params)
	if res.HasError() {
		return 0, res.Error()
	}
	u32, e := res.Value().AsUint32()
	if nil != e {
		return 0, type_error(e)
	}
	return u32, nil
}

func (self *server) GetUint64(metric_name string, paths []P, params MContext) (uint64, error) {
	res := self.InvokeGet(metric_name, paths, params)
	if res.HasError() {
		return 0, res.Error()
	}
	u64, e := res.Value().AsUint64()
	if nil != e {
		return 0, type_error(e)
	}
	return u64, nil
}

func (self *server) GetString(metric_name string, paths []P, params MContext) (string, error) {
	res := self.InvokeGet(metric_name, paths, params)
	if res.HasError() {
		return "", res.Error()
	}

	s, e := res.Value().AsString()
	if nil != e {
		return "", type_error(e)
	}
	return s, nil
}

func (self *server) GetObject(metric_name string, paths []P, params MContext) (map[string]interface{}, error) {
	res := self.InvokeGet(metric_name, paths, params)
	if res.HasError() {
		return nil, res.Error()
	}

	o, e := res.Value().AsObject()
	if nil != e {
		return nil, type_error(e)
	}
	return o, nil
}

func (self *server) GetArray(metric_name string, paths []P, params MContext) ([]interface{}, error) {
	res := self.InvokeGet(metric_name, paths, params)
	if res.HasError() {
		return nil, res.Error()
	}

	a, e := res.Value().AsArray()
	if nil != e {
		return nil, type_error(e)
	}
	return a, nil
}

func (self *server) GetObjects(metric_name string, paths []P, params MContext) ([]map[string]interface{}, error) {
	res := self.InvokeGet(metric_name, paths, params)
	if res.HasError() {
		return nil, res.Error()
	}

	o, e := res.Value().AsObjects()
	if nil != e {
		return nil, type_error(e)
	}
	return o, nil
}

func SrvTest(t *testing.T, file string, cb func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions)) {
	ds.SrvTest(t, file, func(client *ds.Client, definitions *types.TableDefinitions) {
		*ds_url = client.Url
		is_test = true
		Main()

		if nil == srv_instance {
			t.Error("srv_instance is nil")
			return
		}

		hsrv := httptest.NewServer(srv_instance)
		fmt.Println("[sampling-test] serving at '" + hsrv.URL + "'")
		defer hsrv.Close()

		cb(client, hsrv.URL, definitions)

		if nil != srv_instance {
			srv_instance.close()
			srv_instance = nil
		}
	})
}
