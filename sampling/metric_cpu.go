package sampling

import (
	"errors"
)

// type cpu struct {
// 	dispatcherBase
// }

// func (self *cpu) Init(params map[string]interface{}, drvName string)  error {
// 	e := self.dispatcherBase.Init(params, drvName)
// 	if nil != e {
// 		return e
// 	}

// 	self.RegisterGetFunc([]string{"1.3.6.1.4.1.5655"}, func(params MContext) (interface{}, error) {
// 		return self.GetCiscoSCE(params)
// 	})
// 	self.RegisterGetFunc([]string{"1.3.6.1.4.1.9"}, func(params MContext) (interface{}, error) {
// 		return self.GetCisco(params)
// 	})
// 	self.RegisterGetFunc([]string{"1.3.6.1.4.1.9.1.746"}, func(params MContext) (interface{}, error) {
// 		return self.GetCiscoHost(params)
// 	})
// 	self.RegisterGetFunc([]string{"1.3.6.1.4.1.9.12.3.1.3"}, func(params MContext) (interface{}, error) {
// 		return self.GetCiscoSAN(params)
// 	})
// 	self.RegisterGetFunc([]string{"1.3.6.1.4.1.9.1.965", "1.3.6.1.4.1.9.1.966", "1.3.6.1.4.1.9.1.967"},
// 		func(params MContext) (interface{}, error) {
// 			return self.GetCiscoSCE(params)
// 		})

// 	self.get = func(params MContext) (interface{}, error) {
// 		return self.GetWindows(params)
// 	}
// 	return nil
// }

type cpuCiscoSAN struct {
	snmpBase
}

func (self *cpuCiscoSAN) Call(params MContext) (interface{}, error) {
	cpu, e := self.GetInt32(params, "1.3.6.1.4.1.9.9.305.1.1.2.0")
	if nil != e {
		return nil, e
	}

	return map[string]interface{}{"cpu": cpu}, nil
}

type cpuCiscoSCE struct {
	snmpBase
}

func (self *cpuCiscoSCE) Call(params MContext) (interface{}, error) {
	cpu := int32(-1)
	e := self.OneInTable(params, "1.3.6.1.4.1.5655.4.1.9.1.1", "35",
		func(key string, old_row map[string]interface{}) (bool, error) {

			if i, _ := GetInt32(params, old_row, "35", -1); -1 != i {
				cpu = i
				return true, nil
			}

			return false, nil
		})

	if nil != e {
		return nil, e
	}

	return map[string]interface{}{"cpu": cpu}, nil
}

type cpuCiscoCpm struct {
	baseCisco
}

func (self *cpuCiscoCpm) Call(params MContext) (interface{}, error) {
	return self.readCpuWithCpmSytle(params)
}

type cpuCiscoOldStyle struct {
	baseCisco
}

func (self *cpuCiscoOldStyle) Call(params MContext) (interface{}, error) {
	return self.readCpuWithOldSytle(params)
}

type cpuCiscoSystemExt struct {
	baseCisco
}

func (self *cpuCiscoSystemExt) Call(params MContext) (interface{}, error) {
	return self.readCpuWithSystemExt(params)
}

type cpuHostResources struct {
	snmpBase
}

