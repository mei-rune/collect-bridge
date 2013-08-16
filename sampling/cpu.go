package sampling

import (
	"commons"
)

// type cpu struct {
// 	dispatcherBase
// }

// func (self *cpu) Init(params map[string]interface{}, drvName string)  error {
// 	e := self.dispatcherBase.Init(params, drvName)
// 	if nil != e {
// 		return e
// 	}

// 	self.RegisterGetFunc([]string{"1.3.6.1.4.1.5655"}, func(params MContext) commons.Result {
// 		return self.GetCiscoSCE(params)
// 	})
// 	self.RegisterGetFunc([]string{"1.3.6.1.4.1.9"}, func(params MContext) commons.Result {
// 		return self.GetCisco(params)
// 	})
// 	self.RegisterGetFunc([]string{"1.3.6.1.4.1.9.1.746"}, func(params MContext) commons.Result {
// 		return self.GetCiscoHost(params)
// 	})
// 	self.RegisterGetFunc([]string{"1.3.6.1.4.1.9.12.3.1.3"}, func(params MContext) commons.Result {
// 		return self.GetCiscoSAN(params)
// 	})
// 	self.RegisterGetFunc([]string{"1.3.6.1.4.1.9.1.965", "1.3.6.1.4.1.9.1.966", "1.3.6.1.4.1.9.1.967"},
// 		func(params MContext) commons.Result {
// 			return self.GetCiscoSCE(params)
// 		})

// 	self.get = func(params MContext) commons.Result {
// 		return self.GetWindows(params)
// 	}
// 	return nil
// }

type cpuCiscoSAN struct {
	snmpBase
}

func (self *cpuCiscoSAN) Call(params MContext) commons.Result {
	cpu, e := self.GetInt32(params, "1.3.6.1.4.1.9.9.305.1.1.2.0")
	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}

	return commons.Return(map[string]interface{}{"cpu": cpu})
}

type cpuCiscoSCE struct {
	snmpBase
}

func (self *cpuCiscoSCE) Call(params MContext) commons.Result {
	cpu := int32(-1)
	e := self.OneInTable(params, "1.3.6.1.4.1.5655.4.1.9.1.1", "35",
		func(key string, old_row map[string]interface{}) error {

			if i, _ := TryGetInt32(params, old_row, "35", -1); -1 != i {
				cpu = i
				return nil
			}

			return commons.ContinueError
		})

	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}

	return commons.Return(map[string]interface{}{"cpu": cpu})
}

type cpuCisco struct {
	snmpBase
}

func (self *cpuCisco) Call(params MContext) commons.Result {
	i, e := self.GetInt32(params, "1.3.6.1.4.1.9.2.1.57.0")
	if nil == e {
		return commons.Return(map[string]interface{}{"cpu": i})
	}

	cpu := int32(-1)
	e = self.OneInTable(params, "1.3.6.1.4.1.9.9.109.1.1.1.1", "4,7",
		func(key string, old_row map[string]interface{}) error {
			if i, _ := TryGetInt32(params, old_row, "4", -1); -1 != i {
				cpu = i
				return nil
			} else if i, _ := TryGetInt32(params, old_row, "7", -1); -1 != i {
				cpu = i
				return nil
			}

			return commons.ContinueError
		})

	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}

	return commons.Return(map[string]interface{}{"cpu": cpu})
}

type cpuWindows struct {
	snmpBase
}

func (self *cpuWindows) Call(params MContext) commons.Result {
	cpus := make([]int, 0, 4)

	e := self.EachInTable(params, "1.3.6.1.2.1.25.3.3.1", "",
		func(key string, old_row map[string]interface{}) error {
			cpus = append(cpus, int(GetInt32(params, old_row, "2", 0)))
			return nil
		})

	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}

	switch len(cpus) {
	case 0:
		return commons.ReturnError(commons.InternalErrorCode, "cpu list is empty")
	case 1:
		return commons.Return(map[string]interface{}{"cpu": cpus[0], "cpu_list": cpus})
	default:
		total := 0
		for _, v := range cpus {
			total += v
		}
		return commons.Return(map[string]interface{}{"cpu": total / len(cpus), "cpu_list": cpus})
	}
}

func init() {
	Methods["default_cpu"] = newRouteSpec("get", "cpu", "default cpu", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &cpuWindows{}
			return drv, drv.Init(params)
		})

	Methods["cisco_cpu"] = newRouteSpec("get", "cpu", "default cpu", Match().Oid("1.3.6.1.4.1.9").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &cpuCisco{}
			return drv, drv.Init(params)
		})

	Methods["cisco_host_cpu"] = newRouteSpec("get", "cpu", "default cpu", Match().Oid("1.3.6.1.4.1.9.1.746").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &cpuWindows{}
			return drv, drv.Init(params)
		})

	Methods["cisco_sce_cpu"] = newRouteSpec("get", "cpu", "the cpu of cisco sce", Match().Oid("1.3.6.1.4.1.5655").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &cpuCiscoSCE{}
			return drv, drv.Init(params)
		})

	Methods["cisco_sce_cpu_965"] = newRouteSpec("get", "cpu", "the cpu of cisco sce", Match().Oid("1.3.6.1.4.1.9.1.965").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &cpuCiscoSCE{}
			return drv, drv.Init(params)
		})

	Methods["cisco_sce_cpu_966"] = newRouteSpec("get", "cpu", "the cpu of cisco sce", Match().Oid("1.3.6.1.4.1.9.1.966").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &cpuCiscoSCE{}
			return drv, drv.Init(params)
		})

	Methods["cisco_sce_cpu_967"] = newRouteSpec("get", "cpu", "the cpu of cisco sce", Match().Oid("1.3.6.1.4.1.9.1.967").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &cpuCiscoSCE{}
			return drv, drv.Init(params)
		})

	Methods["cisco_san_cpu"] = newRouteSpec("get", "cpu", "the cpu of cisco san", Match().Oid("1.3.6.1.4.1.9.12.3.1.3").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &cpuCiscoSAN{}
			return drv, drv.Init(params)
		})

	Methods["h3c_cpu"] = newRouteSpec("get", "cpu", "the generic cpu of h3c", Match().Oid("1.3.6.1.4.1.25506").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &cpuH3C{}
			return drv, drv.Init(params)
		})
}

type cpuH3C struct {
	baseH3C
}

func (self *cpuH3C) Call(params MContext) commons.Result {
	return self.smartRead("memory", params, self.readCpuWithNewStyle, self.readCpuWithCompatibleStyle)
}
