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
	"strings"
	"testing"
	"time"
)

var (
	alias_names = map[string]string{"snmp": "snmp_param"}
)

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

	routes_for_get    map[string]*Routers
	routes_for_put    map[string]*Routers
	routes_for_create map[string]*Routers
	routes_for_delete map[string]*Routers

	route_for_get    map[string]*Route
	route_for_put    map[string]*Route
	route_for_create map[string]*Route
	route_for_delete map[string]*Route
}

func newServer(ds_url string, refresh time.Duration, params map[string]interface{}) (*server, error) {
	client := ds.NewClient(ds_url)
	caches := ds.NewCaches(refresh, client, "*", map[string]string{"snmp": "snmp_param"})
	mo_cache := ds.NewCacheWithIncludes(refresh, client, "managed_object", "*")

	srv := &server{workers: &backgroundWorkers{c: make(chan func()),
		period_interval: *period_interval,
		workers:         nil},
		caches: caches, mo_cache: mo_cache,
		routes_for_get:    make(map[string]*Routers),
		routes_for_put:    make(map[string]*Routers),
		routes_for_create: make(map[string]*Routers),
		routes_for_delete: make(map[string]*Routers),
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
	return srv, nil
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
	var routes_container map[string]*Routers
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
		route = &Routers{}
		routes_container[rs.name] = route
	}

	return route.register(rs)
}

func (self *server) unregister(name, id string) {
	for _, instances := range []map[string]*Routers{self.routes_for_get,
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
	self.routes_for_get = make(map[string]*Routers)
	self.routes_for_put = make(map[string]*Routers)
	self.routes_for_create = make(map[string]*Routers)
	self.routes_for_delete = make(map[string]*Routers)

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
		case "/debug/pprof/symbol":
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
	url := r.URL.Path
	if '/' == url[0] {
		url = url[1:]
	}

	paths := strings.Split(url, "/")
	if 2 > len(paths) {
		http.NotFound(w, r)
		return
	}

	var managed_type, managed_id, metric_name string
	var mo map[string]interface{}
	var query_paths []P
	query_params := map[string]string{}
	for k, values := range r.URL.Query() {
		query_params[k] = values[len(values)-1]
	}

	if 0 == len(paths)%2 {
		metric_name = paths[len(paths)-1]
		managed_type = "unknow_type"
		managed_id = "unknow_id"
		query_params["uid"] = paths[0]
		query_params["@address"] = paths[0]
		query_params["metric-name"] = metric_name
		query_paths = to2DArray(paths[1 : len(paths)-1])

		mo = map[string]interface{}{}
	} else {
		metric_name = paths[len(paths)-1]
		managed_type = paths[0]
		managed_id = paths[1]
		query_params["type"] = managed_type
		query_params["id"] = managed_id
		query_params["uid"] = managed_id
		query_params["metric-name"] = metric_name
		query_paths = to2DArray(paths[2 : len(paths)-1])

		var e error
		mo, e = self.mo_cache.Get(managed_id)
		if nil != e {
			self.returnResult(w, commons.ReturnWithInternalError(e.Error()))
			return
		}
		if nil == mo {
			self.returnResult(w, commons.ReturnWithRecordNotFound(managed_type, managed_id))
			return
		}
		switch mo["type"] {
		case "network_device_port":
			device_id := fmt.Sprint(mo["device_id"])
			ifIndex := fmt.Sprint(mo["if_index"])
			mo, e = self.mo_cache.Get(device_id)
			if nil != e {
				self.returnResult(w, commons.ReturnWithInternalError(e.Error()))
				return
			}
			if nil == mo {
				self.returnResult(w, commons.ReturnWithRecordNotFound("device", managed_id))
				return
			}
			query_paths = []P{{"port", ifIndex}}
		}
		// if "managed_object" == managed_type && "137" == managed_id {
		// 	bs, _ := json.MarshalIndent(mo, "", "  ")
		// 	fmt.Println(string(bs))
		// }
	}

	var route_by_id map[string]*Route
	var routes_by_name map[string]*Routers

	switch r.Method {
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
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "method '"+r.Method+"' is unsupported.")
		return
	}

	params := &context{params: query_params,
		managed_type: managed_type,
		managed_id:   managed_id,
		mo:           mo,
		local:        make(map[string]map[string]interface{}),
		alias:        alias_names,
		pry:          &proxy{srv: self},
		body_reader:  r.Body}

	route := route_by_id[metric_name]
	if nil != route {
		self.returnResult(w, route.Invoke(query_paths, params))
		return
	}

	routes := routes_by_name[metric_name]
	if nil == routes {
		self.returnResult(w, notAcceptable(metric_name))
		return
	}

	self.returnResult(w, routes.Invoke(query_paths, params))
}

func invoke(route_by_id map[string]*Route, routes_by_name map[string]*Routers,
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

func (self *server) CreateCtx(metric_name string, managed_type, managed_id string) (MContext, error) {
	query_params := map[string]string{}

	query_params["type"] = managed_type
	query_params["id"] = managed_id
	query_params["uid"] = managed_id
	query_params["metric-name"] = metric_name

	mo, e := self.mo_cache.Get(managed_id)
	if nil != e {
		return nil, errors.New(managed_type + " with id was '" + managed_id + "' is not found, " + e.Error())
	}
	if nil == mo {
		return nil, errors.New(managed_type + " with id was '" + managed_id + "' is not found.")
	}

	// switch mo["type"] {
	// case "network_device_port":
	// 	mo, e = self.mo_cache.Get(fmt.Sprint(mo["device_id"]))
	// 	if nil != e {
	// 		return nil, errors.New("device with id was '" + fmt.Sprint(mo["device_id"]) + "' is not found, " + e.Error())
	// 	}
	// 	if nil == mo {
	// 		return nil, errors.New("device with id was '" + fmt.Sprint(mo["device_id"]) + "' is not found.")
	// 	}
	// }

	return &context{params: query_params,
		managed_type: managed_type,
		managed_id:   managed_id,
		mo:           mo,
		local:        make(map[string]map[string]interface{}),
		alias:        alias_names,
		pry:          &proxy{srv: self},
		body_reader:  nil}, nil
}

func (self *server) Get(metric_name string, paths []P, params MContext) commons.Result {
	return invoke(self.route_for_get, self.routes_for_get, metric_name, paths, params)
}

func (self *server) Put(metric_name string, paths []P, params MContext) commons.Result {
	return invoke(self.route_for_put, self.routes_for_put, metric_name, paths, params)
}

func (self *server) Create(metric_name string, paths []P, params MContext) commons.Result {
	return invoke(self.route_for_create, self.routes_for_create, metric_name, paths, params)
}

func (self *server) Delete(metric_name string, paths []P, params MContext) commons.Result {
	return invoke(self.route_for_delete, self.routes_for_delete, metric_name, paths, params)
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
