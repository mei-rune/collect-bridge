package metrics

import (
	"commons"
)

type memory struct {
	dispatcherBase
}

func (self *memory) Init(params map[string]interface{}, drvName string) commons.RuntimeError {
	e := self.dispatcherBase.Init(params, drvName)
	if nil != e {
		return e
	}
	self.RegisterGetFunc([]string{"1.3.6.1.4.1.9"}, func(params map[string]string) (commons.Result, commons.RuntimeError) {
		return self.GetCisco(params)
	})
	self.get = func(params map[string]string) (commons.Result, commons.RuntimeError) {
		return self.GetWindows(params)
	}
	return nil
}

func (self *memory) GetCisco(params map[string]string) (commons.Result, commons.RuntimeError) {
	res, e := self.GetCiscoA(params)
	if nil == e {
		return res, e
	}
	res, e = self.GetCiscoB(params)
	if nil == e {
		return res, e
	}
	return self.GetCiscoHost(params)
}

func (self *memory) GetCiscoHost(params map[string]string) (commons.Result, commons.RuntimeError) {
	return self.GetWindows(params)
}
func (self *memory) GetCiscoA(params map[string]string) (commons.Result, commons.RuntimeError) {
	_, total, e := self.GetInt64Value(params, "1.3.6.1.4.1.9.3.6.6.0", -1)
	if nil == e {
		return nil, e
	}
	_, free, e := self.GetInt64Value(params, "1.3.6.1.4.1.9.2.1.8.0", -1)
	if nil == e {
		return nil, e
	}

	return commons.Return(map[string]interface{}{"total": total, "used_per": float64(total-free) / float64(total), "used": total - free, "free": free}), nil
}

func (self *memory) GetCiscoB(params map[string]string) (commons.Result, commons.RuntimeError) {
	_, used, e := self.GetInt64Value(params, "1.3.6.1.4.1.9.9.109.1.1.1.1.12.1", -1)
	if nil == e {
		return nil, e
	}
	_, free, e := self.GetInt64Value(params, "1.3.6.1.4.1.9.9.109.1.1.1.1.13.1", -1)
	if nil == e {
		return nil, e
	}

	return commons.Return(map[string]interface{}{"total": used + free, "used_per": float64(used) / float64(used+free), "used": used, "free": free}), nil
}

func (self *memory) GetWindows(params map[string]string) (commons.Result, commons.RuntimeError) {
	//HOST-RESOURCES-MIB:hrStorageTable  = ".1.3.6.1.2.1.25.2.3.1.";
	//HOST-RESOURCES-MIB:hrMemorySize  = ".1.3.6.1.2.1.25.2.2.0";
	//Physical Memory type = "1.3.6.1.2.1.25.2.1.2";

	used_per := 0.0
	_, _, e := self.GetOneValue(params, "1.3.6.1.2.1.25.2.3.1", "2,5,6",
		func(old_row map[string]interface{}) (map[string]interface{}, commons.RuntimeError) {
			if "1.3.6.1.2.1.25.2.1.2" != GetOid(params, old_row, "2") {
				return nil, nil
			}
			x := GetInt32(params, old_row, "6", 0)
			y := GetInt32(params, old_row, "5", 0)
			used_per = float64(x) / float64(y)
			return emptyResult, nil
		})

	if nil != e {
		return nil, e
	}

	//HOST-RESOURCES-MIB:hrMemorySize  = ".1.3.6.1.2.1.25.2.2.0";
	_, total, e := self.GetUint64Value(params, "1.3.6.1.2.1.25.2.2.0", 0)
	if nil != e {
		return nil, e
	}

	used := uint64(float64(total) * used_per)
	free := total - used
	return commons.Return(map[string]interface{}{"total": total, "used_per": used_per, "used": used, "free": free}), nil
}

func init() {
	commons.METRIC_DRVS["mem"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &memory{}
		return drv, drv.Init(params, "snmp")
	}
}
