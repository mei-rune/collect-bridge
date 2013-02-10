package metrics

import (
	"bytes"
	"code.google.com/p/mahonia"
	"commons/as"
	"errors"
	"fmt"
	"snmp"
	"strings"
)

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

func GetInt32(params map[string]string, values map[string]interface{}, idx string, defaultValue int32) int32 {
	i, e := TryGetInt32(params, values, idx, defaultValue)
	if nil != e {
		panic(e.Error())
	}
	return i
}

func TryGetInt32(params map[string]string, values map[string]interface{}, idx string, defaultValue int32) (int32, error) {
	value, ok := values[idx]
	if !ok {
		return defaultValue, errors.New("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmp.SnmpValue:
		i, e := snmp.AsInt32(v)
		if nil == e {
			return i, nil
		}
		if v.IsNil() {
			return defaultValue, nil
		}
		value = v.String()
	case string:
		sv, e := snmp.NewSnmpValue(v)
		if nil == e {
			i, e := snmp.AsInt32(sv)
			if nil == e {
				return i, nil
			}
			if sv.IsNil() {
				return defaultValue, nil
			}
		}
	default:
		i, e := as.AsInt32(value)
		if nil == e {
			return i, nil
		}
	}
	return defaultValue, fmt.Errorf("row with key is '%s' cann`t convert to int32, value is `%v`.", idx, value)
}

func GetUint32(params map[string]string, values map[string]interface{}, idx string, defaultValue uint32) uint32 {
	i, e := TryGetUint32(params, values, idx, defaultValue)
	if nil != e {
		panic(e.Error())
	}
	return i
}

func TryGetUint32(params map[string]string, values map[string]interface{}, idx string, defaultValue uint32) (uint32, error) {
	value, ok := values[idx]
	if !ok {
		return defaultValue, errors.New("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmp.SnmpValue:
		i, e := snmp.AsUint32(v)
		if nil == e {
			return i, nil
		}
		if v.IsNil() {
			return defaultValue, nil
		}
		value = v.String()
	case string:
		sv, e := snmp.NewSnmpValue(v)
		if nil == e {
			i, e := snmp.AsUint32(sv)
			if nil == e {
				return i, nil
			}
			if sv.IsNil() {
				return defaultValue, nil
			}
		}
	default:
		i, e := as.AsUint32(value)
		if nil == e {
			return i, nil
		}
	}
	return defaultValue, fmt.Errorf("row with key is '%s' cann`t convert to uint32, value is `%v`.", idx, value)
}

func GetInt64(params map[string]string, values map[string]interface{}, idx string, defaultValue int64) int64 {
	i, e := TryGetInt64(params, values, idx, defaultValue)
	if nil != e {
		panic(e.Error())
	}
	return i
}

func TryGetInt64(params map[string]string, values map[string]interface{}, idx string, defaultValue int64) (int64, error) {
	value, ok := values[idx]
	if !ok {
		return defaultValue, errors.New("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmp.SnmpValue:
		i, e := snmp.AsInt64(v)
		if nil == e {
			return i, nil
		}
		if v.IsNil() {
			return defaultValue, nil
		}
		value = v.String()
	case string:
		sv, e := snmp.NewSnmpValue(v)
		if nil == e {
			i, e := snmp.AsInt64(sv)
			if nil == e {
				return i, nil
			}
			if sv.IsNil() {
				return defaultValue, nil
			}
		}
	default:
		i, e := as.AsInt64(value)
		if nil == e {
			return i, nil
		}
	}
	return defaultValue, fmt.Errorf("row with key is '%s' cann`t convert to int64, value is `%v`.", idx, value)
}

func GetUint64(params map[string]string, values map[string]interface{}, idx string, defaultValue uint64) uint64 {
	i, e := TryGetUint64(params, values, idx, defaultValue)
	if nil != e {
		panic(e.Error())
	}
	return i
}

func TryGetUint64(params map[string]string, values map[string]interface{}, idx string, defaultValue uint64) (uint64, error) {
	value, ok := values[idx]
	if !ok {
		return defaultValue, errors.New("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmp.SnmpValue:
		i, e := snmp.AsUint64(v)
		if nil == e {
			return i, nil
		}
		if v.IsNil() {
			return defaultValue, nil
		}
		value = v.String()
	case string:
		sv, e := snmp.NewSnmpValue(v)
		if nil == e {
			i, e := snmp.AsUint64(sv)
			if nil == e {
				return i, nil
			}
			if sv.IsNil() {
				return defaultValue, nil
			}
		}
	default:
		i, e := as.AsUint64(value)
		if nil == e {
			return i, nil
		}
	}
	return defaultValue, fmt.Errorf("row with key is '%s' cann`t convert to uint64, value is `%v`.", idx, value)
}

func GetOid(params map[string]string, values map[string]interface{}, idx string) string {
	s, e := TryGetOid(params, values, idx)
	if nil != e {
		panic(e.Error())
	}
	return s
}

func TryGetOid(params map[string]string, values map[string]interface{}, idx string) (string, error) {
	value, ok := values[idx]
	if !ok {
		return "", errors.New("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmp.SnmpValue:
		if snmp.SNMP_SYNTAX_OID == v.GetSyntax() {
			return v.GetString(), nil
		}
		value = v.String()
	case string:
		if strings.HasPrefix(v, "[oid]") {
			return v[5:], nil
		}
	}
	panic(fmt.Sprintf("row with key is '%s' cann`t convert to oid, value is `%v`.", idx, value))
}

func GetString(params map[string]string, values map[string]interface{}, idx string) string {
	s, e := TryGetString(params, values, idx)
	if nil != e {
		panic(e.Error())
	}
	return s
}

func TryGetString(params map[string]string, values map[string]interface{}, idx string) (string, error) {
	value, ok := values[idx]
	if !ok {
		return "", errors.New("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmp.SnmpValue:
		if snmp.SNMP_SYNTAX_OCTETSTRING != v.GetSyntax() {
			return "", errors.New("value is not string - '" + v.String() + "'.")
		}
		return bytesToString(params, v.GetBytes()), nil
	case string:
		sv, e := snmp.NewSnmpValue(v)
		if nil == e {
			if snmp.SNMP_SYNTAX_OCTETSTRING != sv.GetSyntax() {
				return "", errors.New("value is not string - '" + v + "'.")
			}
			return bytesToString(params, sv.GetBytes()), nil
		}
		return v, nil
	default:
		s, e := as.AsString(value)
		if nil != e {
			return s, nil
		}
	}
	return "", fmt.Errorf("row with key is '%s' cann`t convert to string, value is `%v`.", idx, value)
}

func GetHardwareAddress(params map[string]string, values map[string]interface{}, idx string) string {
	s, e := TryGetHardwareAddress(params, values, idx)
	if nil != e {
		panic(e.Error())
	}
	return s
}

func TryGetHardwareAddress(params map[string]string, values map[string]interface{}, idx string) (string, error) {
	value, ok := values[idx]
	if !ok {
		return "", errors.New("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmp.SnmpValue:
		if snmp.SNMP_SYNTAX_OCTETSTRING == v.GetSyntax() {
			return v.GetString(), nil
		}
		value = v.String()
	case string:
		if strings.HasPrefix(v, "[octets]") {
			return v[8:], nil
		}
		return v, nil
	}
	return "", fmt.Errorf("row with key is '%s' cann`t convert to hardwareAddress, value is `%v`.", idx, value)
}
