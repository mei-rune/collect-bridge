package sampling

import (
	"commons"
	"errors"
	"strings"
)

// http://www.h3c.com.cn/products___technology/technology/system_management/other_technology/representative_collocate_enchiridion/200906/636062_30003_0.htm

// .1.3.6.1.2.1.47.1.1.1.1.2
// entPhysicalDescr OBJECT-TYPE
//   -- FROM  ENTITY-MIB
//   -- TEXTUAL CONVENTION SnmpAdminString
//   SYNTAX OCTET STRING (0..255)
//   DISPLAY-HINT "255a"
//   MAX-ACCESS read-only
//   STATUS current
//   DESCRIPTION  "A textual description of physical entity.  This object
//             should contain a string that identifies the manufacturer's
//             name for the physical entity, and should be set to a
//             distinct value for each version or model of the physical
//             entity."
// ::= { iso(1) org(3) dod(6) internet(1) mgmt(2) mib-2(1) entityMIB(47) entityMIBObjects(1) entityPhysical(1) entPhysicalTable(1) entPhysicalEntry(1) 2 }

// .1.3.6.1.2.1.47.1.1.1.1.5
// entPhysicalClass OBJECT-TYPE
//   -- FROM  ENTITY-MIB
//   -- TEXTUAL CONVENTION PhysicalClass
//   SYNTAX INTEGER {other(1), unknown(2), chassis(3), backplane(4), container(5), powerSupply(6), fan(7), sensor(8), module(9), port(10), stack(11), cpu(12)}
//   MAX-ACCESS read-only
//   STATUS current
//   DESCRIPTION  "An indication of the general hardware type of the physical
//             entity.

//             An agent should set this object to the standard enumeration
//             value that most accurately indicates the general class of
//             the physical entity, or the primary class if there is more
//             than one entity.

//             If no appropriate standard registration identifier exists
//             for this physical entity, then the value 'other(1)' is
//             returned.  If the value is unknown by this agent, then the
//             value 'unknown(2)' is returned."
// ::= { iso(1) org(3) dod(6) internet(1) mgmt(2) mib-2(1) entityMIB(47) entityMIBObjects(1) entityPhysical(1) entPhysicalTable(1) entPhysicalEntry(1) 5 }

// .1.3.6.1.2.1.47.1.1.1.1.7
// entPhysicalName OBJECT-TYPE
//   -- FROM  ENTITY-MIB
//   -- TEXTUAL CONVENTION SnmpAdminString
//   SYNTAX OCTET STRING (0..255)
//   DISPLAY-HINT "255a"
//   MAX-ACCESS read-only
//   STATUS current
//   DESCRIPTION  "The textual name of the physical entity.  The value of this
//             object should be the name of the component as assigned by
//             the local device and should be suitable for use in commands
//             entered at the device's `console'.  This might be a text
//             name (e.g., `console') or a simple component number (e.g.,
//             port or module number, such as `1'), depending on the
//             physical component naming syntax of the device.

//             If there is no local name, or if this object is otherwise
//             not applicable, then this object contains a zero-length
//             string.

//             Note that the value of entPhysicalName for two physical
//             entities will be the same in the event that the console
//             interface does not distinguish between them, e.g., slot-1
//             and the card in slot-1."
// ::= { iso(1) org(3) dod(6) internet(1) mgmt(2) mib-2(1) entityMIB(47) entityMIBObjects(1) entityPhysical(1)

// .1.3.6.1.2.1.47.1.1.1.1.13
// entPhysicalModelName OBJECT-TYPE
//   -- FROM  ENTITY-MIB
//   -- TEXTUAL CONVENTION SnmpAdminString
//   SYNTAX OCTET STRING (0..255)
//   DISPLAY-HINT "255a"
//   MAX-ACCESS read-only
//   STATUS current
//   DESCRIPTION  "The vendor-specific model name identifier string associated
//             with this physical component.  The preferred value is the
//             customer-visible part number, which may be printed on the
//             component itself.

//             If the model name string associated with the physical
//             component is unknown to the agent, then this object will
//             contain a zero-length string."
// ::= { iso(1) org(3) dod(6) internet(1) mgmt(2) mib-2(1) entityMIB(47) entityMIBObjects(1) entityPhysical(1) entPhysicalTable(1) entPhysicalEntry(1) 13 }

