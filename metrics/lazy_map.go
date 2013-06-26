package metrics

import (
	"commons"
	"ds"
	"errors"
	"strings"
)

func split(exp string) (string, string) {
	idx := strings.IndexRune(exp, '.')
	if -1 == idx {
		return "", exp
	}
	return exp[:idx], exp[idx+2:]
}

type context struct {
	managed_type string
	managed_id   string
	caches       *ds.Caches

	local      map[string]commons.Map
	snmp       commons.Map
	sys        commons.Map
	proxy      *metric_proxy
	top_params commons.Map
}

func (self context) getCache(key string) (*ds.Cache, error) {
	return self.caches.GetCache(key)
}

func (self context) Set(key string, value interface{}) {
	panic("context is only read.")
}
func (self context) cache(t string) (commons.Map, error) {
	if m, ok := self.local[t]; ok {
		return m, nil
	}

	cache, e := self.getCache(t)
	if nil != e {
		return nil, e
	}

	if nil == cache {
		return nil, errors.New("table '" + t + "' is not exists.")
	}

	res, e := cache.Get(self.managed_id)
	if nil != e {
		return nil, e
	}

	self.local[t] = commons.InterfaceMap(res)
	return commons.InterfaceMap(res), nil
}

func (self context) Contains(key string) bool {
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

func (self context) Fetch(key string) (interface{}, bool) {
	t, field := split(key)
	if 0 == len(t) {
		return nil, false
	}
	res, e := self.cache(t)
	if nil != e {
		return nil, false
	}
	return res.Fetch(field)
}

func (self context) GetWithDefault(key string, defaultValue interface{}) interface{} {
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

func (self context) GetBoolWithDefault(key string, defaultValue bool) bool {
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

func (self context) GetIntWithDefault(key string, defaultValue int) int {
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

func (self context) GetInt32WithDefault(key string, defaultValue int32) int32 {
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

func (self context) GetInt64WithDefault(key string, defaultValue int64) int64 {
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

func (self context) GetUintWithDefault(key string, defaultValue uint) uint {
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

func (self context) GetUint32WithDefault(key string, defaultValue uint32) uint32 {
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

func (self context) GetUint64WithDefault(key string, defaultValue uint64) uint64 {
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

func (self context) GetStringWithDefault(key, defaultValue string) string {
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

func (self context) GetArrayWithDefault(key string, defaultValue []interface{}) []interface{} {
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

func (self context) GetObjectWithDefault(key string, defaultValue map[string]interface{}) map[string]interface{} {
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

func (self context) GetObjectsWithDefault(key string, defaultValue []map[string]interface{}) []map[string]interface{} {
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

func (self context) ToMap() map[string]interface{} {
	return nil
}

func (self context) GetBool(key string) (bool, error) {
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

func (self context) GetInt(key string) (int, error) {
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

func (self context) GetInt32(key string) (int32, error) {
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

func (self context) GetInt64(key string) (int64, error) {
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

func (self context) GetUint(key string) (uint, error) {
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

func (self context) GetUint32(key string) (uint32, error) {
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

func (self context) GetUint64(key string) (uint64, error) {
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

func (self context) GetString(key string) (string, error) {
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

func (self context) GetObject(key string) (map[string]interface{}, error) {
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

func (self context) GetArray(key string) ([]interface{}, error) {
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

func (self context) GetObjects(key string) ([]map[string]interface{}, error) {
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
