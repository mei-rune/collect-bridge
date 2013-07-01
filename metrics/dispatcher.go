package metrics

import (
	"commons"
	"errors"
	"flag"
)

var route_debuging = flag.Bool("dispatch.route.debugging", false, "set max size of pdu")

type dispatcher struct {
	routes_for_get    map[string]*Routers
	routes_for_put    map[string]*Routers
	routes_for_create map[string]*Routers
	routes_for_delete map[string]*Routers

	route_for_get    map[string]*Route
	route_for_put    map[string]*Route
	route_for_create map[string]*Route
	route_for_delete map[string]*Route

	drvManager *commons.DriverManager
}

func newDispatcher(params map[string]interface{}) (*dispatcher, error) {
	dispatch := &dispatcher{routes_for_get: make(map[string]*Routers),
		routes_for_put:    make(map[string]*Routers),
		routes_for_create: make(map[string]*Routers),
		routes_for_delete: make(map[string]*Routers),
		route_for_get:     make(map[string]*Route),
		route_for_put:     make(map[string]*Route),
		route_for_create:  make(map[string]*Route),
		route_for_delete:  make(map[string]*Route)}

	for k, rs := range Methods {
		r, e := newRouteWithSpec(k, rs, params)
		if nil != e {
			return nil, errors.New("init '" + k + "' failed, " + e.Error())
		}

		e = dispatch.register(r)
		if nil != e {
			return nil, errors.New("register '" + k + "' failed, " + e.Error())
		}
	}

	return dispatch, nil
}

func (self *dispatcher) register(rs *Route) error {
	switch rs.definition.Method {
	case "get", "Get", "GET":
		if _, ok := self.route_for_get[rs.id]; ok {
			return errors.New("route that id is  '" + rs.id + "' is already exists.")
		}
		self.route_for_get[rs.id] = rs

		route, _ := self.routes_for_get[rs.name]
		if nil == route {
			route = &Routers{}
			self.routes_for_get[rs.name] = route
		}

		return route.register(rs)
	case "put", "Put", "PUT":
		if _, ok := self.route_for_put[rs.id]; ok {
			return errors.New("route that id is  '" + rs.id + "' is already exists.")
		}
		self.route_for_put[rs.id] = rs

		route, _ := self.routes_for_put[rs.name]
		if nil == route {
			route = &Routers{}
			self.routes_for_put[rs.name] = route
		}

		return route.register(rs)
	case "create", "Create", "CREATE":
		if _, ok := self.route_for_create[rs.id]; ok {
			return errors.New("route that id is  '" + rs.id + "' is already exists.")
		}
		self.route_for_create[rs.id] = rs

		route, _ := self.routes_for_create[rs.name]
		if nil == route {
			route = &Routers{}
			self.routes_for_create[rs.name] = route
		}

		return route.register(rs)
	case "delete", "Delete", "DELETE":
		if _, ok := self.route_for_delete[rs.id]; ok {
			return errors.New("route that id is  '" + rs.id + "' is already exists.")
		}
		self.route_for_delete[rs.id] = rs

		route, _ := self.routes_for_delete[rs.name]
		if nil == route {
			route = &Routers{}
			self.routes_for_delete[rs.name] = route
		}

		return route.register(rs)
	default:
		return errors.New("Unsupported method - " + rs.definition.Method)
	}
}

func (self *dispatcher) unregister(name, id string) {
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

func (self *dispatcher) clear() {
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

func (self *dispatcher) Get(metric_name string, params commons.Map) commons.Result {
	route := self.route_for_get[metric_name]
	if nil != route {
		return route.Invoke(params)
	}

	routes := self.routes_for_get[metric_name]
	if nil == routes {
		return notAcceptable(metric_name)
	}

	return routes.Invoke(params)
}

func (self *dispatcher) Put(metric_name string, params commons.Map) commons.Result {
	route := self.route_for_put[metric_name]
	if nil != route {
		return route.Invoke(params)
	}

	routes := self.routes_for_put[metric_name]
	if nil == routes {
		return notAcceptable(metric_name)
	}

	return routes.Invoke(params)
}

func (self *dispatcher) Create(metric_name string, params commons.Map) commons.Result {
	route := self.route_for_create[metric_name]
	if nil != route {
		return route.Invoke(params)
	}

	routes := self.routes_for_create[metric_name]
	if nil == routes {
		return notAcceptable(metric_name)
	}

	return routes.Invoke(params)
}

func (self *dispatcher) Delete(metric_name string, params commons.Map) commons.Result {
	route := self.route_for_delete[metric_name]
	if nil != route {
		return route.Invoke(params)
	}

	routes := self.routes_for_delete[metric_name]
	if nil == routes {
		return notAcceptable(metric_name)
	}

	return routes.Invoke(params)
}