// compatible style
// .1.3.6.1.4.1.2011.10.2.6.1.1.1.1.6
// h3cEntityExtCpuUsage OBJECT-TYPE
//   -- FROM  H3C-ENTITY-EXT-MIB
//   SYNTAX Integer32 (0..100)
//   MAX-ACCESS read-only
//   STATUS current
//   DESCRIPTION  "The CPU usage for this entity. Generally, the CPU usage
//             will calculate the overall CPU usage on the entity, and it
//             is not sensible with the number of CPU on the entity."
// ::= { iso(1) org(3) dod(6) internet(1) private(4) enterprises(1) huawei(2011) h3c(10) h3cCommon(2) h3cEntityExtend(6) h3cEntityExtObjects(1) h3cEntityExtState(1) h3cEntityExtStateTable(1) h3cEntityExtStateEntry(1) 6 }

// .1.3.6.1.4.1.2011.10.2.6.1.1.1.1.7
// h3cEntityExtCpuUsageThreshold OBJECT-TYPE
//   -- FROM  H3C-ENTITY-EXT-MIB
//   SYNTAX Integer32 (0..100)
//   MAX-ACCESS read-write
//   STATUS current
//   DESCRIPTION  "The threshold for the CPU usage. When the CPU usage exceeds
//             the threshold, a notification will be sent."
// ::= { iso(1) org(3) dod(6) internet(1) private(4) enterprises(1) huawei(2011) h3c(10) h3cCommon(2) h3cEntityExtend(6) h3cEntityExtObjects(1) h3cEntityExtState(1) h3cEntityExtStateTable(1) h3cEntityExtStateEntry(1) 7 }

// .1.3.6.1.4.1.2011.10.2.6.1.1.1.1.8
// h3cEntityExtMemUsage OBJECT-TYPE
//   -- FROM  H3C-ENTITY-EXT-MIB
//   SYNTAX Integer32 (0..100)
//   MAX-ACCESS read-only
//   STATUS current
//   DESCRIPTION  "The memory usage for the entity. This object indicates what
//             percent of memory are used."
// ::= { iso(1) org(3) dod(6) internet(1) private(4) enterprises(1) huawei(2011) h3c(10) h3cCommon(2) h3cEntityExtend(6) h3cEntityExtObjects(1) h3cEntityExtState(1) h3cEntityExtStateTable(1) h3cEntityExtStateEntry(1) 8 }

// .1.3.6.1.4.1.2011.10.2.6.1.1.1.1.9
// h3cEntityExtMemUsageThreshold OBJECT-TYPE
//   -- FROM  H3C-ENTITY-EXT-MIB
//   SYNTAX Integer32 (0..100)
//   MAX-ACCESS read-write
//   STATUS current
//   DESCRIPTION  "The threshold for the Memory usage, When the memory usage
//             exceeds the threshold, a notification will be sent."
// ::= { iso(1) org(3) dod(6) internet(1) private(4) enterprises(1) huawei(2011) h3c(10) h3cCommon(2) h3cEntityExtend(6) h3cEntityExtObjects(1) h3cEntityExtState(1) h3cEntityExtStateTable(1) h3cEntityExtStateEntry(1) 9 }

// .1.3.6.1.4.1.2011.10.2.6.1.1.1.1.10
// h3cEntityExtMemSize OBJECT-TYPE
//   -- FROM  H3C-ENTITY-EXT-MIB
//   SYNTAX Unsigned32
//   UNITS  ""
//   MAX-ACCESS read-only
//   STATUS current
//   DESCRIPTION  "The size of memory for the entity."
// ::= { iso(1) org(3) dod(6) internet(1) private(4) enterprises(1) huawei(2011) h3c(10) h3cCommon(2) h3cEntityExtend(6) h3cEntityExtObjects(1) h3cEntityExtState(1) h3cEntityExtStateTable(1) h3cEntityExtStateEntry(1) 10 }

// .1.3.6.1.4.1.2011.10.2.6.1.1.1.1.11
// h3cEntityExtUpTime OBJECT-TYPE
//   -- FROM  H3C-ENTITY-EXT-MIB
//   SYNTAX Integer32
//   UNITS  ""
//   MAX-ACCESS read-only
//   STATUS current
//   DESCRIPTION  "The uptime for the entity. The meaning of uptime is
//             when the entity is up, and the value of the object
//             will add 1 seconds while the entity is running."
// ::= { iso(1) org(3) dod(6) internet(1) private(4) enterprises(1) huawei(2011) h3c(10) h3cCommon(2) h3cEntityExtend(6) h3cEntityExtObjects(1) h3cEntityExtState(1) h3cEntityExtStateTable(1) h3cEntityExtStateEntry(1) 11 }

