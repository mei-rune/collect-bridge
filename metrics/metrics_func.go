package metrics

import (
	"bytes"
	"code.google.com/p/mahonia"
	"commons"
	"commons/as"
	"encoding/hex"
	"fmt"
	"snmp"
	"strings"
)

type Base struct {
	drvMgr *commons.DriverManager
	drv    commons.Driver
}

type RECORD map[string]interface{}

func (self *Base) Init(drvMgr *commons.DriverManager, drvName string) commons.RuntimeError {
	self.drvMgr = drvMgr
	self.drv, _ = drvMgr.Connect(drvName)
	if nil == self.drv {
		return commons.NotFound(drvName)
	}
	return nil
}

func GetIntColumn(params map[string]string, old_row RECORD, idx string, defaultValue int) int {
	value, ok := old_row[idx]
	if !ok {
		panic("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmp.SnmpValue:
		i, e := snmp.AsInt(v)
		if nil == e {
			return i
		}
		if v.IsNil() {
			return defaultValue
		}
		value = v.String()
	case string:
		sv, e := snmp.NewSnmpValue(v)
		if nil == e {
			i, e := snmp.AsInt(sv)
			if nil == e {
				return i
			}
			if sv.IsNil() {
				return defaultValue
			}
		}
	default:
		i, e := as.AsInt(value)
		if nil == e {
			return i
		}
	}
	panic(fmt.Sprintf("row with key is '%s' cann`t convert to int, value is `%v`.", idx, value))
}

func GetStringColumn(params map[string]string, old_row RECORD, idx string) string {
	value, ok := old_row[idx]
	if !ok {
		panic("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmp.SnmpValue:
		return SnmpValueToString(params, v)
	case string:
		sv, e := snmp.NewSnmpValue(v)
		if nil == e {
			return SnmpValueToString(params, sv)
		}
	default:
		s, e := as.AsString(value)
		if nil != e {
			return s
		}
	}
	panic(fmt.Sprintf("row with key is '%s' cann`t convert to uint64, value is `%v`.", idx, value))
}

func GetHardwareAddressColumn(params map[string]string, old_row RECORD, idx string) string {
	value, ok := old_row[idx]
	if !ok {
		panic("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmp.SnmpValue:
		if snmp.SNMP_SYNTAX_OCTETSTRING == v.GetSyntax() {
			return v.GetString()
		}
		value = v.String()
	case string:
		if strings.HasPrefix(v, "[octets]") {
			return v[8:]
		}
		return v
	}
	panic(fmt.Sprintf("row with key is '%s' cann`t convert to uint64, value is `%v`.", idx, value))
}

func GetUint64Column(params map[string]string, old_row RECORD, idx string, defaultValue uint64) uint64 {
	value, ok := old_row[idx]
	if !ok {
		panic("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmp.SnmpValue:
		i, e := snmp.AsUint64(v)
		if nil == e {
			return i
		}
		if v.IsNil() {
			return defaultValue
		}
		value = v.String()
	case string:
		sv, e := snmp.NewSnmpValue(v)
		if nil == e {
			i, e := snmp.AsUint64(sv)
			if nil == e {
				return i
			}
			if sv.IsNil() {
				return defaultValue
			}
		}
	default:
		i, e := as.AsUint64(value)
		if nil == e {
			return i
		}
	}
	panic(fmt.Sprintf("row with key is '%s' cann`t convert to uint64, value is `%v`.", idx, value))
}

func GetCopyWithPrefix(params map[string]string, prefix string) map[string]string {
	cp := map[string]string{}
	for k, v := range params {
		cp[k] = v
	}
	return GetWithPrefix(cp, prefix)
}

func GetWithPrefix(params map[string]string, prefix string) map[string]string {
	for k, v := range params {
		if strings.HasPrefix(k, prefix) {
			params[k[len(prefix):]] = v
		}
	}
	return params
}

func toOid(value string) string {
	if !strings.HasPrefix(value, "[oid") {
		return value
	}

	return value[5:]
}

func SnmpValueToString(params map[string]string, v snmp.SnmpValue) string {
	if snmp.SNMP_SYNTAX_OCTETSTRING != v.GetSyntax() {
		return v.String()
	}
	return bytesToString(params, v.GetBytes())
}

func AnyToString(params map[string]string, v interface{}) string {
	s, e := as.AsString(v)
	if nil != e {
		return "**error** " + e.Error()
	}
	if !strings.HasPrefix(s, "[octets") {
		return s
	}
	s = s[8:]
	bs, e := hex.DecodeString(s)
	if nil != e {
		return "**error** convert from hex failed, " + e.Error()
	}

	return bytesToString(params, bs)
}

func bytesToString(params map[string]string, bs []byte) string {
	charset, _ := params["charset"]
	decoder := mahonia.NewDecoder(charset)
	if nil == decoder {
		return string(bs)
	}

	var buffer bytes.Buffer
	for 0 != len(bs) {
		c, length, status := decoder(bs)
		switch status {
		case mahonia.SUCCESS:
			buffer.WriteRune(c)
			bs = bs[length:]
		case mahonia.INVALID_CHAR:
			buffer.Write([]byte{'.'})
			bs = bs[1:]
		case mahonia.NO_ROOM:
			buffer.Write([]byte{'.'})
			bs = bs[0:0]
		case mahonia.STATE_ONLY:
			bs = bs[length:]
		}
	}
	return buffer.String()
}

func (self *Base) GetString(params map[string]string, key string) (map[string]interface{}, commons.RuntimeError) {
	res, err := self.drv.Get(params)
	if nil == res || nil != err {
		return res, err
	}
	rv := commons.GetReturn(res)
	if nil == rv {
		return res, err
	}
	values, ok := rv.(map[string]snmp.SnmpValue)
	if !ok {
		values2, ok := rv.(map[string]interface{})
		if !ok {
			return nil, commons.NewRuntimeError(commons.InternalErrorCode, fmt.Sprintf("snmp result must is not a map[string]string - [%T]%v.", rv, rv))
		}
		value2, ok := values2[key]
		if !ok {
			return res, commons.NewRuntimeError(commons.InternalErrorCode, "value is not found.")
		}
		s := AnyToString(params, value2)
		return commons.ReturnWithValue(res, s), err

		return res, commons.NewRuntimeError(commons.InternalErrorCode, fmt.Sprintf("snmp result must is not a map[string]snmp.SnmpValue - [%T]%v.", rv, rv))
	}
	value, ok := values[key]
	if !ok {
		return res, commons.NewRuntimeError(commons.InternalErrorCode, "value is not found.")
	}
	s := SnmpValueToString(params, value)
	return commons.ReturnWithValue(res, s), err
}

func (self *Base) GetTable(params map[string]string, cb func(table map[string]RECORD, key string, old_row RECORD) error) (map[string]interface{}, commons.RuntimeError) {
	res, err := self.drv.Get(params)
	if nil == res || nil != err {
		return res, err
	}
	rv := commons.GetReturn(res)
	if nil == rv {
		return res, err
	}
	values, ok := rv.(map[string]map[string]snmp.SnmpValue)
	if !ok {
		values2, ok := rv.(map[string]interface{})
		if !ok {
			return res, commons.NewRuntimeError(commons.InternalErrorCode,
				fmt.Sprintf("snmp result must is not a map[string]map[string]snmp.SnmpValue or map[string]interface{} - [%T]%v.", rv, rv))
		}

		table := map[string]RECORD{}
		for key, r := range values2 {
			row, ok := r.(RECORD)
			if !ok {
				return res, commons.NewRuntimeError(commons.InternalErrorCode, "row with key is '"+key+"' process failed, it is not a RECORD")
			}
			e := cb(table, key, row)
			if nil != e {
				return res, commons.NewRuntimeError(commons.InternalErrorCode, "row with key is '"+key+"' process failed, "+e.Error())
			}
		}
		return commons.ReturnWithValue(res, table), err
	}
	table := map[string]RECORD{}
	for key, r := range values {
		row := RECORD{}
		for k, v := range r {
			row[k] = v
		}
		e := cb(table, key, row)
		if nil != e {
			return res, commons.NewRuntimeError(commons.InternalErrorCode, "row with key is '"+key+"' process failed, "+e.Error())
		}
	}
	return commons.ReturnWithValue(res, table), err
}

func (self *Base) GetOid(params map[string]string, key string) (map[string]interface{}, commons.RuntimeError) {
	res, err := self.drv.Get(params)
	if nil == res || nil != err {
		return res, err
	}
	rv := commons.GetReturn(res)
	if nil == rv {
		return res, err
	}
	values, ok := rv.(map[string]snmp.SnmpValue)
	if !ok {
		values2, ok := rv.(map[string]interface{})
		if !ok {
			return nil, commons.NewRuntimeError(commons.InternalErrorCode, fmt.Sprintf("snmp result must is not a map[string]string - [%T]%v.", rv, rv))
		}
		value2, ok := values2[key]
		if !ok {
			return res, commons.NewRuntimeError(commons.InternalErrorCode, "value is not found.")
		}
		s := toOid(fmt.Sprint(value2))
		return commons.ReturnWithValue(res, s), err
	}
	value, ok := values[key]
	if !ok {
		return res, commons.NewRuntimeError(commons.InternalErrorCode, "value is not found.")
	}
	if snmp.SNMP_SYNTAX_OID != value.GetSyntax() {
		return res, commons.NewRuntimeError(commons.InternalErrorCode, "value is not OID.")
	}
	s := value.GetString()
	return commons.ReturnWithValue(res, s), err
}

func (self *Base) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return nil, commons.NotImplemented
}

func (self *Base) Put(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return nil, commons.NotImplemented
}

func (self *Base) Create(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return nil, commons.NotImplemented
}

func (self *Base) Delete(params map[string]string) (bool, commons.RuntimeError) {
	return false, commons.NotImplemented
}

type SystemOid struct {
	Base
}

func (self *SystemOid) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	pa := GetWithPrefix(params, "snmp.")
	pa["oid"] = "1.3.6.1.2.1.1.2.0"
	pa["action"] = "get"
	return self.GetOid(pa, "1.3.6.1.2.1.1.2.0")
}

type SystemDescr struct {
	Base
}

func (self *SystemDescr) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	pa := GetWithPrefix(params, "snmp.")
	pa["oid"] = "1.3.6.1.2.1.1.1.0"
	pa["action"] = "get"
	return self.GetString(pa, "1.3.6.1.2.1.1.1.0")
}

type Interface struct {
	Base
}

func (self *Interface) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	pa := GetWithPrefix(params, "snmp.")
	pa["oid"] = "1.3.6.1.2.1.2.2.1"
	pa["action"] = "table"
	//pa["columns"] = "1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21"

	return self.GetTable(pa, func(table map[string]RECORD, key string, old_row RECORD) error {
		new_row := RECORD{}
		new_row["ifIndex"] = GetIntColumn(params, old_row, "1", -1)
		new_row["ifDescr"] = GetStringColumn(params, old_row, "2")
		new_row["ifType"] = GetIntColumn(params, old_row, "3", -1)
		new_row["ifMtu"] = GetIntColumn(params, old_row, "4", -1)
		new_row["ifSpeed"] = GetUint64Column(params, old_row, "5", 0)
		new_row["ifPhysAddress"] = GetHardwareAddressColumn(params, old_row, "6")
		new_row["ifAdminStatus"] = GetIntColumn(params, old_row, "7", -1)
		new_row["ifOpStatus"] = GetIntColumn(params, old_row, "8", -1)
		new_row["ifLastChange"] = GetIntColumn(params, old_row, "9", -1)
		new_row["ifInOctets"] = GetUint64Column(params, old_row, "10", 0)
		new_row["ifInUcastPkts"] = GetUint64Column(params, old_row, "11", 0)
		new_row["ifInNUcastPkts"] = GetUint64Column(params, old_row, "12", 0)
		new_row["ifInDiscards"] = GetUint64Column(params, old_row, "13", 0)
		new_row["ifInErrors"] = GetUint64Column(params, old_row, "14", 0)
		new_row["ifInUnknownProtos"] = GetUint64Column(params, old_row, "15", 0)
		new_row["ifOutOctets"] = GetUint64Column(params, old_row, "16", 0)
		new_row["ifOutUcastPkts"] = GetUint64Column(params, old_row, "17", 0)
		new_row["ifOutNUcastPkts"] = GetUint64Column(params, old_row, "18", 0)
		new_row["ifOutDiscards"] = GetUint64Column(params, old_row, "19", 0)
		new_row["ifOutErrors"] = GetUint64Column(params, old_row, "20", 0)
		new_row["ifOutQLen"] = GetUint64Column(params, old_row, "21", 0)
		table[key] = new_row
		return nil
	})
}

func init() {
	commons.METRIC_DRVS["sys.oid"] = func(params map[string]string, drvMgr *commons.DriverManager) (commons.Driver, commons.RuntimeError) {
		drv := &SystemOid{}
		return drv, drv.Init(drvMgr, "snmp")
	}
	commons.METRIC_DRVS["sys.descr"] = func(params map[string]string, drvMgr *commons.DriverManager) (commons.Driver, commons.RuntimeError) {
		drv := &SystemDescr{}
		return drv, drv.Init(drvMgr, "snmp")
	}
	commons.METRIC_DRVS["interface"] = func(params map[string]string, drvMgr *commons.DriverManager) (commons.Driver, commons.RuntimeError) {
		drv := &Interface{}
		return drv, drv.Init(drvMgr, "snmp")
	}
}
