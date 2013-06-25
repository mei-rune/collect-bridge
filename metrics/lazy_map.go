package metrics

import (
	"commons"
	"ds"
	"strings"
)

type lazyMap struct {
	id     string
	caches *ds.Caches
}

func (self lazyMap) getCache(key string) (*ds.Cache, error) {
	return self.caches.GetCache(key)
}

func (self lazyMap) Contains(key string) bool {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return false
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return false
	}
	if nil == cache {
		return false
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return false
	}

	_, ok := res[key]
	return ok
}

func (self lazyMap) Get(key string) interface{} {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return nil
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return nil
	}
	if nil == cache {
		return nil
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return nil
	}

	return res[key[idx+1:]]
}

func (self lazyMap) GetBool(key string, defaultValue bool) bool {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return defaultValue
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return defaultValue
	}
	if nil == cache {
		return defaultValue
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return defaultValue
	}

	return commons.InterfaceMap(res).GetBool(key[idx+1:], defaultValue)
}

func (self lazyMap) GetInt(key string, defaultValue int) int {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return defaultValue
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return defaultValue
	}
	if nil == cache {
		return defaultValue
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return defaultValue
	}

	return commons.InterfaceMap(res).GetInt(key[idx+1:], defaultValue)
}

func (self lazyMap) GetInt32(key string, defaultValue int32) int32 {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return defaultValue
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return defaultValue
	}
	if nil == cache {
		return defaultValue
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return defaultValue
	}

	return commons.InterfaceMap(res).GetInt32(key[idx+1:], defaultValue)
}

func (self lazyMap) GetInt64(key string, defaultValue int64) int64 {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return defaultValue
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return defaultValue
	}
	if nil == cache {
		return defaultValue
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return defaultValue
	}

	return commons.InterfaceMap(res).GetInt64(key[idx+1:], defaultValue)
}

func (self lazyMap) GetUint(key string, defaultValue uint) uint {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return defaultValue
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return defaultValue
	}
	if nil == cache {
		return defaultValue
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return defaultValue
	}

	return commons.InterfaceMap(res).GetUint(key[idx+1:], defaultValue)
}

func (self lazyMap) GetUint32(key string, defaultValue uint32) uint32 {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return defaultValue
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return defaultValue
	}
	if nil == cache {
		return defaultValue
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return defaultValue
	}

	return commons.InterfaceMap(res).GetUint32(key[idx+1:], defaultValue)
}

func (self lazyMap) GetUint64(key string, defaultValue uint64) uint64 {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return defaultValue
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return defaultValue
	}
	if nil == cache {
		return defaultValue
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return defaultValue
	}

	return commons.InterfaceMap(res).GetUint64(key[idx+1:], defaultValue)
}

func (self lazyMap) GetString(key, defaultValue string) string {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return defaultValue
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return defaultValue
	}
	if nil == cache {
		return defaultValue
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return defaultValue
	}

	return commons.InterfaceMap(res).GetString(key[idx+1:], defaultValue)
}

func (self lazyMap) GetArray(key string) []interface{} {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return nil
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return nil
	}
	if nil == cache {
		return nil
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return nil
	}

	return commons.InterfaceMap(res).GetArray(key[idx+1:])
}

func (self lazyMap) GetObject(key string) map[string]interface{} {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return nil
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return nil
	}
	if nil == cache {
		return nil
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return nil
	}

	return commons.InterfaceMap(res).GetObject(key[idx+1:])
}

func (self lazyMap) GetObjects(key string) []map[string]interface{} {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return nil
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return nil
	}
	if nil == cache {
		return nil
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return nil
	}

	return commons.InterfaceMap(res).GetObjects(key[idx+1:])
}

func (self lazyMap) ToMap() map[string]interface{} {
	return nil
}

func (self lazyMap) TryGet(key string) (interface{}, bool) {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return nil, false
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return nil, false
	}
	if nil == cache {
		return nil, false
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return nil, false
	}
	v, ok := res[key[idx+1:]]
	return v, ok
}

func (self lazyMap) TryGetBool(key string) (bool, error) {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return false, commons.NotExists
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return false, e
	}
	if nil == cache {
		return false, commons.NotExists
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return false, commons.NotExists
	}

	return commons.InterfaceMap(res).TryGetBool(key[idx+1:])
}

func (self lazyMap) TryGetInt(key string) (int, error) {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return 0, commons.NotExists
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return 0, e
	}
	if nil == cache {
		return 0, commons.NotExists
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return 0, commons.NotExists
	}

	return commons.InterfaceMap(res).TryGetInt(key[idx+1:])
}

func (self lazyMap) TryGetInt32(key string) (int32, error) {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return 0, commons.NotExists
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return 0, e
	}
	if nil == cache {
		return 0, commons.NotExists
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return 0, commons.NotExists
	}

	return commons.InterfaceMap(res).TryGetInt32(key[idx+1:])
}

func (self lazyMap) TryGetInt64(key string) (int64, error) {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return 0, commons.NotExists
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return 0, e
	}
	if nil == cache {
		return 0, commons.NotExists
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return 0, commons.NotExists
	}

	return commons.InterfaceMap(res).TryGetInt64(key[idx+1:])
}

func (self lazyMap) TryGetUint(key string) (uint, error) {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return 0, commons.NotExists
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return 0, e
	}
	if nil == cache {
		return 0, commons.NotExists
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return 0, commons.NotExists
	}

	return commons.InterfaceMap(res).TryGetUint(key[idx+1:])
}

func (self lazyMap) TryGetUint32(key string) (uint32, error) {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return 0, commons.NotExists
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return 0, e
	}
	if nil == cache {
		return 0, commons.NotExists
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return 0, commons.NotExists
	}

	return commons.InterfaceMap(res).TryGetUint32(key[idx+1:])
}

func (self lazyMap) TryGetUint64(key string) (uint64, error) {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return 0, commons.NotExists
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return 0, e
	}
	if nil == cache {
		return 0, commons.NotExists
	}
	res, e := cache.Get(self.id)
	if nil != e {
		return 0, commons.NotExists
	}

	return commons.InterfaceMap(res).TryGetUint64(key[idx+1:])
}

func (self lazyMap) TryGetString(key string) (string, error) {
	idx := strings.IndexRune(key, '#')
	if -1 == idx {
		return "", commons.NotExists
	}

	cache, e := self.getCache(key[:idx])
	if nil != e {
		return "", e
	}
	if nil == cache {
		return "", commons.NotExists
	}

	res, e := cache.Get(self.id)
	if nil != e {
		return "", commons.NotExists
	}

	return commons.InterfaceMap(res).TryGetString(key[idx+1:])
}
