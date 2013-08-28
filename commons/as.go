package commons

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Map type AsSerts to `map`
func AsMap(value interface{}) (map[string]interface{}, error) {
	if m, ok := value.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, IsNotMap
}

func AsObject(value interface{}) (map[string]interface{}, error) {
	if m, ok := value.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, IsNotMap
}

func AsObjects(v interface{}) ([]map[string]interface{}, error) {
	results := make([]map[string]interface{}, 0, 10)
	switch value := v.(type) {
	case []interface{}:
		for i, o := range value {
			r, ok := o.(map[string]interface{})
			if !ok {
				return nil, TypeError(fmt.Sprintf("v['%s'] is not a map[string]interface{}", i))
			}
			results = append(results, r)
		}
	case map[string]interface{}:
		for k, o := range value {
			r, ok := o.(map[string]interface{})
			if !ok {
				return nil, TypeError(fmt.Sprintf("v['%s'] is not a map[string]interface{}", k))
			}
			results = append(results, r)
		}
	default:
		return nil, IsNotMapOrArray
	}
	return results, nil
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
	case json.Number:
		i64, err := v.Int64()
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
	case string:
		i64, err := strconv.ParseUint(v, 10, 64)
		if nil == err {
			return i64, nil
		}
		return i64, TypeError(err.Error())

	case json.Number:
		i64, err := strconv.ParseUint(v.String(), 10, 64)
		if nil == err {
			return i64, nil
		}
		return i64, TypeError(err.Error())
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
	case string:
		f64, err := strconv.ParseFloat(v, 64)
		if nil == err {
			return f64, nil
		}
		return f64, TypeError(err.Error())
	case json.Number:
		return v.Float64()
	case float32:
		return float64(v), nil
	case float64:
		return float64(v), nil
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
	case json.Number:
		return v.String(), nil
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

func AsTime(v interface{}) (time.Time, error) {
	if t, ok := v.(time.Time); ok {
		return t, nil
	}

	s, ok := v.(string)
	if !ok {
		return time.Time{}, ErrNotString
	}

	m, e := time.Parse(time.RFC3339, s)
	if nil == e {
		return m, nil
	}

	m, e = time.Parse(time.RFC3339Nano, s)
	if nil == e {
		return m, nil
	}
	return time.Time{}, ErrNotTimeString
}

func BoolWithDefault(v interface{}, defaultValue bool) bool {
	b, e := AsBool(v)
	if nil != e {
		return defaultValue
	}
	return b
}

func IntWithDefault(v interface{}, defaultValue int) int {
	i, e := AsInt(v)
	if nil != e {
		return defaultValue
	}
	return i
}

func Int32WithDefault(v interface{}, defaultValue int32) int32 {
	i32, e := AsInt32(v)
	if nil != e {
		return defaultValue
	}
	return i32
}

func Int64WithDefault(v interface{}, defaultValue int64) int64 {
	i64, e := AsInt64(v)
	if nil != e {
		return defaultValue
	}
	return i64
}

func UintWithDefault(v interface{}, defaultValue uint) uint {
	u, e := AsUint(v)
	if nil != e {
		return defaultValue
	}
	return u
}

func Uint32WithDefault(v interface{}, defaultValue uint32) uint32 {
	u32, e := AsUint32(v)
	if nil != e {
		return defaultValue
	}
	return u32
}

func Uint64WithDefault(v interface{}, defaultValue uint64) uint64 {
	u64, e := AsUint64(v)
	if nil != e {
		return defaultValue
	}
	return u64
}

func StringWithDefault(v interface{}, defaultValue string) string {
	s, e := AsString(v)
	if nil != e {
		return defaultValue
	}
	return s
}

func ArrayWithDefault(v interface{}, defaultValue []interface{}) []interface{} {
	if m, ok := v.([]interface{}); ok {
		return m
	}
	return defaultValue
}

func ObjectWithDefault(v interface{}, defaultValue map[string]interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return defaultValue
}

func ObjectsWithDefault(v interface{}, defaultValue []map[string]interface{}) []map[string]interface{} {
	if o, ok := v.([]map[string]interface{}); ok {
		return o
	}

	a, ok := v.([]interface{})
	if !ok {
		return defaultValue
	}

	res := make([]map[string]interface{}, 0, len(a))
	for _, value := range a {
		m, ok := value.(map[string]interface{})
		if !ok {
			return defaultValue
		}
		res = append(res, m)
	}
	return res
}
