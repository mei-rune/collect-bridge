package metrics

import (
	"commons"
	"ds"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func split(exp string) (string, string) {
	idx := strings.IndexRune(exp, '.')
	if -1 == idx {
		return "", exp
	}
	return exp[:idx], exp[idx+1:]
}

type context struct {
	params       map[string]string
	managed_type string
	managed_id   string
	mo           commons.Map

	alias map[string]string
	local map[string]commons.Map
	pry   *proxy
}

func (self *context) Set(key string, value interface{}) {
	if s, ok := value.(string); ok {
		self.params[key] = s
	} else {
		self.params[key] = fmt.Sprint(value)
	}
}

func (self *context) cache(t string) (commons.Map, error) {
	if m, ok := self.local[t]; ok {
		return m, nil
	}

	if tn, ok := self.alias[t]; ok {
		t = tn

		if m, ok := self.local[tn]; ok {
			return m, nil
		}
	}

	n := ds.GetChildrenForm(self.mo.GetWithDefault("$attributes", nil),
		map[string]commons.Matcher{"type": commons.EqualString(t)})
	if nil == n || 0 == len(n) {
		return nil, errors.New("table '" + t + "' is not exists.")
	}

	self.local[t] = commons.InterfaceMap(n[0])
	return commons.InterfaceMap(n[0]), nil
}

func (self *context) Contains(key string) bool {
	if _, ok := self.params[key]; ok {
		return ok
	}

	switch key[0] {
	case '@':
		return self.mo.Contains(key[1:])
	}

	t, field := split(key)
	if 0 == len(t) {
		return false
	}
	res, e := self.cache(t)
	if nil != e {
		return false
	}
	return res.Contains(field)
}

func (self *context) GetWithDefault(key string, defaultValue interface{}) interface{} {
	if s, ok := self.params[key]; ok {
		return s
	}

	switch key[0] {
	case '@':
		return self.mo.GetWithDefault(key[1:], defaultValue)
	case '!':
		return self.pry.GetWithDefault(key[1:], self, defaultValue)
	}

	t, field := split(key)
	if 0 == len(t) {
		return defaultValue
	}
	res, e := self.cache(t)
	if nil != e {
		return defaultValue
	}
	return res.GetWithDefault(field, defaultValue)
}

func (self *context) GetBoolWithDefault(key string, defaultValue bool) bool {
	if s, ok := self.params[key]; ok {
		b, e := commons.AsBool(s)
		if nil != e {
			return defaultValue
		}
		return b
	}

	switch key[0] {
	case '@':
		return self.mo.GetBoolWithDefault(key[1:], defaultValue)
	case '!':
		return self.pry.GetBoolWithDefault(key[1:], self, defaultValue)
	}

	t, field := split(key)
	if 0 == len(t) {
		return defaultValue
	}
	res, e := self.cache(t)
	if nil != e {
		return defaultValue
	}
	return res.GetBoolWithDefault(field, defaultValue)
}

func (self *context) GetIntWithDefault(key string, defaultValue int) int {
	if s, ok := self.params[key]; ok {
		i, e := strconv.ParseInt(s, 10, 0)
		if nil != e {
			return defaultValue
		}
		return int(i)
	}

	switch key[0] {
	case '@':
		return self.mo.GetIntWithDefault(key[1:], defaultValue)
	case '!':
		return self.pry.GetIntWithDefault(key[1:], self, defaultValue)
	}

	t, field := split(key)
	if 0 == len(t) {
		return defaultValue
	}
	res, e := self.cache(t)
	if nil != e {
		return defaultValue
	}
	return res.GetIntWithDefault(field, defaultValue)
}

func (self *context) GetInt32WithDefault(key string, defaultValue int32) int32 {
	if s, ok := self.params[key]; ok {
		i, e := strconv.ParseInt(s, 10, 32)
		if nil != e {
			return defaultValue
		}
		return int32(i)
	}

	switch key[0] {
	case '@':
		return self.mo.GetInt32WithDefault(key[1:], defaultValue)
	case '!':
		return self.pry.GetInt32WithDefault(key[1:], self, defaultValue)
	}

	t, field := split(key)
	if 0 == len(t) {
		return defaultValue
	}
	res, e := self.cache(t)
	if nil != e {
		return defaultValue
	}
	return res.GetInt32WithDefault(field, defaultValue)
}

func (self *context) GetInt64WithDefault(key string, defaultValue int64) int64 {
	if s, ok := self.params[key]; ok {
		i, e := strconv.ParseInt(s, 10, 64)
		if nil != e {
			return defaultValue
		}
		return i
	}

	switch key[0] {
	case '@':
		return self.mo.GetInt64WithDefault(key[1:], defaultValue)
	case '!':
		return self.pry.GetInt64WithDefault(key[1:], self, defaultValue)
	}

	t, field := split(key)
	if 0 == len(t) {
		return defaultValue
	}
	res, e := self.cache(t)
	if nil != e {
		return defaultValue
	}
	return res.GetInt64WithDefault(field, defaultValue)
}

func (self *context) GetUintWithDefault(key string, defaultValue uint) uint {
	if s, ok := self.params[key]; ok {
		u, e := strconv.ParseUint(s, 10, 0)
		if nil != e {
			return defaultValue
		}
		return uint(u)
	}

	switch key[0] {
	case '@':
		return self.mo.GetUintWithDefault(key[1:], defaultValue)
	case '!':
		return self.pry.GetUintWithDefault(key[1:], self, defaultValue)
	}

	t, field := split(key)
	if 0 == len(t) {
		return defaultValue
	}
	res, e := self.cache(t)
	if nil != e {
		return defaultValue
	}
	return res.GetUintWithDefault(field, defaultValue)
}

func (self *context) GetUint32WithDefault(key string, defaultValue uint32) uint32 {
	if s, ok := self.params[key]; ok {
		u, e := strconv.ParseUint(s, 10, 32)
		if nil != e {
			return defaultValue
		}
		return uint32(u)
	}

	switch key[0] {
	case '@':
		return self.mo.GetUint32WithDefault(key[1:], defaultValue)
	case '!':
		return self.pry.GetUint32WithDefault(key[1:], self, defaultValue)
	}

	t, field := split(key)
	if 0 == len(t) {
		return defaultValue
	}
	res, e := self.cache(t)
	if nil != e {
		return defaultValue
	}
	return res.GetUint32WithDefault(field, defaultValue)
}

func (self *context) GetUint64WithDefault(key string, defaultValue uint64) uint64 {
	if s, ok := self.params[key]; ok {
		u, e := strconv.ParseUint(s, 10, 64)
		if nil != e {
			return defaultValue
		}
		return u
	}

	switch key[0] {
	case '@':
		return self.mo.GetUint64WithDefault(key[1:], defaultValue)
	case '!':
		return self.pry.GetUint64WithDefault(key[1:], self, defaultValue)
	}

	t, field := split(key)
	if 0 == len(t) {
		return defaultValue
	}
	res, e := self.cache(t)
	if nil != e {
		return defaultValue
	}
	return res.GetUint64WithDefault(field, defaultValue)
}

func (self *context) GetStringWithDefault(key, defaultValue string) string {
	if s, ok := self.params[key]; ok {
		return s
	}

	switch key[0] {
	case '@':
		return self.mo.GetStringWithDefault(key[1:], defaultValue)
	case '!':
		return self.pry.GetStringWithDefault(key[1:], self, defaultValue)
	}

	t, field := split(key)
	if 0 == len(t) {
		return defaultValue
	}
	res, e := self.cache(t)
	if nil != e {
		return defaultValue
	}
	return res.GetStringWithDefault(field, defaultValue)
}

func (self *context) GetArrayWithDefault(key string, defaultValue []interface{}) []interface{} {
	if _, ok := self.params[key]; ok {
		return defaultValue
	}

	switch key[0] {
	case '@':
		return self.mo.GetArrayWithDefault(key[1:], defaultValue)
	case '!':
		return self.pry.GetArrayWithDefault(key[1:], self, defaultValue)
	}

	t, field := split(key)
	if 0 == len(t) {
		return defaultValue
	}
	res, e := self.cache(t)
	if nil != e {
		return defaultValue
	}
	return res.GetArrayWithDefault(field, defaultValue)
}

func (self *context) GetObjectWithDefault(key string, defaultValue map[string]interface{}) map[string]interface{} {
	if _, ok := self.params[key]; ok {
		return defaultValue
	}

	switch key[0] {
	case '@':
		return self.mo.GetObjectWithDefault(key[1:], defaultValue)
	case '!':
		return self.pry.GetObjectWithDefault(key[1:], self, defaultValue)
	}

	t, field := split(key)
	if 0 == len(t) {
		return defaultValue
	}
	res, e := self.cache(t)
	if nil != e {
		return defaultValue
	}
	return res.GetObjectWithDefault(field, defaultValue)
}

func (self *context) GetObjectsWithDefault(key string, defaultValue []map[string]interface{}) []map[string]interface{} {
	if _, ok := self.params[key]; ok {
		return defaultValue
	}

	switch key[0] {
	case '@':
		return self.mo.GetObjectsWithDefault(key[1:], defaultValue)
	case '!':
		return self.pry.GetObjectsWithDefault(key[1:], self, defaultValue)
	}

	t, field := split(key)
	if 0 == len(t) {
		return defaultValue
	}
	res, e := self.cache(t)
	if nil != e {
		return defaultValue
	}
	return res.GetObjectsWithDefault(field, defaultValue)
}

func (self *context) Get(key string) (interface{}, error) {
	if s, ok := self.params[key]; ok {
		return s, nil
	}

	switch key[0] {
	case '@':
		return self.mo.Get(key[1:])
	case '!':
		return self.pry.Get(key[1:], self)
	}

	t, field := split(key)
	if 0 == len(t) {
		return nil, commons.NotExists
	}
	res, e := self.cache(t)
	if nil != e {
		return nil, e
	}
	return res.GetBool(field)
}

func (self *context) GetBool(key string) (bool, error) {
	if s, ok := self.params[key]; ok {
		return commons.AsBool(s)
	}
	switch key[0] {
	case '@':
		return self.mo.GetBool(key[1:])
	case '!':
		return self.pry.GetBool(key[1:], self)
	}

	t, field := split(key)
	if 0 == len(t) {
		return false, commons.NotExists
	}
	res, e := self.cache(t)
	if nil != e {
		return false, e
	}
	return res.GetBool(field)
}

func (self *context) GetInt(key string) (int, error) {
	if s, ok := self.params[key]; ok {
		i, e := strconv.ParseInt(s, 10, 0)
		return int(i), e
	}

	switch key[0] {
	case '@':
		return self.mo.GetInt(key[1:])
	case '!':
		return self.pry.GetInt(key[1:], self)
	}

	t, field := split(key)
	if 0 == len(t) {
		return 0, commons.NotExists
	}
	res, e := self.cache(t)
	if nil != e {
		return 0, e
	}
	return res.GetInt(field)
}

func (self *context) GetInt32(key string) (int32, error) {
	if s, ok := self.params[key]; ok {
		i, e := strconv.ParseInt(s, 10, 32)
		return int32(i), e
	}

	switch key[0] {
	case '@':
		return self.mo.GetInt32(key[1:])
	case '!':
		return self.pry.GetInt32(key[1:], self)
	}

	t, field := split(key)
	if 0 == len(t) {
		return 0, commons.NotExists
	}
	res, e := self.cache(t)
	if nil != e {
		return 0, e
	}
	return res.GetInt32(field)
}

func (self *context) GetInt64(key string) (int64, error) {
	if s, ok := self.params[key]; ok {
		i, e := strconv.ParseInt(s, 10, 64)
		return int64(i), e
	}

	switch key[0] {
	case '@':
		return self.mo.GetInt64(key[1:])
	case '!':
		return self.pry.GetInt64(key[1:], self)
	}

	t, field := split(key)
	if 0 == len(t) {
		return 0, commons.NotExists
	}
	res, e := self.cache(t)
	if nil != e {
		return 0, e
	}
	return res.GetInt64(field)
}

func (self *context) GetUint(key string) (uint, error) {
	if s, ok := self.params[key]; ok {
		u, e := strconv.ParseUint(s, 10, 0)
		return uint(u), e
	}

	switch key[0] {
	case '@':
		return self.mo.GetUint(key[1:])
	case '!':
		return self.pry.GetUint(key[1:], self)
	}

	t, field := split(key)
	if 0 == len(t) {
		return 0, commons.NotExists
	}
	res, e := self.cache(t)
	if nil != e {
		return 0, e
	}
	return res.GetUint(field)
}

func (self *context) GetUint32(key string) (uint32, error) {
	if s, ok := self.params[key]; ok {
		u, e := strconv.ParseUint(s, 10, 32)
		return uint32(u), e
	}

	switch key[0] {
	case '@':
		return self.mo.GetUint32(key[1:])
	case '!':
		return self.pry.GetUint32(key[1:], self)
	}

	t, field := split(key)
	if 0 == len(t) {
		return 0, commons.NotExists
	}
	res, e := self.cache(t)
	if nil != e {
		return 0, e
	}
	return res.GetUint32(field)
}

func (self *context) GetUint64(key string) (uint64, error) {
	if s, ok := self.params[key]; ok {
		u, e := strconv.ParseUint(s, 10, 64)
		return uint64(u), e
	}

	switch key[0] {
	case '@':
		return self.mo.GetUint64(key[1:])
	case '!':
		return self.pry.GetUint64(key[1:], self)
	}

	t, field := split(key)
	if 0 == len(t) {
		return 0, commons.NotExists
	}
	res, e := self.cache(t)
	if nil != e {
		return 0, e
	}
	return res.GetUint64(field)
}

func (self *context) GetString(key string) (string, error) {
	if s, ok := self.params[key]; ok {
		return s, nil
	}

	switch key[0] {
	case '@':
		return self.mo.GetString(key[1:])
	case '!':
		return self.pry.GetString(key[1:], self)
	}

	t, field := split(key)
	if 0 == len(t) {
		return "", commons.NotExists
	}
	res, e := self.cache(t)
	if nil != e {
		return "", e
	}
	return res.GetString(field)
}

func (self *context) GetObject(key string) (map[string]interface{}, error) {
	if _, ok := self.params[key]; ok {
		return nil, commons.IsNotMap
	}

	switch key[0] {
	case '@':
		return self.mo.GetObject(key[1:])
	case '!':
		return self.pry.GetObject(key[1:], self)
	}

	t, field := split(key)
	if 0 == len(t) {
		return nil, commons.NotExists
	}
	res, e := self.cache(t)
	if nil != e {
		return nil, e
	}
	return res.GetObject(field)
}

func (self *context) GetArray(key string) ([]interface{}, error) {
	if _, ok := self.params[key]; ok {
		return nil, commons.IsNotArray
	}

	switch key[0] {
	case '@':
		return self.mo.GetArray(key[1:])
	case '!':
		return self.pry.GetArray(key[1:], self)
	}

	t, field := split(key)
	if 0 == len(t) {
		return nil, commons.NotExists
	}
	res, e := self.cache(t)
	if nil != e {
		return nil, e
	}
	return res.GetArray(field)
}

func (self *context) GetObjects(key string) ([]map[string]interface{}, error) {
	if _, ok := self.params[key]; ok {
		return nil, commons.IsNotArray
	}

	switch key[0] {
	case '@':
		return self.mo.GetObjects(key[1:])
	case '!':
		return self.pry.GetObjects(key[1:], self)
	}

	t, field := split(key)
	if 0 == len(t) {
		return nil, commons.NotExists
	}
	res, e := self.cache(t)
	if nil != e {
		return nil, e
	}
	return res.GetObjects(field)
}
