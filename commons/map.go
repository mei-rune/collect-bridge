package commons

import (
	"errors"
	"fmt"
	"strconv"
)

var MultipleValuesError = NewRuntimeError(NotAcceptableCode, "Multiple values meet the conditions")

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
		return false, NotExists
	}
	if nil == v {
		return false, ValueIsNil
	}
	return AsBool(v)
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
		return 0, NotExists
	}
	if nil == v {
		return 0, ValueIsNil
	}
	return AsInt(v)
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
		return 0, NotExists
	}
	if nil == v {
		return 0, ValueIsNil
	}
	return AsUint(v)
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
		return 0, NotExists
	}
	if nil == v {
		return 0, ValueIsNil
	}
	return AsFloat64(v)
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
		return 0, NotExists
	}
	if nil == v {
		return 0, ValueIsNil
	}
	return AsInt32(v)
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
		return 0, NotExists
	}
	if nil == v {
		return 0, ValueIsNil
	}
	return AsInt64(v)
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
		return 0, NotExists
	}
	if nil == v {
		return 0, ValueIsNil
	}
	return AsUint32(v)
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
		return 0, NotExists
	}
	if nil == v {
		return 0, ValueIsNil
	}
	return AsUint64(v)
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
		return "", NotExists
	}
	if nil == v {
		return "", ValueIsNil
	}
	return AsString(v)
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
		return nil, NotExists
	}

	if nil == v {
		return nil, ValueIsNil
	}

	res, ok := v.(map[string]interface{})
	if !ok {
		return nil, IsNotMap
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
		return nil, NotExists
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
		return nil, IsNotMapOrArray
	}
	return results, nil
}

type Matcher interface {
	Match(key string, actual interface{}) bool
}

type IntMatcher int

