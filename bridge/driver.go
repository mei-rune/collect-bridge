package main

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	drivers = make(map[string]Driver)
)

func Register(name string, driver Driver) {
	_, ok := drivers[name]
	if ok {
		panic(fmt.Errorf("'%s' always registred.", name))
	}
	drivers[name] = driver
}

func Connect(name string) (Driver, bool) {
	driver, ok := drivers[name]
	return driver, ok
}

type Driver interface {
	Get(map[string]string) (interface{}, error)
	Put(map[string]string) (interface{}, error)
	Create(map[string]string) (bool, error)
	Delete(map[string]string) (bool, error)
}

// Map type asserts to `map`
func asMap(value interface{}) (map[string]interface{}, error) {
	if m, ok := value.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, errors.New("type assertion to map[string]interface{} failed")
}

// Array type asserts to an `array`
func asArray(value interface{}) ([]interface{}, error) {
	if a, ok := value.([]interface{}); ok {
		return a, nil
	}
	return nil, errors.New("type assertion to []interface{} failed")
}

// Bool type asserts to `bool`
func asBool(value interface{}) (bool, error) {
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
	return false, errors.New("type assertion to bool failed")
}

func asInt(value interface{}) (int, error) {
	a, err := asInt32(value)
	return int(a), err
}

func asUint(value interface{}) (uint, error) {
	a, err := asUint32(value)
	return uint(a), err
}

// Int type asserts to `float64` then converts to `int`
func asInt64(value interface{}) (int64, error) {
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
	return 0, errors.New("type assertion to int64 failed")
}

func asInt32(value interface{}) (int32, error) {
	i64, err := asInt64(value)
	if nil != err {
		return 0, errors.New("type assertion to int32 failed")
	}
	if -2147483648 > i64 || 2147483647 < i64 {
		return 0, errors.New("type assertion to int32 failed, it is too big.")
	}
	return int32(i64), nil
}

func asInt16(value interface{}) (int16, error) {
	i64, err := asInt64(value)
	if nil != err {
		return 0, errors.New("type assertion to int16 failed")
	}
	if -32768 > i64 || 32767 < i64 {
		return 0, errors.New("type assertion to int16 failed, it is too big.")
	}
	return int16(i64), nil
}

func asInt8(value interface{}) (int8, error) {
	i64, err := asInt64(value)
	if nil != err {
		return 0, errors.New("type assertion to int8 failed")
	}
	if -128 > i64 || 127 < i64 {
		return 0, errors.New("type assertion to int8 failed, it is too big.")
	}
	return int8(i64), nil
}

// Uint type asserts to `float64` then converts to `int`
func asUint64(value interface{}) (uint64, error) {
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
	}
	return 0, errors.New("type assertion to uint64 failed")
}

func asUint32(value interface{}) (uint32, error) {
	ui64, err := asUint64(value)
	if nil != err {
		return 0, errors.New("type assertion to uint32 failed")
	}
	if 4294967295 < ui64 {
		return 0, errors.New("type assertion to uint32 failed, it is too big.")
	}
	return uint32(ui64), nil
}

func asUint16(value interface{}) (uint16, error) {
	ui64, err := asUint64(value)
	if nil != err {
		return 0, errors.New("type assertion to uint16 failed")
	}
	if 65535 < ui64 {
		return 0, errors.New("type assertion to uint16 failed, it is too big.")
	}
	return uint16(ui64), nil
}

func asUint8(value interface{}) (uint8, error) {
	ui64, err := asUint64(value)
	if nil != err {
		return 0, errors.New("type assertion to uint8 failed")
	}
	if 255 < ui64 {
		return 0, errors.New("type assertion to uint8 failed, it is too big.")
	}
	return uint8(ui64), nil
}

// Uint type asserts to `float64` then converts to `int`
func asFloat64(value interface{}) (float64, error) {
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
	}
	return 0, errors.New("type assertion to float64 failed")
}

func asFloat32(value interface{}) (float32, error) {
	f64, err := asFloat64(value)
	if nil != err {
		return 0, errors.New("type assertion to float32 failed")
	}
	return float32(f64), nil
}

// String type asserts to `string`
func asString(value interface{}) (string, error) {
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
	}
	return "", errors.New("type assertion to string failed")
}
