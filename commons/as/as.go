package as

import (
	"errors"
	"strconv"
)

var (
	IsNotMap       = errors.New("type assertion to map[string]interface{} failed")
	IsNotArray     = errors.New("type assertion to []interface{} failed")
	IsNotBool      = errors.New("type assertion to bool failed")
	IsNotInt8      = errors.New("type assertion to int8 failed")
	IsNotInt16     = errors.New("type assertion to int16 failed")
	IsNotInt32     = errors.New("type assertion to int32 failed")
	IsNotInt64     = errors.New("type assertion to int64 failed")
	Int8OutRange   = errors.New("type assertion to int8 failed, out range")
	Int16OutRange  = errors.New("type assertion to int16 failed, out range")
	Int32OutRange  = errors.New("type assertion to int32 failed, out range")
	Int64OutRange  = errors.New("type assertion to int64 failed, out range")
	IsNotUint8     = errors.New("type assertion to uint8 failed")
	IsNotUint16    = errors.New("type assertion to uint16 failed")
	IsNotUint32    = errors.New("type assertion to uint32 failed")
	IsNotUint64    = errors.New("type assertion to uint64 failed")
	Uint8OutRange  = errors.New("type assertion to uint8 failed, out range")
	Uint16OutRange = errors.New("type assertion to uint16 failed, out range")
	Uint32OutRange = errors.New("type assertion to uint32 failed, out range")
	Uint64OutRange = errors.New("type assertion to uint64 failed, out range")

	IsNotFloat32 = errors.New("type assertion to float32 failed")
	IsNotFloat64 = errors.New("type assertion to float64 failed")
	IsNotString  = errors.New("type assertion to string failed")
)

// Map type AsSerts to `map`
func AsMap(value interface{}) (map[string]interface{}, error) {
	if m, ok := value.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, IsNotMap
}

// Array type AsSerts to an `array`
func AsArray(value interface{}) ([]interface{}, error) {
	if a, ok := value.([]interface{}); ok {
		return a, nil
	}
	return nil, IsNotArray
}

// Bool type AsSerts to `bool`
func AsBool(value interface{}) (bool, error) {
	if s, ok := value.(bool); ok {
		return s, nil
	}
	if s, ok := value.(string); ok {
		switch s {
		case "TRUE", "True", "true", "YES", "Yes", "yes":
			return true, nil
		case "FALSE", "False", "false", "NO", "No", "no":
			return false, nil
		}
	}
	return false, IsNotBool
}

// Bool type AsSerts to `bool`
func AsBoolWithDefaultValue(value interface{}, defaultValue bool) bool {
	if b, ok := value.(bool); ok {
		return b
	}
	if s, ok := value.(string); ok {
		switch s {
		case "TRUE", "True", "true", "YES", "Yes", "yes":
			return true
		case "FALSE", "False", "false", "NO", "No", "no":
			return false
		}
	}
	return defaultValue
}

func AsInt(value interface{}) (int, error) {
	a, err := AsInt32(value)
	return int(a), err
}

func AsUint(value interface{}) (uint, error) {
	a, err := AsUint32(value)
	return uint(a), err
}

// Int type AsSerts to `float64` then converts to `int`
func AsInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		if 9223372036854775807 >= int64(v) {
			return int64(v), nil
		}
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		if 9223372036854775807 >= v {
			return int64(v), nil
		}
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		i64, err := strconv.ParseInt(v, 10, 64)
		if nil == err {
			return i64, nil
		}
	}
	return 0, IsNotInt64
}

func AsInt32(value interface{}) (int32, error) {
	i64, err := AsInt64(value)
	if nil != err {
		return 0, IsNotInt32
	}
	if -2147483648 > i64 || 2147483647 < i64 {
		return 0, Int32OutRange
	}
	return int32(i64), nil
}

func AsInt16(value interface{}) (int16, error) {
	i64, err := AsInt64(value)
	if nil != err {
		return 0, IsNotInt16
	}
	if -32768 > i64 || 32767 < i64 {
		return 0, Int16OutRange
	}
	return int16(i64), nil
}

func AsInt8(value interface{}) (int8, error) {
	i64, err := AsInt64(value)
	if nil != err {
		return 0, IsNotInt8
	}
	if -128 > i64 || 127 < i64 {
		return 0, Int8OutRange
	}
	return int8(i64), nil
}

// Uint type AsSerts to `float64` then converts to `int`
func AsUint64(value interface{}) (uint64, error) {
	switch v := value.(type) {
	case uint:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint64:
		return uint64(v), nil
	case int:
		if v > 0 {
			return uint64(v), nil
		}
	case int8:
		if v > 0 {
			return uint64(v), nil
		}
	case int16:
		if v > 0 {
			return uint64(v), nil
		}
	case int32:
		if v > 0 {
			return uint64(v), nil
		}
	case int64:
		if v > 0 {
			return uint64(v), nil
		}
	case float32:
		if v > 0 && 18446744073709551615 >= v {
			return uint64(v), nil
		}
	case float64:
		if v > 0 && 18446744073709551615 >= v {
			return uint64(v), nil
		}
	case string:
		i64, err := strconv.ParseUint(v, 10, 64)
		if nil == err {
			return i64, nil
		}
		return i64, err
	}
	return 0, IsNotUint64
}

func AsUint32(value interface{}) (uint32, error) {
	ui64, err := AsUint64(value)
	if nil != err {
		return 0, IsNotUint32
	}
	if 4294967295 < ui64 {
		return 0, Uint32OutRange
	}
	return uint32(ui64), nil
}

func AsUint16(value interface{}) (uint16, error) {
	ui64, err := AsUint64(value)
	if nil != err {
		return 0, IsNotUint16
	}
	if 65535 < ui64 {
		return 0, Uint16OutRange
	}
	return uint16(ui64), nil
}

func AsUint8(value interface{}) (uint8, error) {
	ui64, err := AsUint64(value)
	if nil != err {
		return 0, IsNotUint8
	}
	if 255 < ui64 {
		return 0, Uint8OutRange
	}
	return uint8(ui64), nil
}

// Uint type AsSerts to `float64` then converts to `int`
func AsFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return float64(v), nil
	case string:
		f64, err := strconv.ParseFloat(v, 64)
		if nil == err {
			return f64, nil
		}
		return f64, err
	}
	return 0, IsNotFloat64
}

func AsFloat32(value interface{}) (float32, error) {
	f64, err := AsFloat64(value)
	if nil != err {
		return 0, IsNotFloat32
	}
	return float32(f64), nil
}

// String type AsSerts to `string`
func AsString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(uint64(v), 10), nil
	case int:
		return strconv.FormatInt(int64(v), 10), nil
	case int8:
		return strconv.FormatInt(int64(v), 10), nil
	case int16:
		return strconv.FormatInt(int64(v), 10), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case int64:
		return strconv.FormatInt(int64(v), 10), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'e', -1, 64), nil
	case float64:
		return strconv.FormatFloat(float64(v), 'e', -1, 64), nil
	case bool:
		if v {
			return "true", nil
		} else {
			return "false", nil
		}
	}
	return "", IsNotString
}
