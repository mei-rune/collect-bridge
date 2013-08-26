package sampling

import (
	// "commons"
	// "errors"
	"flag"
)

var route_debuging = flag.Bool("dispatch.route.debugging", false, "set max size of pdu")

// type dispatcher struct {

// 	drvManager *commons.DriverManager
// }

// func invoke(route_by_id map[string]*Route, routes_by_name map[string]*Routes,
// 	metric_name string, paths [][2]string, params MContext) commons.Result {
// 	route := route_by_id[metric_name]
// 	if nil != route {
// 		return route.Invoke(params)
// 	}

// 	routes := routes_by_name[metric_name]
// 	if nil == routes {
// 		return notAcceptable(metric_name)
// 	}

// 	return routes.Invoke(params)
// }

// func (self *dispatcher) Get(metric_name string, paths [][2]string, params MContext) commons.Result {
// 	return invoke self.route_for_get[metric_name]
// 	if nil != route {
// 		return route.Invoke(params)
// 	}

// 	routes := self.routes_for_get[metric_name]
// 	if nil == routes {
// 		return notAcceptable(metric_name)
// 	}

// 	return routes.Invoke(params)
// }

// func (self *dispatcher) Put(metric_name string, paths [][2]string, params MContext) commons.Result {
// 	route := self.route_for_put[metric_name]
// 	if nil != route {
// 		return route.Invoke(params)
// 	}

// 	routes := self.routes_for_put[metric_name]
// 	if nil == routes {
// 		return notAcceptable(metric_name)
// 	}

// 	return routes.Invoke(params)
// }

// func (self *dispatcher) Create(metric_name string, paths [][2]string, params MContext) commons.Result {
// 	route := self.route_for_create[metric_name]
// 	if nil != route {
// 		return route.Invoke(params)
// 	}

// 	routes := self.routes_for_create[metric_name]
// 	if nil == routes {
// 		return notAcceptable(metric_name)
// 	}

// 	return routes.Invoke(params)
// }

// func (self *dispatcher) Delete(metric_name string, paths [][2]string, params MContext) commons.Result {
// 	route := self.route_for_delete[metric_name]
// 	if nil != route {
// 		return route.Invoke(params)
// 	}

// 	routes := self.routes_for_delete[metric_name]
// 	if nil == routes {
// 		return notAcceptable(metric_name)
// 	}

// 	return routes.Invoke(params)
// }