func (self *IntMatcher) Match(key string, actual interface{}) bool {
	i, e := AsInt(actual)
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

type InterfaceMap map[string]interface{}

func (self InterfaceMap) Contains(key string) bool {
	_, ok := self[key]
	return ok
}

func (self InterfaceMap) Get(key string) interface{} {
	return self[key]
}

func (self InterfaceMap) GetBool(key string, defaultValue bool) bool {
	return GetBool(self, key, defaultValue)
}

func (self InterfaceMap) GetInt(key string, defaultValue int) int {
	return GetInt(self, key, defaultValue)
}

func (self InterfaceMap) GetInt32(key string, defaultValue int32) int32 {
	return GetInt32(self, key, defaultValue)
}

func (self InterfaceMap) GetInt64(key string, defaultValue int64) int64 {
	return GetInt64(self, key, defaultValue)
}

func (self InterfaceMap) GetUint(key string, defaultValue uint) uint {
	return GetUint(self, key, defaultValue)
}

func (self InterfaceMap) GetUint32(key string, defaultValue uint32) uint32 {
	return GetUint32(self, key, defaultValue)
}

func (self InterfaceMap) GetUint64(key string, defaultValue uint64) uint64 {
	return GetUint64(self, key, defaultValue)
}

func (self InterfaceMap) GetString(key, defaultValue string) string {
	return GetString(self, key, defaultValue)
}

func (self InterfaceMap) GetArray(key string) []interface{} {
	return GetArray(self, key)
}

func (self InterfaceMap) GetObject(key string) map[string]interface{} {
	return GetObject(self, key)
}

func (self InterfaceMap) GetObjects(key string) []map[string]interface{} {
	return GetObjects(self, key)
}

func (self InterfaceMap) ToMap() map[string]interface{} {
	return map[string]interface{}(self)
}

func (self InterfaceMap) TryGet(key string) (interface{}, bool) {
	o, ok := self[key]
	return o, ok
}

func (self InterfaceMap) TryGetBool(key string) (bool, error) {
	return TryGetBool(self, key)
}

func (self InterfaceMap) TryGetInt(key string) (int, error) {
	return TryGetInt(self, key)
}

func (self InterfaceMap) TryGetInt32(key string) (int32, error) {
	return TryGetInt32(self, key)
}

func (self InterfaceMap) TryGetInt64(key string) (int64, error) {
	return TryGetInt64(self, key)
}

func (self InterfaceMap) TryGetUint(key string) (uint, error) {
	return TryGetUint(self, key)
}

func (self InterfaceMap) TryGetUint32(key string) (uint32, error) {
	return TryGetUint32(self, key)
}

func (self InterfaceMap) TryGetUint64(key string) (uint64, error) {
	return TryGetUint64(self, key)
}

func (self InterfaceMap) TryGetString(key string) (string, error) {
	return TryGetString(self, key)
}

func (self InterfaceMap) TryGetObject(key string) (map[string]interface{}, error) {
	return TryGetObject(self, key)
}

func (self InterfaceMap) TryGetObjects(key string) ([]map[string]interface{}, error) {
	return TryGetObjects(self, key)
}

type StringMap map[string]string

func (self StringMap) Contains(key string) bool {
	_, ok := self[key]
	return ok
}

func (self StringMap) Get(key string) interface{} {
	return self[key]
}

func (self StringMap) GetBool(key string, defaultValue bool) bool {
	s, ok := self[key]
	if !ok {
		return defaultValue
	}
	switch s {
	case "true", "True", "TRUE", "1":
		return true
	case "false", "False", "FALSE", "0":
		return false
	default:
		return defaultValue
	}
}

func (self StringMap) GetInt(key string, defaultValue int) int {
	s, ok := self[key]
	if !ok {
		return defaultValue
	}
	i, e := strconv.ParseInt(s, 10, 0)
	if nil != e {
		return defaultValue
	}
	return int(i)
}

func (self StringMap) GetInt32(key string, defaultValue int32) int32 {
	s, ok := self[key]
	if !ok {
		return defaultValue
	}
	i, e := strconv.ParseInt(s, 10, 32)
	if nil != e {
		return defaultValue
	}
	return int32(i)
}

func (self StringMap) GetInt64(key string, defaultValue int64) int64 {
	s, ok := self[key]
	if !ok {
		return defaultValue
	}
	i, e := strconv.ParseInt(s, 10, 64)
	if nil != e {
		return defaultValue
	}
	return i
}

func (self StringMap) GetUint(key string, defaultValue uint) uint {
	s, ok := self[key]
	if !ok {
		return defaultValue
	}
	i, e := strconv.ParseUint(s, 10, 0)
	if nil != e {
		return defaultValue
	}
	return uint(i)
}

func (self StringMap) GetUint32(key string, defaultValue uint32) uint32 {
	s, ok := self[key]
	if !ok {
		return defaultValue
	}
	i, e := strconv.ParseUint(s, 10, 32)
	if nil != e {
		return defaultValue
	}
	return uint32(i)
}

func (self StringMap) GetUint64(key string, defaultValue uint64) uint64 {
	s, ok := self[key]
	if !ok {
		return defaultValue
	}
	i, e := strconv.ParseUint(s, 10, 64)
	if nil != e {
		return defaultValue
	}
	return uint64(i)
}

func (self StringMap) GetString(key, defaultValue string) string {
	s, ok := self[key]
	if !ok {
		return defaultValue
	}
	return s
}

func (self StringMap) GetArray(key string) []interface{} {
	return nil
}

func (self StringMap) GetObject(key string) map[string]interface{} {
	return nil
}

func (self StringMap) GetObjects(key string) []map[string]interface{} {
	return nil
}

func (self StringMap) ToMap() map[string]interface{} {
	return nil
}

func (self StringMap) TryGet(key string) (interface{}, bool) {
	o, ok := self[key]
	return o, ok
}

func (self StringMap) TryGetBool(key string) (bool, error) {
	s, ok := self[key]
	if !ok {
		return false, NotExists
	}
	switch s {
	case "true", "True", "TRUE", "1":
		return true, nil
	case "false", "False", "FALSE", "0":
		return false, nil
	default:
		return false, IsNotBool
	}
}

func (self StringMap) TryGetInt(key string) (int, error) {
	s, ok := self[key]
	if !ok {
		return 0, NotExists
	}
	i, e := strconv.ParseInt(s, 10, 0)
	if nil != e {
		return 0, e
	}
	return int(i), e
}

func (self StringMap) TryGetInt32(key string) (int32, error) {
	s, ok := self[key]
	if !ok {
		return 0, NotExists
	}
	i, e := strconv.ParseInt(s, 10, 32)
	if nil != e {
		return 0, e
	}
	return int32(i), e
}

func (self StringMap) TryGetInt64(key string) (int64, error) {
	s, ok := self[key]
	if !ok {
		return 0, NotExists
	}
	i, e := strconv.ParseInt(s, 10, 64)
	if nil != e {
		return 0, e
	}
	return int64(i), e
}

func (self StringMap) TryGetUint(key string) (uint, error) {
	s, ok := self[key]
	if !ok {
		return 0, NotExists
	}
	i, e := strconv.ParseUint(s, 10, 0)
	if nil != e {
		return 0, e
	}
	return uint(i), e
}

func (self StringMap) TryGetUint32(key string) (uint32, error) {
	s, ok := self[key]
	if !ok {
		return 0, NotExists
	}
	i, e := strconv.ParseUint(s, 10, 32)
	if nil != e {
		return 0, e
	}
	return uint32(i), e
}

func (self StringMap) TryGetUint64(key string) (uint64, error) {
	s, ok := self[key]
	if !ok {
		return 0, NotExists
	}
	return strconv.ParseUint(s, 10, 64)
}

func (self StringMap) TryGetString(key string) (string, error) {
	s, ok := self[key]
	if !ok {
		return "", NotExists
	}
	return s, nil
}

func (self StringMap) TryGetObject(key string) (map[string]interface{}, error) {
	return nil, errors.New("it is a map[string]string, not support TryGetObject")
}

func (self StringMap) TryGetObjects(key string) ([]map[string]interface{}, error) {
	return nil, errors.New("it is a map[string]string, not support TryGetObjects")
}

type ProxyMap struct {
	values Map
	proxy  Map
}

func Proxy(values, proxy Map) Map {
	if nil == values {
		return proxy
	}
	if nil == proxy {
		return values
	}
	return &ProxyMap{values: values, proxy: proxy}
}

func (self ProxyMap) Contains(key string) bool {
	ok := self.values.Contains(key)
	if ok {
		return ok
	}
	return self.proxy.Contains(key)
}

func (self ProxyMap) Get(key string) interface{} {
	v, ok := self.values.TryGet(key)
	if ok {
		return v
	}
	return self.proxy.Get(key)
}

func (self ProxyMap) GetBool(key string, defaultValue bool) bool {
	v := self.Get(key)
	if nil == v {
		return defaultValue
	}
	b, e := AsBool(v)
	if nil != e {
		return defaultValue
	}
	return b
}

func (self ProxyMap) GetInt(key string, defaultValue int) int {
	v := self.Get(key)
	if nil == v {
		return defaultValue
	}
	i, e := AsInt(v)
	if nil != e {
		return defaultValue
	}
	return i
}

func (self ProxyMap) GetInt32(key string, defaultValue int32) int32 {
	v := self.Get(key)
	if nil == v {
		return defaultValue
	}
	i, e := AsInt32(v)
	if nil != e {
		return defaultValue
	}
	return i
}

func (self ProxyMap) GetInt64(key string, defaultValue int64) int64 {
	v := self.Get(key)
	if nil == v {
		return defaultValue
	}
	i, e := AsInt64(v)
	if nil != e {
		return defaultValue
	}
	return i
}

func (self ProxyMap) GetUint(key string, defaultValue uint) uint {
	v := self.Get(key)
	if nil == v {
		return defaultValue
	}
	i, e := AsUint(v)
	if nil != e {
		return defaultValue
	}
	return i
}

func (self ProxyMap) GetUint32(key string, defaultValue uint32) uint32 {
	v := self.Get(key)
	if nil == v {
		return defaultValue
	}
	i, e := AsUint32(v)
	if nil != e {
		return defaultValue
	}
	return i
}

func (self ProxyMap) GetUint64(key string, defaultValue uint64) uint64 {
	v := self.Get(key)
	if nil == v {
		return defaultValue
	}
	i, e := AsUint64(v)
	if nil != e {
		return defaultValue
	}
	return i
}

func (self ProxyMap) GetString(key, defaultValue string) string {
	v := self.Get(key)
	if nil == v {
		return defaultValue
	}
	s, e := AsString(v)
	if nil != e {
		return defaultValue
	}
	return s
}

func (self ProxyMap) GetArray(key string) []interface{} {
	v := self.Get(key)
	if nil == v {
		return nil
	}
	a, e := AsArray(v)
	if nil != e {
		return nil
	}
	return a
}

func (self ProxyMap) GetObject(key string) map[string]interface{} {
	v := self.Get(key)
	if nil == v {
		return nil
	}

	res, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	return res
}

func (self ProxyMap) GetObjects(key string) []map[string]interface{} {
	v := self.values.GetObjects(key)
	if nil == v {
		return nil
	}
	return self.proxy.GetObjects(key)
}

func (self ProxyMap) ToMap() map[string]interface{} {
	return self.values.ToMap()
}

func (self ProxyMap) TryGet(key string) (interface{}, bool) {
	o, ok := self.values.TryGet(key)
	if ok {
		return o, ok
	}
	return self.proxy.TryGet(key)
}

func (self ProxyMap) TryGetBool(key string) (bool, error) {
	b, e := self.values.TryGetBool(key)
	if nil == e {
		return b, e
	}
	return self.proxy.TryGetBool(key)
}

func (self ProxyMap) TryGetInt(key string) (int, error) {
	i, e := self.values.TryGetInt(key)
	if nil == e {
		return i, e
	}
	return self.proxy.TryGetInt(key)
}

func (self ProxyMap) TryGetInt32(key string) (int32, error) {
	i, e := self.values.TryGetInt32(key)
	if nil == e {
		return i, e
	}
	return self.proxy.TryGetInt32(key)
}

func (self ProxyMap) TryGetInt64(key string) (int64, error) {
	i, e := self.values.TryGetInt64(key)
	if nil == e {
		return i, e
	}
	return self.proxy.TryGetInt64(key)
}

func (self ProxyMap) TryGetUint(key string) (uint, error) {
	i, e := self.values.TryGetUint(key)
	if nil == e {
		return i, e
	}
	return self.proxy.TryGetUint(key)
}

func (self ProxyMap) TryGetUint32(key string) (uint32, error) {
	i, e := self.values.TryGetUint32(key)
	if nil == e {
		return i, e
	}
	return self.proxy.TryGetUint32(key)
}

func (self ProxyMap) TryGetUint64(key string) (uint64, error) {
	i, e := self.values.TryGetUint64(key)
	if nil == e {
		return i, e
	}
	return self.proxy.TryGetUint64(key)
}

func (self ProxyMap) TryGetString(key string) (string, error) {
	s, e := self.values.TryGetString(key)
	if nil == e {
		return s, e
	}
	return self.proxy.TryGetString(key)
}
