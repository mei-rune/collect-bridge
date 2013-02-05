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

func (self *Base) Init(drvMgr *commons.DriverManager, drvName string) commons.RuntimeError {
	self.drvMgr = drvMgr
	self.drv, _ = drvMgr.Connect(drvName)
	if nil == self.drv {
		return commons.NotFound(drvName)
	}
	return nil
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
		return res, commons.NewRuntimeError(commons.InternalErrorCode, fmt.Sprintf("snmp result must is not a map[string]string - [%T]%v.", rv, rv))
	}
	value, ok := values[key]
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
	}
	s := SnmpValueToString(params, value)
	return commons.ReturnWithValue(res, s), err
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

	res, err := self.drv.Get(pa)
	if nil == res {
		return res, err
	}
	// v := commons.GetReturn(res)
	// if nil == v {
	// 	return res, err
	// }
	// values, ok := v.(map[string]snmp.SnmpValue)
	// if !ok {
	// 	return nil, commons.NewRuntimeError(commons.InternalErrorCode, fmt.Sprintf("snmp result must is not a map[string]string - [%T]%v.", v, v))
	// }

	return res, err
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
