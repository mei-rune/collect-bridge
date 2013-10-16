package sampling

import (
	"commons"
)

// snAgentSys                 1.3.6.1.4.1.1991.1.1.2
// snAgentCpu                 1.3.6.1.4.1.1991.1.1.2.11
// snAgentCpuUtilTable        1.3.6.1.4.1.1991.1.1.2.11.1
// snAgentCpuUtilEntry        1.3.6.1.4.1.1991.1.1.2.11.1.1
// snAgentCpuUtilSlotNum      1.3.6.1.4.1.1991.1.1.2.11.1.1.1
// snAgentCpuUtilCpuId        1.3.6.1.4.1.1991.1.1.2.11.1.1.2
// snAgentCpuUtilInterval     1.3.6.1.4.1.1991.1.1.2.11.1.1.3
// snAgentCpuUtilValue        1.3.6.1.4.1.1991.1.1.2.11.1.1.4
// snAgentCpuUtilPercent      1.3.6.1.4.1.1991.1.1.2.11.1.1.5
// snAgentCpuUtil100thPercent 1.3.6.1.4.1.1991.1.1.2.11.1.1.6
//
//

// snAgentCpuUtilEntry OBJECT-TYPE
//   SYNTAX  SnAgentCpuUtilEntry
//   MAX-ACCESS  not-accessible
//   STATUS  current
//   DESCRIPTION
//     "A row in the CPU utilization table."
//   INDEX {
//     snAgentCpuUtilSlotNum,
//     snAgentCpuUtilCpuId,
//     snAgentCpuUtilInterval
//   }
//   ::= { snAgentCpuUtilTable 1 }
//
// SnAgentCpuUtilEntry ::= SEQUENCE {
//   snAgentCpuUtilSlotNum
//     Integer32,
//   snAgentCpuUtilCpuId
//     Integer32,
//   snAgentCpuUtilInterval
//     Integer32,
//   snAgentCpuUtilValue
//     Gauge32,
//   snAgentCpuUtilPercent
//     Gauge32,
//   snAgentCpuUtil100thPercent
//     Gauge32
//   }
//
// snAgentCpuUtilSlotNum OBJECT-TYPE
//   SYNTAX Integer32
//   MAX-ACCESS  read-only
//   STATUS  current
//   DESCRIPTION
//     "The slot number of module which contains the cpu."
//   ::= { snAgentCpuUtilEntry 1 }
//
// snAgentCpuUtilCpuId  OBJECT-TYPE
//   SYNTAX   Integer32
//   MAX-ACCESS  read-only
//   STATUS  current
//   DESCRIPTION
//     "The id of cpu. For non-VM1/WSM management module, there is one CPU.
//     For VM1/WSM there's one management CPU and three slave CPUs.
//     The management CPU could be turned off. For POS and ATM
//     there's no management CPU but  two slave CPUs.
//     Id for management cpu is 1. Value of 2 or greater are for slave CPUs. "
//   ::= { snAgentCpuUtilEntry 2 }
//
// snAgentCpuUtilInterval  OBJECT-TYPE
//   SYNTAX   Integer32
//   MAX-ACCESS  read-only
//   STATUS  current
//   DESCRIPTION
//     "The value, in seconds, for this utilization. For both management and slave CPU, we display
//     utilization for 1 sec, 5 sec, 60 sec and 300 sec interval."
//   ::= { snAgentCpuUtilEntry 3 }
//
// snAgentCpuUtilValue OBJECT-TYPE
//   SYNTAX  Gauge32
//   MAX-ACCESS  read-only
//   STATUS  deprecated
//   DESCRIPTION
//     "The statistical CPU utilization in units of one-hundredth
//      of a percent. This value is deprecated. Users are recommended
//      to use snAgentCpuUtilPercent or snAgentCpuUtil100thPercent
//      instead."
//   ::= { snAgentCpuUtilEntry 4 }
//
// snAgentCpuUtilPercent OBJECT-TYPE
//   SYNTAX  Gauge32
//   MAX-ACCESS  read-only
//   STATUS  current
//   DESCRIPTION
//     "The statistical CPU utilization in units of a percent."
//   ::= { snAgentCpuUtilEntry 5 }
//
// snAgentCpuUtil100thPercent OBJECT-TYPE
//   SYNTAX  Gauge32
//   MAX-ACCESS  read-only
//   STATUS  current
//   DESCRIPTION
//     "The statistical CPU utilization in units of one-hundredth
//      of a percent."
//   ::= { snAgentCpuUtilEntry 6 }

