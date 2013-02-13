package metrics

import (
	"commons"
)

type cpu struct {
	dispatcherBase
}

func (self *cpu) Init(params map[string]interface{}, drvName string) commons.RuntimeError {
	e := self.dispatcherBase.Init(params, drvName)
	if nil != e {
		return e
	}

	self.RegisterGetFunc([]string{"1.3.6.1.4.1.5655"}, func(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
		return self.GetCiscoSCE(params)
	})
	self.RegisterGetFunc([]string{"1.3.6.1.4.1.9"}, func(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
		return self.GetCisco(params)
	})
	self.RegisterGetFunc([]string{"1.3.6.1.4.1.9.1.746"}, func(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
		return self.GetCiscoHost(params)
	})
	self.RegisterGetFunc([]string{"1.3.6.1.4.1.9.12.3.1.3"}, func(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
		return self.GetCiscoSAN(params)
	})
	self.RegisterGetFunc([]string{"1.3.6.1.4.1.9.1.965", "1.3.6.1.4.1.9.1.966", "1.3.6.1.4.1.9.1.967"},
		func(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
			return self.GetCiscoSCE(params)
		})

	self.get = func(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
		return self.GetWindows(params)
	}
	return nil
}

func (self *cpu) GetCiscoHost(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return self.GetWindows(params)
}

func (self *cpu) GetCiscoSAN(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	_, cpu, e := self.GetInt32Value(params, "1.3.6.1.4.1.9.9.305.1.1.2.0", -1)
	if nil == e {
		return map[string]interface{}{"cpu": cpu}, nil
	}
	return nil, commons.ContinueError
}
func (self *cpu) GetCiscoSCE(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	cpu := int32(-1)
	_, _, e := self.GetOneValue(params, "1.3.6.1.4.1.5655.4.1.9.1.1", "35",
		func(old_row map[string]interface{}) (map[string]interface{}, commons.RuntimeError) {

			if i, _ := TryGetInt32(params, old_row, "35", -1); -1 != i {
				cpu = i
				return emptyResult, nil
			}

			return nil, nil
		})
	if nil == e {
		return map[string]interface{}{"cpu": cpu}, nil
	}
	return nil, commons.ContinueError
}

func (self *cpu) GetCisco(params map[string]string) (map[string]interface{}, commons.RuntimeError) {

	res, i, e := self.GetInt32Value(params, "1.3.6.1.4.1.9.2.1.57.0", -1)
	if nil == e {
		return commons.ReturnWithValue(res, map[string]interface{}{"cpu": i}), nil
	}

	cpu := int32(-1)
	_, _, e = self.GetOneValue(params, "1.3.6.1.4.1.9.9.109.1.1.1.1", "4,7",
		func(old_row map[string]interface{}) (map[string]interface{}, commons.RuntimeError) {

			if i, _ := TryGetInt32(params, old_row, "4", -1); -1 != i {
				cpu = i
				return emptyResult, nil
			} else if i, _ := TryGetInt32(params, old_row, "7", -1); -1 != i {
				cpu = i
				return emptyResult, nil
			}

			return nil, nil
		})
	if nil == e {
		return map[string]interface{}{"cpu": cpu}, nil
	}
	return nil, e
}

func (self *cpu) GetWindows(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	cpus := make([]int, 0, 4)

	_, _, e := self.GetTableValue(params, "1.3.6.1.2.1.25.3.3.1", "",
		func(table map[string]interface{}, key string, old_row map[string]interface{}) error {
			cpus = append(cpus, int(GetInt32(params, old_row, "2", 0)))
			return nil
		})

	if nil != e {
		return nil, e
	}
	if 0 == len(cpus) {
		return nil, commons.NewRuntimeError(commons.InternalErrorCode, "cpu list is empty")
	}
	total := 0
	for _, v := range cpus {
		total += v
	}

	return map[string]interface{}{"cpu": total / len(cpus), "cpu_list": cpus}, nil
}

func init() {
	commons.METRIC_DRVS["cpu"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &cpu{}
		return drv, drv.Init(params, "snmp")
	}
}
