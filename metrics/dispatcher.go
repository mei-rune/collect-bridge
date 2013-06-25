package metrics

import (
	"commons"
	"errors"
	"flag"
)

var route_debuging = flag.Bool("dispatch.route.debugging", false, "set max size of pdu")

type Dispatcher struct {
	for_get    map[string]*Route
	for_put    map[string]*Route
	for_create map[string]*Route
	for_delete map[string]*Route
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{for_get: make(map[string]*Route),
		for_put:    make(map[string]*Route),
		for_create: make(map[string]*Route),
		for_delete: make(map[string]*Route)}
}

func (self *Dispatcher) registerSpec(rs *RouteSpec) error {
	switch rs.definition.Method {
	case "get", "Get", "GET":
		route, _ := self.for_get[rs.name]
		if nil == route {
			route = &Route{}
			self.for_get[rs.name] = route
		}

		return route.registerSpec(rs)
	case "put", "Put", "PUT":
		route, _ := self.for_put[rs.name]
		if nil == route {
			route = &Route{}
			self.for_put[rs.name] = route
		}

		return route.registerSpec(rs)
	case "create", "Create", "CREATE":
		route, _ := self.for_create[rs.name]
		if nil == route {
			route = &Route{}
			self.for_create[rs.name] = route
		}

		return route.registerSpec(rs)
	case "delete", "Delete", "DELETE":
		route, _ := self.for_delete[rs.name]
		if nil == route {
			route = &Route{}
			self.for_delete[rs.name] = route
		}

		return route.registerSpec(rs)
	default:
		return errors.New("Unsupported method - " + rs.definition.Method)
	}
}

func (self *Dispatcher) unregisterSpec(name, id string) {
	for _, instances := range []map[string]*Route{self.for_get,
		self.for_put, self.for_create, self.for_delete} {
		if "" == name {
			for _, route := range instances {
				route.unregisterSpec(id)
			}
		} else {
			route, _ := instances[name]
			if nil == route {
				return
			}
			route.unregisterSpec(id)
		}
	}
}

func (self *Dispatcher) clear() {
	self.for_get = make(map[string]*Route)
	self.for_put = make(map[string]*Route)
	self.for_create = make(map[string]*Route)
	self.for_delete = make(map[string]*Route)
}

func (self *Dispatcher) Get(params commons.Map) commons.Result {
	metric_name := params.GetString("metric_name", "")
	if 0 == len(metric_name) {
		return commons.ReturnWithIsRequired("metric_name")
	}

	route := self.for_get[metric_name]
	if nil == route {
		return commons.ReturnWithNotAcceptable(metric_name)
	}

	return route.Invoke(params)
}

func (self *Dispatcher) Put(params commons.Map) commons.Result {
	metric_name := params.GetString("metric_name", "")
	if 0 == len(metric_name) {
		return commons.ReturnWithIsRequired("metric_name")
	}

	route := self.for_put[metric_name]
	if nil == route {
		return commons.ReturnWithNotAcceptable(metric_name)
	}

	return route.Invoke(params)
}

func (self *Dispatcher) Create(params commons.Map) commons.Result {
	metric_name := params.GetString("metric_name", "")
	if 0 == len(metric_name) {
		return commons.ReturnWithIsRequired("metric_name")
	}

	route := self.for_create[metric_name]
	if nil == route {
		return commons.ReturnWithNotAcceptable(metric_name)
	}

	return route.Invoke(params)
}

func (self *Dispatcher) Delete(params commons.Map) commons.Result {
	metric_name := params.GetString("metric_name", "")
	if 0 == len(metric_name) {
		return commons.ReturnWithIsRequired("metric_name")
	}

	route := self.for_delete[metric_name]
	if nil == route {
		return commons.ReturnWithNotAcceptable(metric_name)
	}

	return route.Invoke(params)
}