func (self *cpuHostResources) Call(params MContext) (interface{}, error) {
	// http://www.ietf.org/rfc/rfc1514.txt
	// hrProcessorTable OBJECT-TYPE
	//     SYNTAX SEQUENCE OF HrProcessorEntry
	//     ACCESS not-accessible
	//     STATUS mandatory
	//     DESCRIPTION
	//            "The (conceptual) table of processors contained by
	//            the host.
	//
	//            Note that this table is potentially sparse: a
	//            (conceptual) entry exists only if the correspondent
	//            value of the hrDeviceType object is
	//            `hrDeviceProcessor'."
	//     ::= { hrDevice 3 }
	//
	// hrProcessorEntry OBJECT-TYPE
	//     SYNTAX HrProcessorEntry
	//     ACCESS not-accessible
	//     STATUS mandatory
	//     DESCRIPTION
	//            "A (conceptual) entry for one processor contained
	//            by the host.  The hrDeviceIndex in the index
	//            represents the entry in the hrDeviceTable that
	//            corresponds to the hrProcessorEntry.
	//
	//            As an example of how objects in this table are
	//            named, an instance of the hrProcessorFrwID object
	//            might be named hrProcessorFrwID.3"
	//     INDEX { hrDeviceIndex }
	//     ::= { hrProcessorTable 1 }
	//
	// HrProcessorEntry ::= SEQUENCE {
	//         hrProcessorFrwID            ProductID,
	//         hrProcessorLoad             INTEGER
	//     }
	//
	// hrProcessorFrwID OBJECT-TYPE
	//     SYNTAX ProductID
	//     ACCESS read-only
	//     STATUS mandatory
	//     DESCRIPTION
	//            "The product ID of the firmware associated with the
	//            processor."
	//     ::= { hrProcessorEntry 1 }
	//
	// hrProcessorLoad OBJECT-TYPE
	//     SYNTAX INTEGER (0..100)
	//     ACCESS read-only
	//     STATUS mandatory
	//     DESCRIPTION
	//            "The average, over the last minute, of the
	//            percentage of time that this processor was not
	//            idle."
	//     ::= { hrProcessorEntry 2 }

	cpus := make([]int, 0, 4)

	e := self.EachInTable(params, "1.3.6.1.2.1.25.3.3.1", "",
		func(key string, old_row map[string]interface{}) error {
			cpus = append(cpus, int(GetInt32WithDefault(params, old_row, "2", 0)))
			return nil
		})

	if nil != e {
		return nil, e
	}

	switch len(cpus) {
	case 0:
		return nil, errors.New("cpu list is empty")
	case 1:
		return map[string]interface{}{"cpu": cpus[0], "cpu_list": cpus}, nil
	default:
		total := 0
		for _, v := range cpus {
			total += v
		}
		return map[string]interface{}{"cpu": total / len(cpus), "cpu_list": cpus}, nil
	}
}

func init() {
	Methods["default_cpu"] = newRouteSpec("get", "cpu", "default cpu", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &cpuHostResources{}
			return drv, drv.Init(params)
		})

	Methods["cisco_cpu_cpm_style"] = newRouteSpec("get", "cpu", "cisco cpu by cpm style", Match().Oid("1.3.6.1.4.1.9").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &cpuCiscoCpm{}
			return drv, drv.Init(params)
		})

	Methods["cisco_cpu_old_style"] = newRouteSpec("get", "cpu", "cisco cpu by old style", Match().Oid("1.3.6.1.4.1.9").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &cpuCiscoOldStyle{}
			return drv, drv.Init(params)
		})

	Methods["cisco_cpu_system_ext_style"] = newRouteSpec("get", "cpu", "cisco cpu by system ext style", Match().Oid("1.3.6.1.4.1.9").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &cpuCiscoSystemExt{}
			return drv, drv.Init(params)
		})

	Methods["cisco_host_cpu"] = newRouteSpec("get", "cpu", "cisco host cpu", Match().Oid("1.3.6.1.4.1.9.1.746").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &cpuHostResources{}
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

	Methods["h3c_cpu"] = newRouteSpec("get", "cpu", "the cpu of h3c", Match().Oid("1.3.6.1.4.1.25506").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &cpuH3C{}
			return drv, drv.Init(params)
		})

	Methods["foundry_cpu"] = newRouteSpec("get", "cpu", "the cpu of foundry", Match().Oid("1.3.6.1.4.1.1991").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &cpuFoundry{}
			return drv, drv.Init(params)
		})
}

type cpuH3C struct {
	baseH3C
}

func (self *cpuH3C) Call(params MContext) (interface{}, error) {
	return self.smartRead("memory", params, self.readCpuWithNewStyle, self.readCpuWithCompatibleStyle)
}

type cpuFoundry struct {
	baseFoundry
}

func (self *cpuFoundry) Call(params MContext) (interface{}, error) {
	return self.readCpu(params)
}
