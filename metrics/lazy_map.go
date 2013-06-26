package metrics

import (
	"commons"
	"ds"
	"strings"
)

func split(exp string) (string, string) {
	idx := strings.IndexRune(exp, '.')
	if -1 == idx {
		return "", exp
	}
	return exp[:idx], exp[idx+2:]
}

type lazyMap struct {
	managed_type string
	managed_id   string
	caches       *ds.Caches

	local      map[string]*ds.Cache
	snmp       commons.Map
	sys        commons.Map
	proxy      *metric_proxy
	top_params commons.Map
}

func (self lazyMap) getCache(key string) (*ds.Cache, error) {
	return self.caches.GetCache(key)
}

func (self lazyMap) Set(key string, value interface{}) {
	panic("lazyMap is only read.")
}

func (self lazyMap) Contains(key string) bool {
	idx := strings.IndexRune(key, '.')
	if -1 == idx {
		return false
	}

	t := key[:idx]
	if "sys" == t {
		t = self.managed_type
	}
	cache, e := self.getCache(t)
	if nil != e {
		return false
	}
	if nil == cache {
		return false
	}
	res, e := cache.Get(self.managed_id)
	if nil != e {
		return false
	}

	_, ok := res[key[idx+1:]]
	return ok
}

func (self lazyMap) cache(t string) (commons.Map, error) {
	if nil == self.local {
		self.local = make(map[string]commons.Map)
	} else if m, ok := self.local[t]; ok {
		return m, nil
	}

	cache, e := self.getCache(t)
	if nil != e {
		return nil, e
	}

	if nil == cache {
		return nil, commons.NewRuntimeError(commons.NotFoundCode, "table '"+t+"' is not exists.")
	}

	res, e := cache.Get(self.managed_id)
	if nil != e {
		return nil, e
	}

	self.local[t] = commons.InterfaceMap(res)
	return commons.InterfaceMap(res)
}

func (self lazyMap) Fetch(key string) (interface{}, bool) {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return nil, false
	}

	t := key[:idx]
	if "sys" == t {
		t = self.managed_type
	}
	cache, e := self.getCache(t)
	if nil != e {
		return nil, false
	}
	if nil == cache {
		return nil, false
	}
	res, e := cache.Get(self.managed_id)
	if nil != e {
		return nil, false
	}
	v, ok := res[key[idx+1:]]
	return v, ok
}

func (self lazyMap) GetWithDefault(key string, defaultValue interface{}) interface{} {
	t, field := split(key)
	switch t {
	case "sys":
		return self.sys().GetWithDefault(field, defaultValue)
	case "snmp":
		return self.snmp().GetWithDefault(field, defaultValue)
	}
	return self.cache(t).GetWithDefault(field, defaultValue)
}

func (self lazyMap) GetBoolWithDefault(key string, defaultValue bool) bool {
	v, ok := self.Fetch(key)
	if !ok {
		return defaultValue
	}
	b, e := commons.AsBool(v)
	if nil != e {
		return defaultValue
	}
	return b
}

func (self lazyMap) GetIntWithDefault(key string, defaultValue int) int {
	v, ok := self.Fetch(key)
	if !ok {
		return defaultValue
	}
	i, e := commons.AsInt(v)
	if nil != e {
		return defaultValue
	}
	return i
}

func (self lazyMap) GetInt32WithDefault(key string, defaultValue int32) int32 {
	v, ok := self.Fetch(key)
	if !ok {
		return defaultValue
	}
	i, e := commons.AsInt32(v)
	if nil != e {
		return defaultValue
	}
	return i
}

func (self lazyMap) GetInt64WithDefault(key string, defaultValue int64) int64 {
	v, ok := self.Fetch(key)
	if !ok {
		return defaultValue
	}
	i, e := commons.AsInt64(v)
	if nil != e {
		return defaultValue
	}
	return i
}

func (self lazyMap) GetUintWithDefault(key string, defaultValue uint) uint {
	v, ok := self.Fetch(key)
	if !ok {
		return defaultValue
	}
	u, e := commons.AsUint(v)
	if nil != e {
		return defaultValue
	}
	return u
}