type baseFoundry struct {
	snmpBase
}

func (self *baseFoundry) readCpu(params MContext) commons.Result {
	cpu_list := make([]map[string]interface{}, 10)
	total := uint32(0)
	e := self.EachInTable(params, "1.3.6.1.4.1.1991.1.1.2.11.1.1", "1,2,3,4,5,6",
		func(key string, old_row map[string]interface{}) error {
			percent := GetUint32WithDefault(params, old_row, "5", 0)
			total += percent
			cpu_list = append(cpu_list, map[string]interface{}{
				"snAgentCpuUtilSlotNum":      GetInt32WithDefault(params, old_row, "1", 0),
				"snAgentCpuUtilCpuId":        GetInt32WithDefault(params, old_row, "2", 0),
				"snAgentCpuUtilInterval":     GetInt32WithDefault(params, old_row, "3", 0),
				"snAgentCpuUtilValue":        GetUint32WithDefault(params, old_row, "4", 0),
				"snAgentCpuUtilPercent":      percent,
				"snAgentCpuUtil100thPercent": GetUint32WithDefault(params, old_row, "6", 0),
			})
			return nil
		})

	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}
	return commons.Return(map[string]interface{}{"cpu": total / uint32(len(cpu_list)), "cpu_list": cpu_list})
}

// -- System DRAM info Group.
// snAgentSys                 1.3.6.1.4.1.1991.1.1.2
// snAgentHw                  1.3.6.1.4.1.1991.1.1.2.12
// snAgSystemDRAM             1.3.6.1.4.1.1991.1.1.2.12.4
// snAgSystemDRAM       OBJECT IDENTIFIER ::= { snAgentHw 4 }
//
// snAgSystemDRAMUtil OBJECT-TYPE
//   SYNTAX  Gauge32
//   MAX-ACCESS  read-only
//   STATUS  current
//   DESCRIPTION
//     "The system dynamic memory utilization, in unit of percentage."
//   ::= { snAgSystemDRAM 1 }
//
// snAgSystemDRAMTotal OBJECT-TYPE
//   SYNTAX   Integer32
//   MAX-ACCESS  read-only
//   STATUS  current
//   DESCRIPTION
//     "The total amount of system dynamic memory, in number of bytes."
//   ::= { snAgSystemDRAM 2 }
//
// snAgSystemDRAMFree OBJECT-TYPE
//   SYNTAX   Integer32
//   MAX-ACCESS  read-only
//   STATUS  current
//   DESCRIPTION
//     "The free amount of system dynamic memory, in number of bytes."
//   ::= { snAgSystemDRAM 3 }
//
// snAgSystemDRAMForBGP OBJECT-TYPE
//   SYNTAX   Integer32
//   MAX-ACCESS  read-only
//   STATUS  current
//   DESCRIPTION
//     "The free amount of system dynamic memory used by BGP, in number of bytes."
//   ::= { snAgSystemDRAM 4 }
//
// snAgSystemDRAMForOSPF OBJECT-TYPE
//   SYNTAX   Integer32
//   MAX-ACCESS  read-only
//   STATUS  current
//   DESCRIPTION
//     "The free amount of system dynamic memory used by OSPF, in number of bytes."
//   ::= { snAgSystemDRAM 5 }

func (self *baseFoundry) readMem(params MContext) commons.Result {
	snAgSystemDRAMUtil, e := self.GetUint32(params, "1.3.6.1.4.1.1991.1.1.2.12.4.1.0")
	if nil != e {
		return self.ErrorResult(e)
	}
	snAgSystemDRAMTotal, e := self.GetUint32(params, "1.3.6.1.4.1.1991.1.1.2.12.4.2.0")
	if nil != e {
		return self.ErrorResult(e)
	}
	snAgSystemDRAMFree, e := self.GetUint32(params, "1.3.6.1.4.1.1991.1.1.2.12.4.3.0")
	if nil != e {
		return self.ErrorResult(e)
	}

	return commons.Return(map[string]interface{}{"total": snAgSystemDRAMTotal,
		"used_per": snAgSystemDRAMUtil,
		"used":     snAgSystemDRAMTotal - snAgSystemDRAMFree,
		"free":     snAgSystemDRAMFree})
}
