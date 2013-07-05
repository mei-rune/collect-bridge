package commons

import (
	"fmt"
	"strconv"
	"time"
)

func GetBoolWithDefault(attributes map[string]interface{}, key string, defaultValue bool) bool {
	res, e := GetBool(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func GetBool(attributes map[string]interface{}, key string) (bool, error) {
	v, ok := attributes[key]
	if !ok {
		return false, NotExists
	}
	if nil == v {
		return false, ParameterIsNil
	}
	return AsBool(v)
}

func GetIntWithDefault(attributes map[string]interface{}, key string, defaultValue int) int {
	res, e := GetInt(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func GetInt(attributes map[string]interface{}, key string) (int, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, NotExists
	}
	if nil == v {
		return 0, ParameterIsNil
	}
	return AsInt(v)
}

func GetUintWithDefault(attributes map[string]interface{}, key string, defaultValue uint) uint {
	res, e := GetUint(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func GetUint(attributes map[string]interface{}, key string) (uint, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, NotExists
	}
	if nil == v {
		return 0, ParameterIsNil
	}
	return AsUint(v)
}

func GetFloatWithDefault(attributes map[string]interface{}, key string, defaultValue float64) float64 {
	res, e := GetFloat(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func GetFloat(attributes map[string]interface{}, key string) (float64, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, NotExists
	}
	if nil == v {
		return 0, ParameterIsNil
	}
	return AsFloat64(v)
}

func GetInt32WithDefault(attributes map[string]interface{}, key string, defaultValue int32) int32 {
	res, e := GetInt32(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func GetInt32(attributes map[string]interface{}, key string) (int32, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, NotExists
	}
	if nil == v {
		return 0, ParameterIsNil
	}
	return AsInt32(v)
}

func GetInt64WithDefault(attributes map[string]interface{}, key string, defaultValue int64) int64 {
	res, e := GetInt64(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func GetInt64(attributes map[string]interface{}, key string) (int64, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, NotExists
	}
	if nil == v {
		return 0, ParameterIsNil
	}
	return AsInt64(v)
}

func GetUint32WithDefault(attributes map[string]interface{}, key string, defaultValue uint32) uint32 {
	res, e := GetUint32(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func GetUint32(attributes map[string]interface{}, key string) (uint32, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, NotExists
	}
	if nil == v {
		return 0, ParameterIsNil
	}
	return AsUint32(v)
}

func GetUint64WithDefault(attributes map[string]interface{}, key string, defaultValue uint64) uint64 {
	res, e := GetUint64(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}
func GetUint64(attributes map[string]interface{}, key string) (uint64, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, NotExists
	}
	if nil == v {
		return 0, ParameterIsNil
	}
	return AsUint64(v)
}
func GetStringWithDefault(attributes map[string]interface{}, key string, defaultValue string) string {
	res, e := GetString(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func GetString(attributes map[string]interface{}, key string) (string, error) {
	v, ok := attributes[key]
	if !ok {
		return "", NotExists
	}
	if nil == v {
		return "", ParameterIsNil
	}
	return AsString(v)
}
func GetTimeWithDefault(attributes map[string]interface{}, key string, defaultValue time.Time) time.Time {
	res, e := GetTime(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func GetTime(attributes map[string]interface{}, key string) (time.Time, error) {
	v, ok := attributes[key]
	if !ok {
		return time.Time{}, NotExists
	}
	if nil == v {
		return time.Time{}, ParameterIsNil
	}
	return AsTime(v)
}

func GetArray(attributes map[string]interface{}, key string) ([]interface{}, error) {
	v, ok := attributes[key]
	if !ok {
		return nil, NotExists
	}

	if nil == v {
		return nil, ParameterIsNil
	}

	res, ok := v.([]interface{})
	if !ok {
		return nil, IsNotArray
	}
	return res, nil
}

func GetArrayWithDefault(attributes map[string]interface{}, key string, defaultValue []interface{}) []interface{} {
	v, ok := attributes[key]
	if !ok {
		return defaultValue
	}

	if nil == v {
		return defaultValue
	}

	res, ok := v.([]interface{})
	if !ok {
		return defaultValue
	}
	return res
}

func GetObjectWithDefault(attributes map[string]interface{}, key string, defaultValue map[string]interface{}) map[string]interface{} {
	v, ok := attributes[key]
	if !ok {
		return defaultValue
	}

	if nil == v {
		return defaultValue
	}

	res, ok := v.(map[string]interface{})
	if !ok {
		return defaultValue
	}
	return res
}

func GetObject(attributes map[string]interface{}, key string) (map[string]interface{}, error) {
	v, ok := attributes[key]
	if !ok {
		return nil, NotExists
	}

	if nil == v {
		return nil, ParameterIsNil
	}

	res, ok := v.(map[string]interface{})
	if !ok {
		return nil, IsNotMap
	}
	return res, nil
}

func GetObjectsWithDefault(attributes map[string]interface{}, key string, defaultValue []map[string]interface{}) []map[string]interface{} {
	v, ok := attributes[key]
	if !ok {
		return defaultValue
	}

	if nil == v {
		return defaultValue
	}

	results, e := AsObjects(v)
	if nil != e {
		return defaultValue
	}
	return results
}

func GetObjects(attributes map[string]interface{}, key string) ([]map[string]interface{}, error) {
	v, ok := attributes[key]
	if !ok {
		return nil, NotExists
	}

	if nil == v {
		return nil, ParameterIsNil
	}

	return AsObjects(v)
}

type Matcher interface {
	Match(actual interface{}) bool
}

type IntMatcher int

func (self *IntMatcher) Match(actual interface{}) bool {
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

type StringMatcher string

func (self *StringMatcher) Match(actual interface{}) bool {
	s, e := AsString(actual)
	if nil != e {
		return false
	}
	return s == string(*self)
}

func EqualString(v string) Matcher {
	m := StringMatcher(v)
	return &m
}

func IsMatch(instance map[string]interface{}, matchers map[string]Matcher) bool {
	for n, m := range matchers {
		value, ok := instance[n]
		if !ok {
			return false
		}

		if !m.Match(value) {
			return false
		}
	}
	return true
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
				if !matcher.Match(value) {
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

func (self InterfaceMap) CopyTo(copy map[string]interface{}) {
	for k, v := range self {
		copy[k] = v
	}
}

func (self InterfaceMap) Set(key string, value interface{}) {
	self[key] = value
}

func (self InterfaceMap) Contains(key string) bool {
	_, ok := self[key]
	return ok
}

func (self InterfaceMap) Fetch(key string) (interface{}, bool) {
	v, ok := self[key]
	return v, ok
}

func (self InterfaceMap) GetWithDefault(key string, defaultValue interface{}) interface{} {
	if o, ok := self[key]; ok {
		return o
	}
	return defaultValue
}

func (self InterfaceMap) GetBoolWithDefault(key string, defaultValue bool) bool {
	return GetBoolWithDefault(self, key, defaultValue)
}

func (self InterfaceMap) GetIntWithDefault(key string, defaultValue int) int {
	return GetIntWithDefault(self, key, defaultValue)
}

func (self InterfaceMap) GetInt32WithDefault(key string, defaultValue int32) int32 {
	return GetInt32WithDefault(self, key, defaultValue)
}

func (self InterfaceMap) GetInt64WithDefault(key string, defaultValue int64) int64 {
	return GetInt64WithDefault(self, key, defaultValue)
}

func (self InterfaceMap) GetUintWithDefault(key string, defaultValue uint) uint {
	return GetUintWithDefault(self, key, defaultValue)
}

func (self InterfaceMap) GetUint32WithDefault(key string, defaultValue uint32) uint32 {
	return GetUint32WithDefault(self, key, defaultValue)
}

func (self InterfaceMap) GetUint64WithDefault(key string, defaultValue uint64) uint64 {
	return GetUint64WithDefault(self, key, defaultValue)
}

func (self InterfaceMap) GetStringWithDefault(key, defaultValue string) string {
	return GetStringWithDefault(self, key, defaultValue)
}

func (self InterfaceMap) GetArrayWithDefault(key string, defaultValue []interface{}) []interface{} {
	return GetArrayWithDefault(self, key, defaultValue)
}

func (self InterfaceMap) GetObjectWithDefault(key string, defaultValue map[string]interface{}) map[string]interface{} {
	return GetObjectWithDefault(self, key, defaultValue)
}

func (self InterfaceMap) GetObjectsWithDefault(key string, defaultValue []map[string]interface{}) []map[string]interface{} {
	return GetObjectsWithDefault(self, key, defaultValue)
}

func (self InterfaceMap) Get(key string) (interface{}, error) {
	if v, ok := self[key]; ok {
		return v, nil
	}
	return nil, NotExists
}

func (self InterfaceMap) GetBool(key string) (bool, error) {
	return GetBool(self, key)
}

func (self InterfaceMap) GetInt(key string) (int, error) {
	return GetInt(self, key)
}

func (self InterfaceMap) GetInt32(key string) (int32, error) {
	return GetInt32(self, key)
}

func (self InterfaceMap) GetInt64(key string) (int64, error) {
	return GetInt64(self, key)
}

func (self InterfaceMap) GetUint(key string) (uint, error) {
	return GetUint(self, key)
}

func (self InterfaceMap) GetUint32(key string) (uint32, error) {
	return GetUint32(self, key)
}

func (self InterfaceMap) GetUint64(key string) (uint64, error) {
	return GetUint64(self, key)
}

func (self InterfaceMap) GetString(key string) (string, error) {
	return GetString(self, key)
}

func (self InterfaceMap) GetObject(key string) (map[string]interface{}, error) {
	return GetObject(self, key)
}

func (self InterfaceMap) GetArray(key string) ([]interface{}, error) {
	return GetArray(self, key)
}

func (self InterfaceMap) GetObjects(key string) ([]map[string]interface{}, error) {
	return GetObjects(self, key)
}

type StringMap map[string]string

func (self StringMap) CopyTo(copy map[string]interface{}) {
	for k, v := range self {
		copy[k] = v
	}
}

func (self StringMap) Set(key string, value interface{}) {
	if s, ok := value.(string); ok {
		self[key] = s
	} else {
		self[key] = fmt.Sprint(value)
	}
}

func (self StringMap) Contains(key string) bool {
	_, ok := self[key]
	return ok
}

func (self StringMap) Fetch(key string) (interface{}, bool) {
	o, ok := self[key]
	return o, ok
}

func (self StringMap) GetWithDefault(key string, defaultValue interface{}) interface{} {
	if o, ok := self[key]; ok {
		return o
	}
	return defaultValue
}

func (self StringMap) GetBoolWithDefault(key string, defaultValue bool) bool {
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

func (self StringMap) GetIntWithDefault(key string, defaultValue int) int {
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

func (self StringMap) GetInt32WithDefault(key string, defaultValue int32) int32 {
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

func (self StringMap) GetInt64WithDefault(key string, defaultValue int64) int64 {
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

func (self StringMap) GetUintWithDefault(key string, defaultValue uint) uint {
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

func (self StringMap) GetUint32WithDefault(key string, defaultValue uint32) uint32 {
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

func (self StringMap) GetUint64WithDefault(key string, defaultValue uint64) uint64 {
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

func (self StringMap) GetStringWithDefault(key, defaultValue string) string {
	s, ok := self[key]
	if !ok {
		return defaultValue
	}
	return s
}

func (self StringMap) GetArrayWithDefault(key string, defaultValue []interface{}) []interface{} {
	return defaultValue
}

func (self StringMap) GetObjectWithDefault(key string, defaultValue map[string]interface{}) map[string]interface{} {
	return defaultValue
}

func (self StringMap) GetObjectsWithDefault(key string, defaultValue []map[string]interface{}) []map[string]interface{} {
	return defaultValue
}

func (self StringMap) Get(key string) (interface{}, error) {
	s, ok := self[key]
	if !ok {
		return nil, NotExists
	}
	return s, nil
}

func (self StringMap) GetBool(key string) (bool, error) {
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

func (self StringMap) GetInt(key string) (int, error) {
	s, ok := self[key]
	if !ok {
		return 0, NotExists
	}
	i, e := strconv.ParseInt(s, 10, 0)
	if nil != e {
		return 0, typeError(e.Error())
	}
	return int(i), nil
}

func (self StringMap) GetInt32(key string) (int32, error) {
	s, ok := self[key]
	if !ok {
		return 0, NotExists
	}
	i, e := strconv.ParseInt(s, 10, 32)
	if nil != e {
		return 0, typeError(e.Error())
	}
	return int32(i), nil
}

func (self StringMap) GetInt64(key string) (int64, error) {
	s, ok := self[key]
	if !ok {
		return 0, NotExists
	}
	i, e := strconv.ParseInt(s, 10, 64)
	if nil != e {
		return 0, typeError(e.Error())
	}
	return int64(i), nil
}

func (self StringMap) GetUint(key string) (uint, error) {
	s, ok := self[key]
	if !ok {
		return 0, NotExists
	}
	i, e := strconv.ParseUint(s, 10, 0)
	if nil != e {
		return 0, typeError(e.Error())
	}
	return uint(i), nil
}

func (self StringMap) GetUint32(key string) (uint32, error) {
	s, ok := self[key]
	if !ok {
		return 0, NotExists
	}
	i, e := strconv.ParseUint(s, 10, 32)
	if nil != e {
		return 0, typeError(e.Error())
	}
	return uint32(i), nil
}

func (self StringMap) GetUint64(key string) (uint64, error) {
	s, ok := self[key]
	if !ok {
		return 0, NotExists
	}

	u64, e := strconv.ParseUint(s, 10, 64)
	if nil != e {
		return 0, typeError(e.Error())
	}
	return u64, nil

}

func (self StringMap) GetString(key string) (string, error) {
	s, ok := self[key]
	if !ok {
		return "", NotExists
	}
	return s, nil
}

func (self StringMap) GetObject(key string) (map[string]interface{}, error) {
	return nil, typeError("it is a map[string]string, not support GetObject")
}

func (self StringMap) GetArray(key string) ([]interface{}, error) {
	return nil, typeError("it is a map[string]string, not support GetArray")
}

func (self StringMap) GetObjects(key string) ([]map[string]interface{}, error) {
	return nil, typeError("it is a map[string]string, not support GetObjects")
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

func (self ProxyMap) CopyTo(copy map[string]interface{}) {
	self.values.CopyTo(copy)
	self.proxy.CopyTo(copy)
}

func (self ProxyMap) Set(key string, value interface{}) {
	self.values.Set(key, fmt.Sprint(value))
}

func (self ProxyMap) Contains(key string) bool {
	ok := self.values.Contains(key)
	if ok {
		return ok
	}
	return self.proxy.Contains(key)
}

func (self ProxyMap) GetWithDefault(key string, defaultValue interface{}) interface{} {
	if b, e := self.values.Get(key); nil == e {
		return b
	}
	return self.proxy.GetWithDefault(key, defaultValue)
}

func (self ProxyMap) GetBoolWithDefault(key string, defaultValue bool) bool {
	if b, e := self.values.GetBool(key); nil == e {
		return b
	}
	return self.proxy.GetBoolWithDefault(key, defaultValue)
}

func (self ProxyMap) GetIntWithDefault(key string, defaultValue int) int {
	if b, e := self.values.GetInt(key); nil == e {
		return b
	}
	return self.proxy.GetIntWithDefault(key, defaultValue)
}

func (self ProxyMap) GetInt32WithDefault(key string, defaultValue int32) int32 {
	if b, e := self.values.GetInt32(key); nil == e {
		return b
	}
	return self.proxy.GetInt32WithDefault(key, defaultValue)
}

func (self ProxyMap) GetInt64WithDefault(key string, defaultValue int64) int64 {
	if b, e := self.values.GetInt64(key); nil == e {
		return b
	}
	return self.proxy.GetInt64WithDefault(key, defaultValue)
}

func (self ProxyMap) GetUintWithDefault(key string, defaultValue uint) uint {
	if b, e := self.values.GetUint(key); nil == e {
		return b
	}
	return self.proxy.GetUintWithDefault(key, defaultValue)
}

func (self ProxyMap) GetUint32WithDefault(key string, defaultValue uint32) uint32 {
	if b, e := self.values.GetUint32(key); nil == e {
		return b
	}
	return self.proxy.GetUint32WithDefault(key, defaultValue)
}

func (self ProxyMap) GetUint64WithDefault(key string, defaultValue uint64) uint64 {
	if b, e := self.values.GetUint64(key); nil == e {
		return b
	}
	return self.proxy.GetUint64WithDefault(key, defaultValue)
}

func (self ProxyMap) GetStringWithDefault(key, defaultValue string) string {
	if b, e := self.values.GetString(key); nil == e {
		return b
	}
	return self.proxy.GetStringWithDefault(key, defaultValue)
}

func (self ProxyMap) GetArrayWithDefault(key string, defaultValue []interface{}) []interface{} {
	if b, e := self.values.GetArray(key); nil == e {
		return b
	}
	return self.proxy.GetArrayWithDefault(key, defaultValue)
}

func (self ProxyMap) GetObjectWithDefault(key string, defaultValue map[string]interface{}) map[string]interface{} {
	if b, e := self.values.GetObject(key); nil == e {
		return b
	}
	return self.proxy.GetObjectWithDefault(key, defaultValue)
}

func (self ProxyMap) GetObjectsWithDefault(key string, defaultValue []map[string]interface{}) []map[string]interface{} {
	if b, e := self.values.GetObjects(key); nil == e {
		return b
	}
	return self.proxy.GetObjectsWithDefault(key, defaultValue)
}

func (self ProxyMap) Get(key string) (interface{}, error) {
	if r, e := self.values.Get(key); nil == e {
		return r, nil
	}
	return self.proxy.Get(key)
}

func (self ProxyMap) GetBool(key string) (bool, error) {
	b, e := self.values.GetBool(key)
	if nil == e {
		return b, e
	}
	return self.proxy.GetBool(key)
}

func (self ProxyMap) GetInt(key string) (int, error) {
	i, e := self.values.GetInt(key)
	if nil == e {
		return i, e
	}
	return self.proxy.GetInt(key)
}

func (self ProxyMap) GetInt32(key string) (int32, error) {
	i, e := self.values.GetInt32(key)
	if nil == e {
		return i, e
	}
	return self.proxy.GetInt32(key)
}

func (self ProxyMap) GetInt64(key string) (int64, error) {
	i, e := self.values.GetInt64(key)
	if nil == e {
		return i, e
	}
	return self.proxy.GetInt64(key)
}

func (self ProxyMap) GetUint(key string) (uint, error) {
	i, e := self.values.GetUint(key)
	if nil == e {
		return i, e
	}
	return self.proxy.GetUint(key)
}

func (self ProxyMap) GetUint32(key string) (uint32, error) {
	i, e := self.values.GetUint32(key)
	if nil == e {
		return i, e
	}
	return self.proxy.GetUint32(key)
}

func (self ProxyMap) GetUint64(key string) (uint64, error) {
	i, e := self.values.GetUint64(key)
	if nil == e {
		return i, e
	}
	return self.proxy.GetUint64(key)
}

func (self ProxyMap) GetString(key string) (string, error) {
	s, e := self.values.GetString(key)
	if nil == e {
		return s, e
	}
	return self.proxy.GetString(key)
}

func (self ProxyMap) GetObject(key string) (map[string]interface{}, error) {
	s, e := self.values.GetObject(key)
	if nil == e {
		return s, e
	}
	return self.proxy.GetObject(key)
}

func (self ProxyMap) GetArray(key string) ([]interface{}, error) {
	s, e := self.values.GetArray(key)
	if nil == e {
		return s, e
	}
	return self.proxy.GetArray(key)
}

func (self ProxyMap) GetObjects(key string) ([]map[string]interface{}, error) {
	s, e := self.values.GetObjects(key)
	if nil == e {
		return s, e
	}
	return self.proxy.GetObjects(key)
}
