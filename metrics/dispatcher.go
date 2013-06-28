package metrics

import (
	"commons"
	"errors"
	"flag"
	"fmt"
)

var route_debuging = flag.Bool("dispatch.route.debugging", false, "set max size of pdu")

type dispatcher struct {
	for_get    map[string]*Route
	for_put    map[string]*Route
	for_create map[string]*Route
	for_delete map[string]*Route

	drvManager *commons.DriverManager
}

func newdispatcher(params map[string]interface{}) (*dispatcher, error) {
	dispatch := &dispatcher{for_get: make(map[string]*Route),
		for_put:    make(map[string]*Route),
		for_create: make(map[string]*Route),
		for_delete: make(map[string]*Route)}

	for k, f := range Methods {
		method, e := f(params)
		if nil != e {
			return nil, errors.New("init '" + k + "' failed, " + e.Error())
		}

		rs, e := NewRouteSpec(&RouteDefinition{Name: k, Method: "get"})
		if nil != e {
			return nil, errors.New("init '" + k + "' failed, " + e.Error())
		}
		rs.invoke = method

		e = dispatch.registerSpec(rs)
		if nil != e {
			return nil, errors.New("init '" + k + "' failed, " + e.Error())
		}
	}
	// drv := &systemInfo{}
	// drv.Init(map[string]interface{}{"snmp": snmp.NewSnmpDriver(10*time.Second, nil)}, "snmp")
	// rs, _ := NewRouteSpec(&RouteDefinition{Name: "sys", Method: "get"})
	// rs.invoke = func(params commons.Map) commons.Result {
	// 	drv.Get(params)
	// }
	// dispatch.registerSpec(rs)

	return dispatch, nil
}

func (self *dispatcher) registerSpec(rs *RouteSpec) error {
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

func (self *dispatcher) unregisterSpec(name, id string) {
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

func (self *dispatcher) clear() {
	self.for_get = make(map[string]*Route)
	self.for_put = make(map[string]*Route)
	self.for_create = make(map[string]*Route)
	self.for_delete = make(map[string]*Route)
}

func notAcceptable(metric_name string) commons.Result {
	return commons.ReturnWithNotAcceptable("'" + metric_name + "' is not acceptable.")
}

func (self *dispatcher) Get(metric_name string, params commons.Map) commons.Result {
	route := self.for_get[metric_name]
	if nil == route {
		fmt.Println("Get========")
		for k, _ := range self.for_get {
			fmt.Println(k)
		}
		return notAcceptable(metric_name)
	}

	return route.Invoke(params)
}

func (self *dispatcher) Put(metric_name string, params commons.Map) commons.Result {
	route := self.for_put[metric_name]
	if nil == route {
		fmt.Println("Put========")
		for k, _ := range self.for_get {
			fmt.Println(k)
		}
		return notAcceptable(metric_name)
	}

	return route.Invoke(params)
}

func (self *dispatcher) Create(metric_name string, params commons.Map) commons.Result {
	route := self.for_create[metric_name]
	if nil == route {
		fmt.Println("Create========")
		for k, _ := range self.for_get {
			fmt.Println(k)
		}

		return notAcceptable(metric_name)
	}

	return route.Invoke(params)
}

func (self *dispatcher) Delete(metric_name string, params commons.Map) commons.Result {
	route := self.for_delete[metric_name]
	if nil == route {
		fmt.Println("Delete========")
		for k, _ := range self.for_get {
			fmt.Println(k)
		}
		return notAcceptable(metric_name)
	}

	return route.Invoke(params)
}
