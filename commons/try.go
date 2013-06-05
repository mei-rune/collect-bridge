package commons

import (
	"commons/as"
	"fmt"
)

var MultipleValuesError = NewRuntimeError(NotAcceptableCode, "Multiple values meet the conditions")

func nilError(key string) RuntimeError {
	return NewRuntimeError(InternalErrorCode, "value of '"+key+"'is nil.")
}

func typeError(key, t string) RuntimeError {
	return NewRuntimeError(InternalErrorCode, "value of '"+key+"'is not a "+t+".")
}

func GetBool(attributes map[string]interface{}, key string, defaultValue bool) bool {
	res, e := TryGetBool(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func TryGetBool(attributes map[string]interface{}, key string) (bool, error) {
	v, ok := attributes[key]
	if !ok {
		return false, NotFound(key)
	}
	if nil == v {
		return false, nilError(key)
	}
	b, e := as.AsBool(v)
	if nil != e {
		return false, typeError(key, "bool")
	}
	return b, nil
}

func GetInt(attributes map[string]interface{}, key string, defaultValue int) int {
	res, e := TryGetInt(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func TryGetInt(attributes map[string]interface{}, key string) (int, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, NotFound(key)
	}
	if nil == v {
		return 0, nilError(key)
	}
	i, e := as.AsInt(v)
	if nil != e {
		return 0, typeError(key, "int")
	}
	return i, nil
}

func GetUint(attributes map[string]interface{}, key string, defaultValue uint) uint {
	res, e := TryGetUint(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func TryGetUint(attributes map[string]interface{}, key string) (uint, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, NotFound(key)
	}
	if nil == v {
		return 0, nilError(key)
	}
	ui, e := as.AsUint(v)
	if nil != e {
		return 0, typeError(key, "uint")
	}
	return ui, nil
}

func GetFloat(attributes map[string]interface{}, key string, defaultValue float64) float64 {
	res, e := TryGetFloat(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func TryGetFloat(attributes map[string]interface{}, key string) (float64, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, NotFound(key)
	}
	if nil == v {
		return 0, nilError(key)
	}
	f, e := as.AsFloat64(v)
	if nil != e {
		return 0, typeError(key, "float")
	}
	return f, nil
}

func GetInt32(attributes map[string]interface{}, key string, defaultValue int32) int32 {
	res, e := TryGetInt32(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func TryGetInt32(attributes map[string]interface{}, key string) (int32, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, NotFound(key)
	}
	if nil == v {
		return 0, nilError(key)
	}
	i32, e := as.AsInt32(v)
	if nil != e {
		return 0, typeError(key, "int32")
	}
	return i32, nil
}

func GetInt64(attributes map[string]interface{}, key string, defaultValue int64) int64 {
	res, e := TryGetInt64(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func TryGetInt64(attributes map[string]interface{}, key string) (int64, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, NotFound(key)
	}
	if nil == v {
		return 0, nilError(key)
	}
	i64, e := as.AsInt64(v)
	if nil != e {
		return 0, typeError(key, "int64")
	}
	return i64, nil
}

func GetUint32(attributes map[string]interface{}, key string, defaultValue uint32) uint32 {
	res, e := TryGetUint32(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func TryGetUint32(attributes map[string]interface{}, key string) (uint32, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, NotFound(key)
	}
	if nil == v {
		return 0, nilError(key)
	}
	ui32, e := as.AsUint32(v)
	if nil != e {
		return 0, typeError(key, "uint32")
	}
	return ui32, nil
}

func GetUint64(attributes map[string]interface{}, key string, defaultValue uint64) uint64 {
	res, e := TryGetUint64(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}
func TryGetUint64(attributes map[string]interface{}, key string) (uint64, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, NotFound(key)
	}
	if nil == v {
		return 0, nilError(key)
	}
	ui64, e := as.AsUint64(v)
	if nil != e {
		return 0, typeError(key, "uint64")
	}
	return ui64, nil
}
func GetString(attributes map[string]interface{}, key string, defaultValue string) string {
	res, e := TryGetString(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}
func TryGetString(attributes map[string]interface{}, key string) (string, error) {
	v, ok := attributes[key]
	if !ok {
		return "", NotFound(key)
	}
	if nil == v {
		return "", nilError(key)
	}
	s, e := as.AsString(v)
	if nil != e {
		return "", typeError(key, "string")
	}
	return s, nil
}

func GetArray(attributes map[string]interface{}, key string) []interface{} {
	v, ok := attributes[key]
	if !ok {
		return nil
	}

	if nil == v {
		return nil
	}

	res, ok := v.([]interface{})
	if !ok {
		return nil
	}
	return res
}

func GetObject(attributes map[string]interface{}, key string) map[string]interface{} {
	v, ok := attributes[key]
	if !ok {
		return nil
	}

	if nil == v {
		return nil
	}

	res, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	return res
}

func TryGetObject(attributes map[string]interface{}, key string) (map[string]interface{}, error) {
	v, ok := attributes[key]
	if !ok {
		return nil, NotFound(key)
	}

	if nil == v {
		return nil, ValueIsNil
	}

	res, ok := v.(map[string]interface{})
	if !ok {
		return nil, typeError(key, "map[string]interface{}")
	}
	return res, nil
}

func GetObjects(attributes map[string]interface{}, key string) []map[string]interface{} {
	v, ok := attributes[key]
	if !ok {
		return nil
	}

	if nil == v {
		return nil
	}

	results := make([]map[string]interface{}, 0, 10)
	switch value := v.(type) {
	case []interface{}:
		for _, o := range value {
			r, ok := o.(map[string]interface{})
			if !ok {
				return nil
			}
			results = append(results, r)
		}
	case map[string]interface{}:
		for _, o := range value {
			r, ok := o.(map[string]interface{})
			if !ok {
				return nil
			}
			results = append(results, r)
		}
	default:
		return nil
	}
	return results
}

func TryGetObjects(attributes map[string]interface{}, key string) ([]map[string]interface{}, error) {
	v, ok := attributes[key]
	if !ok {
		return nil, NotFound(key)
	}

	if nil == v {
		return nil, ValueIsNil
	}

	results := make([]map[string]interface{}, 0, 10)
	switch value := v.(type) {
	case []interface{}:
		for i, o := range value {
			r, ok := o.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("'%v' of '%s' is not a map[string]interface{}", i, key)
			}
			results = append(results, r)
		}
	case map[string]interface{}:
		for k, o := range value {
			r, ok := o.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("'%v' of '%s' is not a map[string]interface{}", k, key)
			}
			results = append(results, r)
		}
	default:
		return nil, typeError(key, "[]interface{} or map[string]interface{}")
	}
	return results, nil
}

// type MatchFunc func(key string, actual interface{}) bool

// func (self MatchFunc) Match(key string, actual interface{}) bool {
//	return self(key, actual, expected)
// }

type Matcher interface {
	Match(key string, actual interface{}) bool
}

type IntMatcher int

func (self *IntMatcher) Match(key string, actual interface{}) bool {
	i, e := as.AsInt(actual)
	if nil != e {
		return false
	}
	return i == int(*self)
}

func EqualInt(v int) Matcher {
	m := IntMatcher(v)
	return &m
}

func SearchBy(instance interface{}, query map[string]interface{}) []map[string]interface{} {
	if nil == instance {
		return nil
	}
	results := make([]map[string]interface{}, 0, 10)
	Each(instance, func(k interface{}, v interface{}) {
		r := v.(map[string]interface{})
		all := true
		for n, m := range query {
			value, ok := r[n]
			if !ok {
				all = false
				break
			}
			if matcher, ok := m.(Matcher); ok {
				if !matcher.Match(n, value) {
					all = false
					break
				}
				continue
			}
			if m != value {
				all = false
				break
			}
		}
		if all {
			results = append(results, r)
		}
	}, ThrowPanic)
	if 0 == len(results) {
		return nil
	}
	return results
}

func SearchOneBy(instance interface{}, query map[string]interface{}) map[string]interface{} {
	res := SearchBy(instance, query)
	if nil == res {
		return nil
	}
	if 1 != len(res) {
		return nil
	}
	return res[0]
}