func (self lazyMap) GetUint32WithDefault(key string, defaultValue uint32) uint32 {
	v, ok := self.Fetch(key)
	if !ok {
		return defaultValue
	}
	u, e := commons.AsUint32(v)
	if nil != e {
		return defaultValue
	}
	return u
}

func (self lazyMap) GetUint64WithDefault(key string, defaultValue uint64) uint64 {
	v, ok := self.Fetch(key)
	if !ok {
		return defaultValue
	}
	u, e := commons.AsUint64(v)
	if nil != e {
		return defaultValue
	}
	return u
}

func (self lazyMap) GetStringWithDefault(key, defaultValue string) string {
	v, ok := self.Fetch(key)
	if !ok {
		return defaultValue
	}
	u, e := commons.AsString(v)
	if nil != e {
		return defaultValue
	}
	return u
}

func (self lazyMap) GetArrayWithDefault(key string, defaultValue []interface{}) []interface{} {
	v, ok := self.Fetch(key)
	if !ok {
		return defaultValue
	}
	u, e := commons.AsArray(v)
	if nil != e {
		return defaultValue
	}
	return u
}

func (self lazyMap) GetObjectWithDefault(key string, defaultValue map[string]interface{}) map[string]interface{} {
	v, ok := self.Fetch(key)
	if !ok {
		return defaultValue
	}

	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return defaultValue
}

func (self lazyMap) GetObjectsWithDefault(key string, defaultValue []map[string]interface{}) []map[string]interface{} {
	v, ok := self.Fetch(key)
	if !ok {
		return defaultValue
	}

	o, e := commons.AsObjects(v)
	if nil != e {
		return defaultValue
	}
	return o
}

func (self lazyMap) ToMap() map[string]interface{} {
	return nil
}

func (self lazyMap) GetBool(key string) (bool, commons.RuntimeError) {
	v, ok := self.Fetch(key)
	if !ok {
		return false, commons.NotExists
	}
	return commons.AsBool(v)
}

func (self lazyMap) GetInt(key string) (int, commons.RuntimeError) {
	v, ok := self.Fetch(key)
	if !ok {
		return 0, commons.NotExists
	}
	return commons.AsInt(v)
}

func (self lazyMap) GetInt32(key string) (int32, commons.RuntimeError) {
	v, ok := self.Fetch(key)
	if !ok {
		return 0, commons.NotExists
	}
	return commons.AsInt32(v)
}

func (self lazyMap) GetInt64(key string) (int64, commons.RuntimeError) {
	v, ok := self.Fetch(key)
	if !ok {
		return 0, commons.NotExists
	}
	return commons.AsInt64(v)
}

func (self lazyMap) GetUint(key string) (uint, commons.RuntimeError) {
	v, ok := self.Fetch(key)
	if !ok {
		return 0, commons.NotExists
	}
	return commons.AsUint(v)
}

func (self lazyMap) GetUint32(key string) (uint32, commons.RuntimeError) {
	v, ok := self.Fetch(key)
	if !ok {
		return 0, commons.NotExists
	}
	return commons.AsUint32(v)
}

func (self lazyMap) GetUint64(key string) (uint64, commons.RuntimeError) {
	v, ok := self.Fetch(key)
	if !ok {
		return 0, commons.NotExists
	}
	return commons.AsUint64(v)
}

func (self lazyMap) GetString(key string) (string, commons.RuntimeError) {
	v, ok := self.Fetch(key)
	if !ok {
		return "", commons.NotExists
	}
	return commons.AsString(v)
}

func (self lazyMap) GetObject(key string) (map[string]interface{}, commons.RuntimeError) {
	v, ok := self.Fetch(key)
	if !ok {
		return nil, commons.NotExists
	}
	return commons.AsObject(v)
}

func (self lazyMap) GetArray(key string) ([]interface{}, commons.RuntimeError) {
	v, ok := self.Fetch(key)
	if !ok {
		return nil, commons.NotExists
	}
	return commons.AsArray(v)
}

func (self lazyMap) GetObjects(key string) ([]map[string]interface{}, commons.RuntimeError) {
	v, ok := self.Fetch(key)
	if !ok {
		return nil, commons.NotExists
	}
	return commons.AsObjects(v)
}
