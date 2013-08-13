package sampling

import (
	"bytes"
	"code.google.com/p/mahonia"
	"commons"
	"errors"
	"fmt"
	"github.com/runner-mei/snmpclient"
	"strings"
)

func nilString(b []byte) string {
	i := bytes.IndexByte(b, byte(0))
	if -1 == i {
		return string(b)
	}
	return string(b[0:i])
}

func bytesToString(params commons.Map, input []byte) string {
	if nil == input || 0 == len(input) {
		return ""
	}

	bs := input
	if byte(0) == bs[len(bs)-1] {
		bs = bs[:len(bs)-1]
	}

	charset := params.GetStringWithDefault("charset", "GB18030")
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

func GetInt32(params commons.Map, values map[string]interface{}, idx string, defaultValue int32) int32 {
	i, e := TryGetInt32(params, values, idx, defaultValue)
	if nil != e {
		panic(e.Error())
	}
	return i
}

func TryGetInt32(params commons.Map, values map[string]interface{}, idx string, defaultValue int32) (int32, error) {
	value, ok := values[idx]
	if !ok {
		return defaultValue, nil //errors.New("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmpclient.SnmpValue:
		i, e := snmpclient.AsInt32(v)
		if nil == e {
			return i, nil
		}
		if v.IsNil() {
			return defaultValue, nil
		}
		value = v.String()
	case string:
		sv, e := snmpclient.NewSnmpValue(v)
		if nil == e {
			i, e := snmpclient.AsInt32(sv)
			if nil == e {
				return i, nil
			}
			if sv.IsNil() {
				return defaultValue, nil
			}
		}
	default:
		i, e := commons.AsInt32(value)
		if nil == e {
			return i, nil
		}
	}
	return defaultValue, fmt.Errorf("row with key is '%s' cann`t convert to int32, value is `%v`.", idx, value)
}

func GetUint32(params commons.Map, values map[string]interface{}, idx string, defaultValue uint32) uint32 {
	i, e := TryGetUint32(params, values, idx, defaultValue)
	if nil != e {
		panic(e.Error())
	}
	return i
}

func TryGetUint32(params commons.Map, values map[string]interface{}, idx string, defaultValue uint32) (uint32, error) {
	value, ok := values[idx]
	if !ok {
		return defaultValue, nil //errors.New("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmpclient.SnmpValue:
		i, e := snmpclient.AsUint32(v)
		if nil == e {
			return i, nil
		}
		if v.IsNil() {
			return defaultValue, nil
		}
		value = v.String()
	case string:
		sv, e := snmpclient.NewSnmpValue(v)
		if nil == e {
			i, e := snmpclient.AsUint32(sv)
			if nil == e {
				return i, nil
			}
			if sv.IsNil() {
				return defaultValue, nil
			}
		}
	default:
		i, e := commons.AsUint32(value)
		if nil == e {
			return i, nil
		}
	}
	return defaultValue, fmt.Errorf("row with key is '%s' cann`t convert to uint32, value is `%v`.", idx, value)
}

func GetInt64(params commons.Map, values map[string]interface{}, idx string, defaultValue int64) int64 {
	i, e := TryGetInt64(params, values, idx, defaultValue)
	if nil != e {
		panic(e.Error())
	}
	return i
}

func TryGetInt64(params commons.Map, values map[string]interface{}, idx string, defaultValue int64) (int64, error) {
	value, ok := values[idx]
	if !ok {
		return defaultValue, nil //errors.New("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmpclient.SnmpValue:
		i, e := snmpclient.AsInt64(v)
		if nil == e {
			return i, nil
		}
		if v.IsNil() {
			return defaultValue, nil
		}
		value = v.String()
	case string:
		sv, e := snmpclient.NewSnmpValue(v)
		if nil == e {
			i, e := snmpclient.AsInt64(sv)
			if nil == e {
				return i, nil
			}
			if sv.IsNil() {
				return defaultValue, nil
			}
		}
	default:
		i, e := commons.AsInt64(value)
		if nil == e {
			return i, nil
		}
	}
	return defaultValue, fmt.Errorf("row with key is '%s' cann`t convert to int64, value is `%v`.", idx, value)
}

func GetUint64(params commons.Map, values map[string]interface{}, idx string, defaultValue uint64) uint64 {
	i, e := TryGetUint64(params, values, idx, defaultValue)
	if nil != e {
		panic(e.Error())
	}
	return i
}

func TryGetUint64(params commons.Map, values map[string]interface{}, idx string, defaultValue uint64) (uint64, error) {
	value, ok := values[idx]
	if !ok {
		return defaultValue, nil //errors.New("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmpclient.SnmpValue:
		i, e := snmpclient.AsUint64(v)
		if nil == e {
			return i, nil
		}
		if v.IsNil() {
			return defaultValue, nil
		}
		value = v.String()
	case string:
		sv, e := snmpclient.NewSnmpValue(v)
		if nil == e {
			i, e := snmpclient.AsUint64(sv)
			if nil == e {
				return i, nil
			}
			if sv.IsNil() {
				return defaultValue, nil
			}
		}
	default:
		i, e := commons.AsUint64(value)
		if nil == e {
			return i, nil
		}
	}
	return defaultValue, fmt.Errorf("row with key is '%s' cann`t convert to uint64, value is `%v`.", idx, value)
}

func GetOid(params commons.Map, values map[string]interface{}, idx string) string {
	s, e := TryGetOid(params, values, idx)
	if nil != e {
		panic(e.Error())
	}
	return s
}

func TryGetOid(params commons.Map, values map[string]interface{}, idx string) (string, error) {
	value, ok := values[idx]
	if !ok {
		return "", nil //errors.New("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmpclient.SnmpValue:
		if snmpclient.SNMP_SYNTAX_OID == v.GetSyntax() {
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

func GetString(params commons.Map, values map[string]interface{}, idx string) string {
	s, e := TryGetString(params, values, idx)
	if nil != e {
		panic(e.Error())
	}
	return s
}

func TryGetString(params commons.Map, values map[string]interface{}, idx string) (string, error) {
	value, ok := values[idx]
	if !ok {
		return "", nil //errors.New("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmpclient.SnmpValue:
		if snmpclient.SNMP_SYNTAX_OCTETSTRING != v.GetSyntax() {
			return "", errors.New("value is not string - '" + v.String() + "'.")
		}
		return bytesToString(params, v.GetBytes()), nil
	case string:
		sv, e := snmpclient.NewSnmpValue(v)
		if nil == e {
			if snmpclient.SNMP_SYNTAX_OCTETSTRING != sv.GetSyntax() {
				return "", errors.New("value is not string - '" + v + "'.")
			}
			return bytesToString(params, sv.GetBytes()), nil
		}
		return v, nil
	default:
		s, e := commons.AsString(value)
		if nil != e {
			return s, nil
		}
	}
	return "", fmt.Errorf("row with key is '%s' cann`t convert to string, value is `%v`.", idx, value)
}

func GetHardwareAddress(params commons.Map, values map[string]interface{}, idx string) string {
	s, e := TryGetHardwareAddress(params, values, idx)
	if nil != e {
		panic(e.Error())
	}
	return s
}

func parseMAC(s string) (string, error) {
	switch len(s) {
	case 0:
		return "", nil
	case 2:
		if "30" == s { // skip invalid address
			return "", nil
		}
	case 12:
		return s[:2] + ":" + s[2:4] + ":" + s[4:6] + ":" + s[6:8] + ":" + s[8:10] + ":" + s[10:], nil
	case 14:
		return s[:2] + ":" + s[2:4] + ":" + s[4:6] + ":" + s[6:8] + ":" + s[8:10] + ":" + s[10:12] + ":" + s[12:], nil
	case 16:
		return s[:2] + ":" + s[2:4] + ":" + s[4:6] + ":" + s[6:8] + ":" + s[8:10] + ":" + s[10:12] + ":" + s[12:14] + ":" + s[14:], nil
	}
	return "", errors.New("'" + s + "' is invalid hardware address")
}
func TryGetHardwareAddress(params commons.Map, values map[string]interface{}, idx string) (string, error) {
	value, ok := values[idx]
	if !ok {
		return "", nil //errors.New("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmpclient.SnmpValue:
		if snmpclient.SNMP_SYNTAX_OCTETSTRING == v.GetSyntax() {
			return parseMAC(v.GetString())
		}
		value = v.String()
	case string:
		if strings.HasPrefix(v, "[octets]") {
			return parseMAC(v[8:])
		}
		return v, nil
	}
	return "", fmt.Errorf("row with key is '%s' cann`t convert to hardwareAddress, value is `%v`.", idx, value)
}

func GetIPAddress(params commons.Map, values map[string]interface{}, idx string) string {
	s, e := TryGetIPAddress(params, values, idx)
	if nil != e {
		panic(e.Error())
	}
	return s
}

func TryGetIPAddress(params commons.Map, values map[string]interface{}, idx string) (string, error) {
	value, ok := values[idx]
	if !ok {
		return "", nil //errors.New("row with key is '" + idx + "' is not found")
	}
	switch v := value.(type) {
	case snmpclient.SnmpValue:
		if snmpclient.SNMP_SYNTAX_IPADDRESS == v.GetSyntax() {
			return v.GetString(), nil
		}
		value = v.String()
	case string:
		if strings.HasPrefix(v, "[ip]") {
			return v[4:], nil
		}
		return v, nil
	}
	return "", fmt.Errorf("row with key is '%s' cann`t convert to ipAddress, value is `%v`.", idx, value)
}
