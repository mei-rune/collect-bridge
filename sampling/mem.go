package sampling

import (
	"commons"
)

type memCiscoCpmStyle struct {
	baseCisco
}

func (self *memCiscoCpmStyle) Call(params MContext) commons.Result {
	return self.readMemWithCpmSytle(params)
}

type memCiscoOldStyle struct {
	baseCisco
}

func (self *memCiscoOldStyle) Call(params MContext) commons.Result {
	return self.readMemWithOldSytle(params)
}

type memCiscoMempool struct {
	baseCisco
}

func (self *memCiscoMempool) Call(params MContext) commons.Result {
	return self.readMemWithPoolStyle(params)
}

type memHostResources struct {
	snmpBase
}

func (self *memHostResources) Call(params MContext) commons.Result {
	// http://www.ietf.org/rfc/rfc1514.txt
	// HOST-RESOURCES-MIB:hrStorageTable  = ".1.3.6.1.2.1.25.2.3.1.";
	// HOST-RESOURCES-MIB:hrMemorySize  = ".1.3.6.1.2.1.25.2.2.0";
	// Physical Memory type = "1.3.6.1.2.1.25.2.1.2";
	// HrStorageEntry ::= SEQUENCE {
	//         hrStorageIndex               INTEGER,
	//         hrStorageType                OBJECT IDENTIFIER,
	//         hrStorageDescr               DisplayString,
	//         hrStorageAllocationUnits     INTEGER,
	//         hrStorageSize                INTEGER,
	//         hrStorageUsed                INTEGER,
	//         hrStorageAllocationFailures  Counter
	//     }
	// -- Registration for some storage types, for use with hrStorageType
	//  hrStorageTypes          OBJECT IDENTIFIER ::= { hrStorage 1 }
	//  hrStorageOther          OBJECT IDENTIFIER ::= { hrStorageTypes 1 }
	//  hrStorageRam            OBJECT IDENTIFIER ::= { hrStorageTypes 2 }
	//  -- hrStorageVirtualMemory is temporary storage of swapped
	//  -- or paged memory
	//  hrStorageVirtualMemory  OBJECT IDENTIFIER ::= { hrStorageTypes 3 }
	//  hrStorageFixedDisk      OBJECT IDENTIFIER ::= { hrStorageTypes 4 }
	//  hrStorageRemovableDisk  OBJECT IDENTIFIER ::= { hrStorageTypes 5 }
	//  hrStorageFloppyDisk     OBJECT IDENTIFIER ::= { hrStorageTypes 6 }

	used_per := 0.0
	e := self.OneInTable(params, "1.3.6.1.2.1.25.2.3.1", "2,5,6",
		func(key string, old_row map[string]interface{}) error {
			if "1.3.6.1.2.1.25.2.1.2" != GetOidWithDefault(params, old_row, "2") {
				return commons.ContinueError
			}
			x := GetInt32WithDefault(params, old_row, "6", 0)
			y := GetInt32WithDefault(params, old_row, "5", 0)
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
			drv := &memHostResources{}
			return drv, drv.Init(params)
		})

	Methods["cisco_mem_cpm_style"] = newRouteSpec("get", "cpu", "cisco cpu by cpm style", Match().Oid("1.3.6.1.4.1.9").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &memCiscoCpmStyle{}
			return drv, drv.Init(params)
		})

	Methods["cisco_mem_old_style"] = newRouteSpec("get", "cpu", "cisco cpu by old style", Match().Oid("1.3.6.1.4.1.9").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &memCiscoOldStyle{}
			return drv, drv.Init(params)
		})

	// Methods["cisco_mem_pool_style"] = newRouteSpec("get", "cpu", "cisco cpu by mem pool style", Match().Oid("1.3.6.1.4.1.9").Build(),
	// 	func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
	// 		drv := &memCiscoMempool{}
	// 		return drv, drv.Init(params)
	// 	})

	Methods["h3c_mem"] = newRouteSpec("get", "mem", "the mem of h3c", Match().Oid("1.3.6.1.4.1.25506").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &memH3C{}
			return drv, drv.Init(params)
		})
	Methods["foundry_mem"] = newRouteSpec("get", "mem", "the mem of foundry", Match().Oid("1.3.6.1.4.1.1991").Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &cpuFoundry{}
			return drv, drv.Init(params)
		})
}

type memH3C struct {
	baseH3C
}

func (self *memH3C) Call(params MContext) commons.Result {
	return self.smartRead("memory", params, self.readMemoryWithNewStyle, self.readMemoryWithCompatibleStyle)
}

type memFoundry struct {
	baseFoundry
}

func (self *memFoundry) Call(params MContext) commons.Result {
	return self.readMem(params)
}