// new style
// .1.3.6.1.4.1.25506.2.6.1.1.1.1.6
// hh3cEntityExtCpuUsage OBJECT-TYPE
//   -- FROM  HH3C-ENTITY-EXT-MIB
//   SYNTAX Integer32 (0..100)
//   MAX-ACCESS read-only
//   STATUS current
//   DESCRIPTION  "The CPU usage for this entity. Generally, the CPU usage
//             will calculate the overall CPU usage on the entity, and it
//             is not sensible with the number of CPU on the entity."
// ::= { iso(1) org(3) dod(6) internet(1) private(4) enterprises(1) hh3c(25506) hh3cCommon(2) hh3cEntityExtend(6) hh3cEntityExtObjects(1) hh3cEntityExtState(1) hh3cEntityExtStateTable(1) hh3cEntityExtStateEntry(1) 6 }

// .1.3.6.1.4.1.25506.2.6.1.1.1.1.7
// hh3cEntityExtCpuUsageThreshold OBJECT-TYPE
//   -- FROM  HH3C-ENTITY-EXT-MIB
//   SYNTAX Integer32 (0..100)
//   MAX-ACCESS read-write
//   STATUS current
//   DESCRIPTION  "The threshold for the CPU usage. When the CPU usage exceeds
//             the threshold, a notification will be sent."
// ::= { iso(1) org(3) dod(6) internet(1) private(4) enterprises(1) hh3c(25506) hh3cCommon(2) hh3cEntityExtend(6) hh3cEntityExtObjects(1) hh3cEntityExtState(1) hh3cEntityExtStateTable(1) hh3cEntityExtStateEntry(1) 7 }

// .1.3.6.1.4.1.25506.2.6.1.1.1.1.8
// hh3cEntityExtMemUsage OBJECT-TYPE
//   -- FROM  HH3C-ENTITY-EXT-MIB
//   SYNTAX Integer32 (0..100)
//   MAX-ACCESS read-only
//   STATUS current
//   DESCRIPTION  "The memory usage for the entity. This object indicates what
//             percent of memory are used."
// ::= { iso(1) org(3) dod(6) internet(1) private(4) enterprises(1) hh3c(25506) hh3cCommon(2) hh3cEntityExtend(6) hh3cEntityExtObjects(1) hh3cEntityExtState(1) hh3cEntityExtStateTable(1) hh3cEntityExtStateEntry(1) 8 }

// .1.3.6.1.4.1.25506.2.6.1.1.1.1.9
// hh3cEntityExtMemUsageThreshold OBJECT-TYPE
//   -- FROM  HH3C-ENTITY-EXT-MIB
//   SYNTAX Integer32 (0..100)
//   MAX-ACCESS read-write
//   STATUS current
//   DESCRIPTION  "The threshold for the Memory usage, When the memory usage
//             exceeds the threshold, a notification will be sent."
// ::= { iso(1) org(3) dod(6) internet(1) private(4) enterprises(1) hh3c(25506) hh3cCommon(2) hh3cEntityExtend(6) hh3cEntityExtObjects(1) hh3cEntityExtState(1) hh3cEntityExtStateTable(1) hh3cEntityExtStateEntry(1) 9 }

// .1.3.6.1.4.1.25506.2.6.1.1.1.1.10
// hh3cEntityExtMemSize OBJECT-TYPE
//   -- FROM  HH3C-ENTITY-EXT-MIB
//   SYNTAX Unsigned32
//   UNITS  ""
//   MAX-ACCESS read-only
//   STATUS current
//   DESCRIPTION  "The size of memory for the entity."
// ::= { iso(1) org(3) dod(6) internet(1) private(4) enterprises(1) hh3c(25506) hh3cCommon(2) hh3cEntityExtend(6) hh3cEntityExtObjects(1) hh3cEntityExtState(1) hh3cEntityExtStateTable(1) hh3cEntityExtStateEntry(1) 10 }

// .1.3.6.1.4.1.2011.10.2.6.1.1.1.1.23
// h3cEntityExtPhyMemSize OBJECT-TYPE
//   -- FROM  H3C-ENTITY-EXT-MIB
//   SYNTAX Unsigned32
//   MAX-ACCESS read-only
//   STATUS current
//   DESCRIPTION  "The memory size of entity. This is the physical attribute of entity."
// ::= { iso(1) org(3) dod(6) internet(1) private(4) enterprises(1) huawei(2011) h3c(10) h3cCommon(2) h3cEntityExtend(6) h3cEntityExtObjects(1) h3cEntityExtState(1) h3cEntityExtStateTable(1) h3cEntityExtStateEntry(1) 23 }

