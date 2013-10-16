package sampling

import (
	"commons"
)

type baseCisco struct {
	snmpBase
}

func (self *baseCisco) readSystemUpTime(params MContext) commons.Result {
	// CISCO-SYSTEM-EXT-MIB
	//
	// cseSysUpTime OBJECT-TYPE
	//   SYNTAX          TimeIntervalSec
	//   UNITS           "Seconds"
	//   MAX-ACCESS      read-only
	//   STATUS          current
	//   DESCRIPTION
	//       "The time (in seconds) since the entire system
	//       was last re-initialized as a result of reload. The
	//       initialization being system loaded and running with a system
	//       image software on the first active supervisor (SUP). In High
	//       Availibility (HA) capable system / System that provides
	//       Supervisor module redundancy, this uptime indicates time
	//       elapsed since the first active SUP was booted. This will not
	//       change even if the active SUP goes down and stand-by takes
	//       over as active."
	//   ::= { ciscoSysInfoGroup 10 }
	return self.GetResult(params, "1.3.6.1.4.1.9.9.305.1.1.10.0", RES_UINT64)
}

func (self *baseCisco) readCpuWithOldSytle(params MContext) commons.Result {
	cpu, e := self.GetInt32(params, "1.3.6.1.4.1.9.2.1.57.0")
	if nil == e {
		return commons.Return(map[string]interface{}{"cpu": cpu})
	}
	return self.ErrorResult(e)
}

func (self *baseCisco) readCpuWithSystemExt(params MContext) commons.Result {
	// CISCO-SYSTEM-EXT-MIB
	//
	// cseSysCPUUtilization OBJECT-TYPE
	//   SYNTAX          Gauge32 (0..100 )
	//   UNITS           "%"
	//   MAX-ACCESS      read-only
	//   STATUS          current
	//   DESCRIPTION
	//       "The average utilization of CPU on the active
	//       supervisor."
	//   ::= { ciscoSysInfoGroup 1 }
	cpu, e := self.GetUint32(params, "1.3.6.1.4.1.9.9.305.1.1.1.0")
	if nil == e {
		return commons.Return(map[string]interface{}{"cpu": cpu})
	}
	return self.ErrorResult(e)
}

func GetUint32WithKeys(params MContext, values map[string]interface{}, idx1, idx2 string, defaultValue uint32) uint32 {
	percent := GetUint32WithDefault(params, values, idx1, 0)
	if 0 == percent {
		percent = GetUint32WithDefault(params, values, idx2, 0)
	}
	return percent
}

func GetUint64WithKeys(params MContext, values map[string]interface{}, idx1, idx2 string, defaultValue uint64) uint64 {
	percent := GetUint64WithDefault(params, values, idx1, 0)
	if 0 == percent {
		percent = GetUint64WithDefault(params, values, idx2, 0)
	}
	return percent
}

func (self *baseCisco) readCpuWithCpmSytle(params MContext) commons.Result {
	// CISCO-PROCESS-MIB.my: MIB for CPU and process statistics
	//
	// cpmCPUTotalTable OBJECT-TYPE
	//     SYNTAX          SEQUENCE OF CpmCPUTotalEntry
	//     MAX-ACCESS      not-accessible
	//     STATUS          current
	//     DESCRIPTION
	//         "A table of overall CPU statistics."
	//     ::= { cpmCPU 1 }
	//
	// cpmCPUTotalEntry OBJECT-TYPE
	//     SYNTAX          CpmCPUTotalEntry
	//     MAX-ACCESS      not-accessible
	//     STATUS          current
	//     DESCRIPTION
	//         "Overall information about the CPU load. Entries in this
	//         table come and go as CPUs are added and removed from the
	//         system."
	//     INDEX           { cpmCPUTotalIndex }
	//     ::= { cpmCPUTotalTable 1 }
	//
	// CpmCPUTotalEntry ::= SEQUENCE {
	//         cpmCPUTotalIndex                 Unsigned32,
	//         cpmCPUTotalPhysicalIndex         EntPhysicalIndexOrZero,
	//         cpmCPUTotal5sec                  Gauge32,
	//         cpmCPUTotal1min                  Gauge32,
	//         cpmCPUTotal5min                  Gauge32,
	//         cpmCPUTotal5secRev               Gauge32,
	//         cpmCPUTotal1minRev               Gauge32,
	//         cpmCPUTotal5minRev               Gauge32,
	//         cpmCPUMonInterval                Unsigned32,
	//         cpmCPUTotalMonIntervalValue      Gauge32,
	//         cpmCPUInterruptMonIntervalValue  Gauge32,
	//         cpmCPUMemoryUsed                 Gauge32,
	//         cpmCPUMemoryFree                 Gauge32,
	//         cpmCPUMemoryKernelReserved       Gauge32,
	//         cpmCPUMemoryLowest               Gauge32,
	//         cpmCPUMemoryUsedOvrflw           Gauge32,
	//         cpmCPUMemoryHCUsed               CounterBasedGauge64,
	//         cpmCPUMemoryFreeOvrflw           Gauge32,
	//         cpmCPUMemoryHCFree               Counter64,
	//         cpmCPUMemoryKernelReservedOvrflw Gauge32,
	//         cpmCPUMemoryHCKernelReserved     CounterBasedGauge64,
	//         cpmCPUMemoryLowestOvrflw         Gauge32,
	//         cpmCPUMemoryHCLowest             CounterBasedGauge64,
	//         cpmCPULoadAvg1min                CPULoadAverage,
	//         cpmCPULoadAvg5min                CPULoadAverage,
	//         cpmCPULoadAvg15min               CPULoadAverage,
	//         cpmCPUMemoryCommitted            Gauge32,
	//         cpmCPUMemoryCommittedOvrflw      Gauge32,
	//         cpmCPUMemoryHCCommitted          CounterBasedGauge64
	// }

	cpu_list := make([]map[string]interface{}, 10)
	total := uint32(0)
	e := self.EachInTable(params, "1.3.6.1.4.1.9.9.109.1.1.1.1", "1,3,4,5,6,7,8",
		func(key string, old_row map[string]interface{}) error {
			percent := GetUint32WithKeys(params, old_row, "7", "4", 0)
			total += percent
			//         cpmCPUTotal5sec                  Gauge32, 3
			//         cpmCPUTotal1min                  Gauge32, 4
			//         cpmCPUTotal5min                  Gauge32, 5

			//         cpmCPUTotal5secRev               Gauge32, 6
			//         cpmCPUTotal1minRev               Gauge32, 7
			//         cpmCPUTotal5minRev               Gauge32, 8
			cpu_list = append(cpu_list, map[string]interface{}{
				"cpmCPUTotalIndex": GetUint32WithDefault(params, old_row, "1", 0),
				"cpmCPUTotal5sec":  GetUint32WithKeys(params, old_row, "6", "3", 0),
				"cpmCPUTotal1min":  percent,
				"cpmCPUTotal5min":  GetUint32WithKeys(params, old_row, "8", "5", 0),
			})
			return nil
		})

	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}

	if 1 == len(cpu_list) {
		cpu_list[0]["cpu"] = total
		return commons.Return(cpu_list[0])
	}

	return commons.Return(map[string]interface{}{"cpu": total / uint32(len(cpu_list)), "cpu_list": cpu_list})
}

