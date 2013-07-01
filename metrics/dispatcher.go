package metrics

import (
	"commons"
	"errors"
	"flag"
)

var route_debuging = flag.Bool("dispatch.route.debugging", false, "set max size of pdu")

type dispatcher struct {
	for_get    map[string]*Routers
	for_put    map[string]*Routers
	for_create map[string]*Routers
	for_delete map[string]*Routers

	drvManager *commons.DriverManager
}

func newDispatcher(params map[string]interface{}) (*dispatcher, error) {
	dispatch := &dispatcher{for_get: make(map[string]*Routers),
		for_put:    make(map[string]*Routers),
		for_create: make(map[string]*Routers),
		for_delete: make(map[string]*Routers)}

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
		route, _ := self.for_get[rs.name]
		if nil == route {
			route = &Routers{}
			self.for_get[rs.name] = route
		}

		return route.register(rs)
	case "put", "Put", "PUT":
		route, _ := self.for_put[rs.name]
		if nil == route {
			route = &Routers{}
			self.for_put[rs.name] = route
		}

		return route.register(rs)
	case "create", "Create", "CREATE":
		route, _ := self.for_create[rs.name]
		if nil == route {
			route = &Routers{}
			self.for_create[rs.name] = route
		}

		return route.register(rs)
	case "delete", "Delete", "DELETE":
		route, _ := self.for_delete[rs.name]
		if nil == route {
			route = &Routers{}
			self.for_delete[rs.name] = route
		}

		return route.register(rs)
	default:
		return errors.New("Unsupported method - " + rs.definition.Method)
	}
}

func (self *dispatcher) unregister(name, id string) {
	for _, instances := range []map[string]*Routers{self.for_get,
		self.for_put, self.for_create, self.for_delete} {
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
	self.for_get = make(map[string]*Routers)
	self.for_put = make(map[string]*Routers)
	self.for_create = make(map[string]*Routers)
	self.for_delete = make(map[string]*Routers)
}

func notAcceptable(metric_name string) commons.Result {
	return commons.ReturnWithNotAcceptable("'" + metric_name + "' is not acceptable.")
}

func (self *dispatcher) Get(metric_name string, params commons.Map) commons.Result {
	route := self.for_get[metric_name]
	if nil == route {
		return notAcceptable(metric_name)
	}

	return route.Invoke(params)
}

func (self *dispatcher) Put(metric_name string, params commons.Map) commons.Result {
	route := self.for_put[metric_name]
	if nil == route {
		return notAcceptable(metric_name)
	}

	return route.Invoke(params)
}

func (self *dispatcher) Create(metric_name string, params commons.Map) commons.Result {
	route := self.for_create[metric_name]
	if nil == route {
		return notAcceptable(metric_name)
	}

	return route.Invoke(params)
}

func (self *dispatcher) Delete(metric_name string, params commons.Map) commons.Result {
	route := self.for_delete[metric_name]
	if nil == route {
		return notAcceptable(metric_name)
	}

	return route.Invoke(params)
}
