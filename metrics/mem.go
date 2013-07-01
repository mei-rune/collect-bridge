package metrics

import (
	"commons"
)

type memCisco struct {
	snmpBase
}

func (self *memCisco) Call(params commons.Map) commons.Result {
	res := self.CallA(params)
	if !res.HasError() {
		return res
	}
	res = self.CallB(params)
	if !res.HasError() {
		return res
	}
	return self.CallHost(params)
}

func (self *memCisco) CallHost(params commons.Map) commons.Result {
	var windows memWindows
	windows.CopyFrom(&self.snmpBase)
	return windows.Call(params)
}

func (self *memCisco) CallA(params commons.Map) commons.Result {
	total, e := self.GetInt64(params, "1.3.6.1.4.1.9.3.6.6.0")
	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}

	free, e := self.GetInt64(params, "1.3.6.1.4.1.9.2.1.8.0")
	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}

	return commons.Return(map[string]interface{}{"total": total,
		"used_per": float64(total-free) / float64(total),
		"used":     total - free,
		"free":     free})
}

func (self *memCisco) CallB(params commons.Map) commons.Result {
	used, e := self.GetInt64(params, "1.3.6.1.4.1.9.9.109.1.1.1.1.12.1")
	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}

	free, e := self.GetInt64(params, "1.3.6.1.4.1.9.9.109.1.1.1.1.13.1")
	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}

	return commons.Return(map[string]interface{}{"total": used + free,
		"used_per": float64(used) / float64(used+free),
		"used":     used,
		"free":     free})
}

type memWindows struct {
	snmpBase
}

func (self *memWindows) Call(params commons.Map) commons.Result {
	//HOST-RESOURCES-MIB:hrStorageTable  = ".1.3.6.1.2.1.25.2.3.1.";
	//HOST-RESOURCES-MIB:hrMemorySize  = ".1.3.6.1.2.1.25.2.2.0";
	//Physical Memory type = "1.3.6.1.2.1.25.2.1.2";

	used_per := 0.0
	e := self.OneInTable(params, "1.3.6.1.2.1.25.2.3.1", "2,5,6",
		func(key string, old_row map[string]interface{}) error {
			if "1.3.6.1.2.1.25.2.1.2" != GetOid(params, old_row, "2") {
				return commons.ContinueError
			}
			x := GetInt32(params, old_row, "6", 0)
			y := GetInt32(params, old_row, "5", 0)
			used_per = float64(x) / float64(y)
			return nil
		})

	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}

	//HOST-RESOURCES-MIB:hrMemorySize  = ".1.3.6.1.2.1.25.2.2.0";
	total, e := self.GetUint64(params, "1.3.6.1.2.1.25.2.2.0")
	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}

	used := uint64(float64(total) * used_per)
	free := total - used
	return commons.Return(map[string]interface{}{"total": total, "used_per": used_per, "used": used, "free": free})
}

func init() {
	Methods["default_mem"] = newRouteSpec("mem", "default mem", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &memWindows{}
			return drv, drv.Init(params)
		})
	Methods["cisco_mem"] = newRouteSpec("mem", "the mem of cisco", Match().Oid("1.3.6.1.4.1.9").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &memCisco{}
			return drv, drv.Init(params)
		})
}