func (self *baseCisco) readMemWithOldSytle(params MContext) commons.Result {
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

func (self *baseCisco) readMemWithCpmSytle(params MContext) commons.Result {
	// see readCpuWithCpmSytle
	// "cpmCPUMemoryUsed"	             	"1.3.6.1.4.1.9.9.109.1.1.1.1.12"
	// "cpmCPUMemoryFree"		            "1.3.6.1.4.1.9.9.109.1.1.1.1.13"
	// "cpmCPUMemoryKernelReserved"		  "1.3.6.1.4.1.9.9.109.1.1.1.1.14"
	// cpmCPUMemoryHCUsed               "1.3.6.1.4.1.9.9.109.1.1.1.1.17"
	// cpmCPUMemoryHCFree               "1.3.6.1.4.1.9.9.109.1.1.1.1.19"
	// cpmCPUMemoryHCKernelReserved     "1.3.6.1.4.1.9.9.109.1.1.1.1.21"

	mem_pool := make([]map[string]interface{}, 10)
	total_reserved := uint64(0)
	total_used := uint64(0)
	total_free := uint64(0)

	e := self.EachInTable(params, "1.3.6.1.4.1.9.9.109.1.1.1.1", "1,12,13,14,17,19,21",
		func(key string, old_row map[string]interface{}) error {
			used := GetUint64WithKeys(params, old_row, "17", "12", 0)
			free := GetUint64WithKeys(params, old_row, "19", "13", 0)
			reserved := GetUint64WithKeys(params, old_row, "21", "14", 0)

			total_reserved += reserved
			total_used += used
			total_free += free

			mem_pool = append(mem_pool, map[string]interface{}{
				"cpmCPUTotalIndex": GetUint32WithDefault(params, old_row, "1", 0),
				"total":            used + free + reserved,
				"used_per":         (float64(used) / (float64(used) + float64(free) + float64(reserved))) * 100,
				"used":             used + reserved,
				"free":             free,
				"reserved":         reserved})
			return nil
		})

	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}

	if 1 == len(mem_pool) {
		return commons.Return(mem_pool[0])
	}

	return commons.Return(map[string]interface{}{"total": total_reserved + total_used + total_free,
		"used_per": (float64(total_used) / (float64(total_used) + float64(total_free) + float64(total_reserved))) * 100,
		"used":     total_used + total_reserved,
		"free":     total_free,
		"reserved": total_reserved,
		"mem_list": mem_pool})
}

func (self *baseCisco) readMem(params MContext) commons.Result {
	// CISCO-SYSTEM-EXT-MIB
	//
	// cseSysMemoryUtilization OBJECT-TYPE
	//   SYNTAX          Gauge32 (0..100 )
	//   UNITS           "%"
	//   MAX-ACCESS      read-only
	//   STATUS          current
	//   DESCRIPTION
	//       "The average utilization of memory on the active
	//       supervisor."
	//   ::= { ciscoSysInfoGroup 2 }
	return self.GetResult(params, "1.3.6.1.4.1.9.9.305.1.1.2", RES_UINT32)
}

func (self *baseCisco) readMemWithPoolStyle(params MContext) commons.Result {
	// http://tools.cisco.com/Support/SNMP/do/BrowseMIB.do?local=en&step=2&mibName=CISCO-MEMORY-POOL-MIB
	// . iso (1) . org (3) . dod (6) . internet (1) . private (4) . enterprises (1) . cisco (9) . ciscoMgmt (9) . ciscoMemoryPoolMIB (48)
	//    |
	//    +--+- ciscoMemoryPoolObjects (1)
	//       |
	//       +---- ciscoMemoryPoolTable (1)
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
	//       + -- ciscoMemoryPoolUtilizationTable (2)

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