// .1.3.6.1.4.1.2011.10.2.6.1.1.1.1.24
// h3cEntityExtPhyCpuFrequency OBJECT-TYPE
//   -- FROM  H3C-ENTITY-EXT-MIB
//   SYNTAX Integer32
//   MAX-ACCESS read-only
//   STATUS current
//   DESCRIPTION  "The CPU frequency of entity. Unit of measure is MHZ."
// ::= { iso(1) org(3) dod(6) internet(1) private(4) enterprises(1) huawei(2011) h3c(10) h3cCommon(2) h3cEntityExtend(6) h3cEntityExtObjects(1) h3cEntityExtState(1) h3cEntityExtStateTable(1) h3cEntityExtStateEntry(1) 24 }

// .1.3.6.1.4.1.2011.10.2.6.1.1.1.1.26
// h3cEntityExtCpuAvgUsage OBJECT-TYPE
//   -- FROM  H3C-ENTITY-EXT-MIB
//   SYNTAX Integer32 (0..100)
//   MAX-ACCESS read-only
//   STATUS current
//   DESCRIPTION  "The average CPU usage for the entity in a period of time."
// ::= { iso(1) org(3) dod(6) internet(1) private(4) enterprises(1) huawei(2011) h3c(10) h3cCommon(2) h3cEntityExtend(6) h3cEntityExtObjects(1) h3cEntityExtState(1) h3cEntityExtStateTable(1) h3cEntityExtStateEntry(1) 26 }

// .1.3.6.1.4.1.2011.10.2.6.1.1.1.1.27
// h3cEntityExtMemAvgUsage OBJECT-TYPE
//   -- FROM  H3C-ENTITY-EXT-MIB
//   SYNTAX Integer32 (0..100)
//   MAX-ACCESS read-only
//   STATUS current
//   DESCRIPTION  "The average memory usage for the entity in a period of time."
// ::= { iso(1) org(3) dod(6) internet(1) private(4) enterprises(1) huawei(2011) h3c(10) h3cCommon(2) h3cEntityExtend(6) h3cEntityExtObjects(1) h3cEntityExtState(1) h3cEntityExtStateTable(1) h3cEntityExtStateEntry(1) 27 }

// .1.3.6.1.4.1.2011.10.2.6.1.1.1.1.28
// h3cEntityExtMemType OBJECT-TYPE
//   -- FROM  H3C-ENTITY-EXT-MIB
//   SYNTAX OCTET STRING (0..64)
//   MAX-ACCESS read-only
//   STATUS current
//   DESCRIPTION  "The memory type of entity."
// ::= { iso(1) org(3) dod(6) internet(1) private(4) enterprises(1) huawei(2011) h3c(10) h3cCommon(2) h3cEntityExtend(6) h3cEntityExtObjects(1) h3cEntityExtState(1) h3cEntityExtStateTable(1) h3cEntityExtStateEntry(1) 28 }

type baseH3C struct {
	snmpBase
}

func (self *baseH3C) readMemoryWithStyle(size_oid, usage_oid string, params MContext, id_list map[string]map[string]interface{}) (map[string]interface{}, error) {

	var total int
	var used int
	var used_per float64
	var e error

	var results []map[string]interface{}
	for id, res := range id_list {
		var memSize int32
		var memUsage int32

		memSize, e = self.GetInt32(params, size_oid+id)
		if nil != e {
			return nil, e
		}

		if 0 == memSize {
			continue
		}

		total += int(memSize)
		memUsage, e = self.GetInt32(params, usage_oid+id)
		if nil != e {
			return nil, e
		}

		used_per = float64(memUsage)
		used += (int(memSize*memUsage) / 100)

		res["used_per"] = used_per
		res["totel"] = memSize
		res["used"] = used
		res["free"] = int(memSize) - used
		if nil == results {
			results = make([]map[string]interface{}, 0, len(id_list))
		}
		results = append(results, res)
	}

	switch len(results) {
	case 0:
		return nil, errors.New("the results of memory is empty.")
	case 1:
		return map[string]interface{}{"total": total,
			"used_per": used_per,
			"used":     used,
			"free":     total - used}, nil
	default:
		used_per = float64(used*100) / float64(total)
		return map[string]interface{}{"total": total,
			"used_per": used_per,
			"used":     used,
			"free":     total - used, "details": results}, nil
	}
}

func (self *baseH3C) readMemoryWithCompatibleStyle(params MContext,
	id_list map[string]map[string]interface{}) (map[string]interface{}, error) {
	return self.readMemoryWithStyle("1.3.6.1.4.1.2011.10.2.6.1.1.1.1.10.",
		"1.3.6.1.4.1.2011.10.2.6.1.1.1.1.8.", params, id_list)
}

