package metrics

import (
	"commons"
	"ds"
	"errors"
	"fmt"
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
	mo           map[string]interface{}

	alias         map[string]string
	local         map[string]map[string]interface{}
	pry           *proxy
	metrics_cache map[string]interface{}
}

func (self *context) CopyTo(copy map[string]interface{}) {
	for k, v := range self.params {
		copy[k] = v
	}
}

func (self *context) Set(key string, value interface{}) {
	switch key[0] {
	case '&', '!':
		if nil == self.metrics_cache {
			self.metrics_cache = make(map[string]interface{})
		}
		self.metrics_cache[key[1:]] = value
	default:
		if s, ok := value.(string); ok {
			self.params[key] = s
		} else {
			self.params[key] = fmt.Sprint(value)
		}
	}
}

func (self *context) cache(t, field string) (interface{}, error) {
	var n []map[string]interface{}
	m, ok := self.local[t]
	if ok {
		goto ok
	}

	if tn, ok := self.alias[t]; ok {
		t = tn

		m, ok = self.local[tn]
		if ok {
			goto ok
		}
	}

	n = ds.GetChildrenForm(self.mo["$attributes"],
		map[string]commons.Matcher{"type": commons.EqualString(t)})
	if nil == n || 0 == len(n) {
		return nil, errors.New("table '" + t + "' is not exists.")
	}

	m = n[0]
	self.local[t] = m
ok:
	if v, ok := m[field]; ok {
		return v, nil
	}
	return nil, commons.NotExists
}

func (self *context) Contains(key string) bool {
	if _, ok := self.params[key]; ok {
		return ok
	}

	switch key[0] {
	case '@':
		_, ok := self.mo[key[1:]]
		return ok
	case '$':
		new_key := key[1:]
		if _, ok := self.params[new_key]; ok {
			return true
		}
		if _, ok := self.params["@"+new_key]; ok {
			return true
		}

		if _, ok := self.mo[key[1:]]; ok {
			return true
		}

		t, field := split(new_key)
		if 0 != len(t) {
			_, e := self.cache(t, field)
			if nil == e {
				return true
			}
		}
		fallthrough
	case '!':
		if nil != self.metrics_cache {
			if _, ok := self.metrics_cache[key[1:]]; ok {
				return true
			}
		}
		return false
	}

	t, field := split(key)
	if 0 == len(t) {
		return false
	}
	_, e := self.cache(t, field)
	if nil != e {
		return false
	}
	return true
}

func (self *context) Get(key string) (interface{}, error) {
	if s, ok := self.params[key]; ok {
		return s, nil
	}

	switch key[0] {
	case '@':
		if v, ok := self.mo[key[1:]]; ok {
			return v, nil
		}
		return nil, commons.NotExists
	case '$':
		new_key := key[1:]
		if s, ok := self.params[new_key]; ok {
			return s, nil
		}
		if s, ok := self.params["@"+new_key]; ok {
			return s, nil
		}

		if v, ok := self.mo[key[1:]]; ok {
			return v, nil
		}

		t, field := split(new_key)
		if 0 != len(t) {
			v, e := self.cache(t, field)
			if nil == e {
				return v, nil
			}
		}

		if nil != self.metrics_cache {
			if v, ok := self.metrics_cache[new_key]; ok {
				return v, nil
			}
		}

		if nil == self.pry {
			return nil, commons.NotExists
		}

		v, e := self.pry.Get(new_key, self)
		if nil == e {
			if nil == self.metrics_cache {
				self.metrics_cache = make(map[string]interface{})
			}
			self.metrics_cache[new_key] = v
		}
		return v, e
	case '&':
		new_key := key[1:]
		if nil != self.metrics_cache {
			if v, ok := self.metrics_cache[new_key]; ok {
				return v, nil
			}
		}

		if nil == self.pry {
			return nil, commons.NotExists
		}

		v, e := self.pry.Get(new_key, self)
		if nil == e {
			if nil == self.metrics_cache {
				self.metrics_cache = make(map[string]interface{})
			}
			self.metrics_cache[new_key] = v
		}
		return v, nil
	case '!':
		if nil == self.pry {
			return nil, commons.NotExists
		}

		new_key := key[1:]
		v, e := self.pry.Get(new_key, self)
		if nil == e {
			if nil == self.metrics_cache {
				self.metrics_cache = make(map[string]interface{})
			}
			self.metrics_cache[new_key] = v
		}
		return v, e
	}

	t, field := split(key)
	if 0 == len(t) {
		return nil, commons.NotExists
	}
	return self.cache(t, field)
}

func (self *context) GetWithDefault(key string, defaultValue interface{}) interface{} {
	v, e := self.Get(key)
	if nil != e {
		return defaultValue
	}
	return v
}

