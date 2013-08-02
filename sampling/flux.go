package sampling

import (
	"commons"
	"sync"
	"time"
)

type baseWorker struct {
	id        string
	timestamp int64
	l         sync.Mutex

	bw          *backgroundWorker
	results     circularBuffer
	dispatcher  *dispatcher
	metric_name string
	ctx         *context
	invoke_cb   invoke_func
}

func (self *baseWorker) isTimeout(now int64, default_interval time.Duration) bool {
	self.l.Lock()
	defer self.l.Unlock()
	return (self.timestamp + int64(default_interval.Seconds())) <= now
}

func (self *baseWorker) invoke() error {
	self.l.Lock()
	defer self.l.Unlock()

	res := self.invoke_cb(self.dispatcher, self.metric_name, self.ctx)
	self.results.push(res)
	if res.HasError() {
		return res.Error()
	}
	return nil
}

func (self *baseWorker) close() {
	self.bw.remove(self.id)
}

func (self *baseWorker) stats() interface{} {
	return nil
}

// func (self *baseWorker) get() bool {
// 	res := self.invoke(self.dispatcher, self.metric_name, self.ctx)
//   return
// }

// type flux_worker struct {
// 	baseWorker
// }

// func (self *flux_worker) get() (interface{}, error) {
// }

type interface_flux_create struct {
	bw *backgroundWorker
}

func (self *interface_flux_create) Call(ctx MContext) commons.Result {
	params := ctx.GetStringWithDefault("metric-params", "")
	if 0 == len(params) {
		return commons.ReturnError(commons.BadRequestCode, "metric-params is missing.")
	}

	return nil
}

// func init() {
// 	bw := &backgroundWorker{name: "default",
// 		c:                make(chan func()),
// 		period_interval:  *period_interval,
// 		timeout_interval: timeout_interval,
// 		workers:          make(map[string]worker)}

// 	go bw.run()

// 	Methods["default_interface_flux_create"] = newRouteSpec("create", "interface_flux", "interface_flux", nil,
// 		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
// 			return &interface_flux_create{bw: bw}, nil
// 		})

// 	Methods["default_interface_flux"] = newRouteSpec("get", "interface_flux", "interface_flux", nil,
// 		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
// 			drv := &memWindows{}
// 			return drv, drv.Init(params)
// 		})
// 	Methods["default_link_flux"] = newRouteSpec("get", "link_flux", "interface_flux", Match().Oid("1.3.6.1.4.1.9").Build(),
// 		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
// 			drv := &memCisco{}
// 			return drv, drv.Init(params)
// 		})
// }
