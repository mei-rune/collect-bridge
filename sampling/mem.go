package sampling

import (
	"commons"
)

type memCisco struct {
	snmpBase
}

func (self *memCisco) Call(params MContext) commons.Result {
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

func (self *memCisco) CallHost(params MContext) commons.Result {
	var windows memWindows
	windows.CopyFrom(&self.snmpBase)
	return windows.Call(params)
}

func (self *memCisco) CallA(params MContext) commons.Result {
	total, e := self.GetInt64(params, "1.3.6.1.4.1.9.3.6.6.0")
	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}

	free, e := self.GetInt64(params, "1.3.6.1.4.1.9.2.1.8.0")
	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}

	return commons.Return(map[string]interface{}{"total": total,
		"used_per": (float64(total-free) / float64(total)) * 100,
		"used":     total - free,
		"free":     free})
}

func (self *memCisco) CallB(params MContext) commons.Result {
	// ftp://ftp.cisco.com/pub/mibs/oid/CISCO-PROCESS-MIB.oid
	// "cpmCPUMemoryUsed"		"1.3.6.1.4.1.9.9.109.1.1.1.1.12"
	// "cpmCPUMemoryFree"		"1.3.6.1.4.1.9.9.109.1.1.1.1.13"
	// "cpmCPUMemoryKernelReserved"		"1.3.6.1.4.1.9.9.109.1.1.1.1.14"
	// "cpmCPUMemoryLowest"		"1.3.6.1.4.1.9.9.109.1.1.1.1.15"
	used, e := self.GetInt64(params, "1.3.6.1.4.1.9.9.109.1.1.1.1.12.1")
	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}

	free, e := self.GetInt64(params, "1.3.6.1.4.1.9.9.109.1.1.1.1.13.1")
	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}

	return commons.Return(map[string]interface{}{"total": used + free,
		"used_per": (float64(used) / float64(used+free)) * 100,
		"used":     used,
		"free":     free})
}

type memPoolCisco struct {
	snmpBase
}

func (self *memPoolCisco) Call(params MContext) commons.Result {
	// http://tools.cisco.com/Support/SNMP/do/BrowseMIB.do?local=en&step=2&mibName=CISCO-MEMORY-POOL-MIB
	// . iso (1) . org (3) . dod (6) . internet (1) . private (4) . enterprises (1) . cisco (9) . ciscoMgmt (9) . ciscoMemoryPoolMIB (48)
	//    |
	//     - -- ciscoMemoryPoolObjects (1)
	//       |
	//        - -- ciscoMemoryPoolTable (1)
	//       |      |
	//       |       - -- ciscoMemoryPoolEntry (1) object Details
	//       |         |
	//       |         | --   ciscoMemoryPoolType (1)
	//       |         |
	//       |         | --   ciscoMemoryPoolName (2)
	//       |         |
	//       |         | --   ciscoMemoryPoolAlternate (3)
	//       |         |
	//       |         | --   ciscoMemoryPoolValid (4)
	//       |         |
	//       |         | --   ciscoMemoryPoolUsed (5)
	//       |         |
	//       |         | --   ciscoMemoryPoolFree (6)
	//       |         |
	//       |         | --   ciscoMemoryPoolLargestFree (7)
	//       |
	//        + -- ciscoMemoryPoolUtilizationTable (2)

	// ciscoMemoryPoolGroup OBJECT-GROUP
	//     OBJECTS {
	//         ciscoMemoryPoolName,
	//         ciscoMemoryPoolAlternate,
	//         ciscoMemoryPoolValid,
	//         ciscoMemoryPoolUsed,
	//         ciscoMemoryPoolFree,
	//         ciscoMemoryPoolLargestFree
	//     }
	//     STATUS        current
	//     DESCRIPTION        "A collection of objects providing memory pool monitoring.
	// "
	//     ::= { ciscoMemoryPoolGroups 1 }
	return commons.ReturnWithNotImplemented()
}

type memWindows struct {
	snmpBase
}

func (self *memWindows) Call(params MContext) commons.Result {
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
	return commons.Return(map[string]interface{}{"total": total, "used_per": used_per * 100, "used": used, "free": free})
}

func init() {
	Methods["default_mem"] = newRouteSpec("get", "mem", "default mem", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &memWindows{}
			return drv, drv.Init(params)
		})
	Methods["cisco_mem"] = newRouteSpec("get", "mem", "the mem of cisco", Match().Oid("1.3.6.1.4.1.9").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &memCisco{}
			return drv, drv.Init(params)
		})

	Methods["h3c_mem"] = newRouteSpec("get", "mem", "the generic mem of h3c", Match().Oid("1.3.6.1.4.1.25506").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &memH3C{}
			return drv, drv.Init(params)
		})
}

type memH3C struct {
	baseH3C
}

func (self *memH3C) Call(params MContext) commons.Result {
	return self.smartRead("memory", params, self.readMemoryWithNewStyle, self.readMemoryWithCompatibleStyle)
}