func (self *baseH3C) readMemoryWithNewStyle(params MContext,
	id_list map[string]map[string]interface{}) (map[string]interface{}, error) {
	return self.readMemoryWithStyle("1.3.6.1.4.1.25506.2.6.1.1.1.1.10.",
		"1.3.6.1.4.1.25506.2.6.1.1.1.1.8.", params, id_list)
}

func (self *baseH3C) readCpuWithStyle(usage_oid string, params MContext, id_list map[string]map[string]interface{}) (map[string]interface{}, error) {

	var total int
	var results []int
	var detail_results []map[string]interface{}
	for id, res := range id_list {
		cpuUsage, e := self.GetInt32(params, usage_oid+id)
		if nil != e {
			return nil, e
		}
		if 0 == cpuUsage {
			continue
		}

		total += int(cpuUsage)
		if nil == results {
			results = make([]int, 0, len(id_list))
		}
		results = append(results, int(cpuUsage))

		if nil == detail_results {
			detail_results = make([]map[string]interface{}, 0, len(id_list))
		}
		res["cpuUsage"] = cpuUsage
		detail_results = append(detail_results, res)
	}

	switch len(results) {
	case 0:
		return nil, errors.New("the results of cpu is empty.")
	case 1:
		return map[string]interface{}{"cpu": total, "cpu_list": results}, nil
	default:
		return map[string]interface{}{"cpu": total / len(results),
			"cpu_list":    results,
			"cpu_details": detail_results}, nil
	}
}

func (self *baseH3C) readCpuWithCompatibleStyle(params MContext,
	id_list map[string]map[string]interface{}) (map[string]interface{}, error) {
	return self.readCpuWithStyle("1.3.6.1.4.1.2011.10.2.6.1.1.1.1.6.", params, id_list)
}

func (self *baseH3C) readCpuWithNewStyle(params MContext,
	id_list map[string]map[string]interface{}) (map[string]interface{}, error) {
	return self.readCpuWithStyle("1.3.6.1.4.1.25506.2.6.1.1.1.1.6.", params, id_list)
}

type readWithStyle func(params MContext, id_list map[string]map[string]interface{}) (map[string]interface{}, error)

func (self *baseH3C) smartRead(name string, params MContext, new_style, compatible_style readWithStyle) commons.Result {
	id_list := make(map[string]map[string]interface{}, 10)
	e := self.EachInTable(params, "1.3.6.1.2.1.47.1.1.1.1", "2,5,7,13",
		func(key string, old_row map[string]interface{}) error {
			//fmt.Println("oid=", key)
			clazz := GetInt32(params, old_row, "5", 0)
			if 1 != clazz && 5 != clazz && 9 != clazz {
				return nil
			}

			slot := key
			idx := strings.LastIndex(key, ".")
			if -1 != idx {
				slot = key[idx+1:]
			}

			id_list[slot] = map[string]interface{}{
				"entPhysicalDescr":     GetString(params, old_row, "2"),
				"entPhysicalClass":     GetInt32(params, old_row, "5", 0),
				"entPhysicalName":      GetString(params, old_row, "7"),
				"entPhysicalModelName": GetString(params, old_row, "13"),
			}
			return nil
		})

	if nil != e {
		return commons.ReturnWithInternalError("read slots of " + name + " failed, " + e.Error())
	}

	if 0 == len(id_list) {
		return commons.ReturnWithNotAcceptable("not support device.")
	}

	h3c_style := params.GetIntWithDefault("@h2c_style", 0)
	old_h3c_style := h3c_style

	var res map[string]interface{}

restart_read:
	switch h3c_style {
	case 0:
		fallthrough
	case 2011:
		res, e = compatible_style(params, id_list)
		if nil == e {
			h3c_style = 2011
			break
		}
		if nil != e {
			return commons.ReturnWithInternalError("read " + name + " with compatible style failed, " + e.Error())
		}
		fallthrough
	case 25506:
		res, e = new_style(params, id_list)
		if nil != e {
			return commons.ReturnWithInternalError("read " + name + " with new style failed, " + e.Error())
		}

		h3c_style = 25506
	default:
		h3c_style = 0
		goto restart_read
	}

	if nil == res {
		return commons.ReturnWithInternalError("the results of " + name + " is empty.")
	}

	if h3c_style != old_h3c_style {
		params.Set("@h2c_style", h3c_style)
	}

	return commons.Return(res)
}
