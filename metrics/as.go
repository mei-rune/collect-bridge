package metrics

import (
	"bytes"
	"code.google.com/p/mahonia"
	"commons"
	"commons/as"
	"encoding/hex"
	"errors"
	"fmt"
	"snmp"
	"strings"
)

func toOid(value string) string {
	if !strings.HasPrefix(value, "[oid]") {
		return value
	}

	return value[5:]
}
func ToOidString(value string) (string, error) {
	if !strings.HasPrefix(value, "[oid]") {
		return "", errors.New("It is not a oid - '" + value + "'")
	}

	return value[5:], nil
}

func AsString(params map[string]string, v interface{}) (string, commons.RuntimeError) {
	s, e := as.AsString(v)
	if nil != e {
		return "", commons.NewRuntimeError(commons.InternalErrorCode, e.Error())
	}
	if !strings.HasPrefix(s, "[octets") {
		return s, nil
	}
	s = s[8:]
	bs, e := hex.DecodeString(s)
	if nil != e {
		return "", commons.NewRuntimeError(commons.InternalErrorCode,
			"**error** convert from hex failed, "+e.Error())
	}

	return bytesToString(params, bs), nil
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

func GetString(params map[string]string, res map[string]interface{}, oid string) (string, commons.RuntimeError) {
	rv := commons.GetReturn(res)
	if nil == rv {
		return "", commons.ValueIsNil
	}

	switch values := rv.(type) {
	case map[string]snmp.SnmpValue:
		value, ok := values[oid]
		if !ok {
			return "", commons.NewRuntimeError(commons.InternalErrorCode, "value is not found.")
		}
		if snmp.SNMP_SYNTAX_OCTETSTRING == value.GetSyntax() {
			return bytesToString(params, value.GetBytes()), nil
		}
		return "", commons.NewRuntimeError(commons.InternalErrorCode,
			"value is not string - '"+value.String()+"'.")
	case map[string]interface{}:
		value, ok := values[oid]
		if !ok {
			return "", commons.NewRuntimeError(commons.InternalErrorCode, "value is not found.")
		}
		s, e := AsString(params, value)
		return s, e
	}
	return "", commons.NewRuntimeError(commons.InternalErrorCode,
		fmt.Sprintf("snmp result must is not a map[string]snmp.SnmpValue or map[string]interface{} - [%T]%v.", rv, rv))
}

func GetOid(params map[string]string, res map[string]interface{}, oid string) (string, commons.RuntimeError) {
	rv := commons.GetReturn(res)
	if nil == rv {
		return "", commons.ValueIsNil
	}

	switch values := rv.(type) {
	case map[string]snmp.SnmpValue:
		value, ok := values[oid]
		if !ok {
			return "", commons.NewRuntimeError(commons.InternalErrorCode, "value is not found.")
		}
		if snmp.SNMP_SYNTAX_OID != value.GetSyntax() {
			return "", commons.NewRuntimeError(commons.InternalErrorCode, "value is not OID.")
		}
		s := value.GetString()
		return s, nil
	case map[string]interface{}:
		value, ok := values[oid]
		if !ok {
			return "", commons.NewRuntimeError(commons.InternalErrorCode, "value is not found.")
		}
		s := toOid(fmt.Sprint(value))
		return s, nil
	}
	return "", commons.NewRuntimeError(commons.InternalErrorCode,
		fmt.Sprintf("snmp result must is not a map[string]snmp.SnmpValue or map[string]interface{} - [%T]%v.", rv, rv))
}

func GetInt32(params map[string]string, res map[string]interface{}, oid string, defaultValue int32) (int32, commons.RuntimeError) {
	rv := commons.GetReturn(res)
	if nil == rv {
		return defaultValue, commons.ValueIsNil
	}
	switch values := rv.(type) {
	case map[string]snmp.SnmpValue:
		value, ok := values[oid]
		if !ok {
			return defaultValue, commons.NewRuntimeError(commons.InternalErrorCode, "value is not found.")
		}
		i, e := snmp.AsInt32(value)
		if nil == e {
			return i, nil
		}
		if value.IsNil() {
			return defaultValue, commons.ValueIsNil
		}
		return defaultValue, commons.NewRuntimeError(commons.InternalErrorCode,
			"type of the value is error - "+value.String()+".")
	case map[string]interface{}:
		value, ok := values[oid]
		if !ok {
			return defaultValue, commons.NewRuntimeError(commons.InternalErrorCode, "value is not found.")
		}
		i, e := as.AsInt32(value)
		if nil != e {
			return defaultValue, commons.NewRuntimeError(commons.InternalErrorCode, e.Error())
		}
		return i, nil
	}
	return defaultValue, commons.NewRuntimeError(commons.InternalErrorCode,
		fmt.Sprintf("snmp result must is not a map[string]snmp.SnmpValue or map[string]interface{} - [%T]%v.", rv, rv))
}

func GetUint32(params map[string]string, res map[string]interface{}, oid string, defaultValue uint32) (uint32, commons.RuntimeError) {
	rv := commons.GetReturn(res)
	if nil == rv {
		return defaultValue, commons.ValueIsNil
	}
	switch values := rv.(type) {
	case map[string]snmp.SnmpValue:
		value, ok := values[oid]
		if !ok {
			return defaultValue, commons.NewRuntimeError(commons.InternalErrorCode, "value is not found.")
		}
		i, e := snmp.AsUint32(value)
		if nil == e {
			return i, nil
		}
		if value.IsNil() {
			return defaultValue, commons.ValueIsNil
		}
		return defaultValue, commons.NewRuntimeError(commons.InternalErrorCode,
			"type of the value is error - "+value.String()+".")
	case map[string]interface{}:
		value, ok := values[oid]
		if !ok {
			return defaultValue, commons.NewRuntimeError(commons.InternalErrorCode, "value is not found.")
		}
		i, e := as.AsUint32(value)
		if nil != e {
			return defaultValue, commons.NewRuntimeError(commons.InternalErrorCode, e.Error())
		}
		return i, nil
	}
	return defaultValue, commons.NewRuntimeError(commons.InternalErrorCode,
		fmt.Sprintf("snmp result must is not a map[string]snmp.SnmpValue or map[string]interface{} - [%T]%v.", rv, rv))
}

func GetInt32Column(params map[string]string, old_row RECORD, idx string, defaultValue int32) int32 {
	value, ok := old_row[idx]
	if !ok {
		panic("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmp.SnmpValue:
		i, e := snmp.AsInt32(v)
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
			i, e := snmp.AsInt32(sv)
			if nil == e {
				return i
			}
			if sv.IsNil() {
				return defaultValue
			}
		}
	default:
		i, e := as.AsInt32(value)
		if nil == e {
			return i
		}
	}
	panic(fmt.Sprintf("row with key is '%s' cann`t convert to int32, value is `%v`.", idx, value))
}

func GetUint32Column(params map[string]string, old_row RECORD, idx string, defaultValue uint32) uint32 {
	value, ok := old_row[idx]
	if !ok {
		panic("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmp.SnmpValue:
		i, e := snmp.AsUint32(v)
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
			i, e := snmp.AsUint32(sv)
			if nil == e {
				return i
			}
			if sv.IsNil() {
				return defaultValue
			}
		}
	default:
		i, e := as.AsUint32(value)
		if nil == e {
			return i
		}
	}
	panic(fmt.Sprintf("row with key is '%s' cann`t convert to uint32, value is `%v`.", idx, value))
}

func GetInt64Column(params map[string]string, old_row RECORD, idx string, defaultValue int64) int64 {
	value, ok := old_row[idx]
	if !ok {
		panic("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmp.SnmpValue:
		i, e := snmp.AsInt64(v)
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
			i, e := snmp.AsInt64(sv)
			if nil == e {
				return i
			}
			if sv.IsNil() {
				return defaultValue
			}
		}
	default:
		i, e := as.AsInt64(value)
		if nil == e {
			return i
		}
	}
	panic(fmt.Sprintf("row with key is '%s' cann`t convert to int64, value is `%v`.", idx, value))
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

func GetOidColumn(params map[string]string, old_row RECORD, idx string) string {
	value, ok := old_row[idx]
	if !ok {
		panic("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmp.SnmpValue:
		if snmp.SNMP_SYNTAX_OID == v.GetSyntax() {
			return v.GetString()
		}
		value = v.String()
	case string:
		s, e := ToOidString(v)
		if nil == e {
			return s
		}
	}
	panic(fmt.Sprintf("row with key is '%s' cann`t convert to oid, value is `%v`.", idx, value))
}

func GetStringColumn(params map[string]string, old_row RECORD, idx string) string {
	value, ok := old_row[idx]
	if !ok {
		panic("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmp.SnmpValue:
		if snmp.SNMP_SYNTAX_OCTETSTRING != v.GetSyntax() {
			panic("value is not string - '" + v.String() + "'.")
		}
		return bytesToString(params, v.GetBytes())
	case string:
		sv, e := snmp.NewSnmpValue(v)
		if nil == e {
			if snmp.SNMP_SYNTAX_OCTETSTRING != sv.GetSyntax() {
				panic("value is not string - '" + v + "'.")
			}
			return bytesToString(params, sv.GetBytes())
		}
		return v
	default:
		s, e := as.AsString(value)
		if nil != e {
			return s
		}
	}
	panic(fmt.Sprintf("row with key is '%s' cann`t convert to string, value is `%v`.", idx, value))
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
	panic(fmt.Sprintf("row with key is '%s' cann`t convert to hardwareAddress, value is `%v`.", idx, value))
}