func (self *context) GetBoolWithDefault(key string, defaultValue bool) bool {
	v, e := self.Get(key)
	if nil != e {
		return defaultValue
	}

	res, e := commons.AsBool(v)
	if nil != e {
		return defaultValue
	}
	return res
}

func (self *context) GetIntWithDefault(key string, defaultValue int) int {
	v, e := self.Get(key)
	if nil != e {
		return defaultValue
	}

	res, e := commons.AsInt(v)
	if nil != e {
		return defaultValue
	}
	return res
}

func (self *context) GetInt32WithDefault(key string, defaultValue int32) int32 {
	v, e := self.Get(key)
	if nil != e {
		return defaultValue
	}

	res, e := commons.AsInt32(v)
	if nil != e {
		return defaultValue
	}
	return res
}

func (self *context) GetInt64WithDefault(key string, defaultValue int64) int64 {
	v, e := self.Get(key)
	if nil != e {
		return defaultValue
	}

	res, e := commons.AsInt64(v)
	if nil != e {
		return defaultValue
	}
	return res
}

func (self *context) GetUintWithDefault(key string, defaultValue uint) uint {
	v, e := self.Get(key)
	if nil != e {
		return defaultValue
	}

	res, e := commons.AsUint(v)
	if nil != e {
		return defaultValue
	}
	return res
}

func (self *context) GetUint32WithDefault(key string, defaultValue uint32) uint32 {
	v, e := self.Get(key)
	if nil != e {
		return defaultValue
	}

	res, e := commons.AsUint32(v)
	if nil != e {
		return defaultValue
	}
	return res
}

func (self *context) GetUint64WithDefault(key string, defaultValue uint64) uint64 {
	v, e := self.Get(key)
	if nil != e {
		return defaultValue
	}

	res, e := commons.AsUint64(v)
	if nil != e {
		return defaultValue
	}
	return res
}

func (self *context) GetStringWithDefault(key, defaultValue string) string {
	v, e := self.Get(key)
	if nil != e {
		return defaultValue
	}

	res, e := commons.AsString(v)
	if nil != e {
		return defaultValue
	}
	return res
}

func (self *context) GetArrayWithDefault(key string, defaultValue []interface{}) []interface{} {
	v, e := self.Get(key)
	if nil != e {
		return defaultValue
	}

	res, e := commons.AsArray(v)
	if nil != e {
		return defaultValue
	}
	return res
}

func (self *context) GetObjectWithDefault(key string, defaultValue map[string]interface{}) map[string]interface{} {
	v, e := self.Get(key)
	if nil != e {
		return defaultValue
	}

	res, e := commons.AsObject(v)
	if nil != e {
		return defaultValue
	}
	return res
}

func (self *context) GetObjectsWithDefault(key string, defaultValue []map[string]interface{}) []map[string]interface{} {
	v, e := self.Get(key)
	if nil != e {
		return defaultValue
	}

	res, e := commons.AsObjects(v)
	if nil != e {
		return defaultValue
	}
	return res
}

func (self *context) GetBool(key string) (bool, error) {
	v, e := self.Get(key)
	if nil != e {
		return false, e
	}

	return commons.AsBool(v)
}

func (self *context) GetInt(key string) (int, error) {
	v, e := self.Get(key)
	if nil != e {
		return 0, e
	}

	return commons.AsInt(v)
}

func (self *context) GetInt32(key string) (int32, error) {
	v, e := self.Get(key)
	if nil != e {
		return 0, e
	}

	return commons.AsInt32(v)
}

func (self *context) GetInt64(key string) (int64, error) {
	v, e := self.Get(key)
	if nil != e {
		return 0, e
	}

	return commons.AsInt64(v)
}

func (self *context) GetUint(key string) (uint, error) {
	v, e := self.Get(key)
	if nil != e {
		return 0, e
	}

	return commons.AsUint(v)
}

func (self *context) GetUint32(key string) (uint32, error) {
	v, e := self.Get(key)
	if nil != e {
		return 0, e
	}

	return commons.AsUint32(v)
}

func (self *context) GetUint64(key string) (uint64, error) {
	v, e := self.Get(key)
	if nil != e {
		return 0, e
	}

	return commons.AsUint64(v)
}

func (self *context) GetString(key string) (string, error) {
	v, e := self.Get(key)
	if nil != e {
		return "", e
	}

	return commons.AsString(v)
}

func (self *context) GetObject(key string) (map[string]interface{}, error) {
	v, e := self.Get(key)
	if nil != e {
		return nil, e
	}

	return commons.AsObject(v)
}

func (self *context) GetArray(key string) ([]interface{}, error) {
	v, e := self.Get(key)
	if nil != e {
		return nil, e
	}

	return commons.AsArray(v)
}

func (self *context) GetObjects(key string) ([]map[string]interface{}, error) {
	v, e := self.Get(key)
	if nil != e {
		return nil, e
	}

	return commons.AsObjects(v)
}
