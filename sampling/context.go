package sampling

import (
	"commons"
	ds "data_store"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	//"io/ioutil"
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
	srv              *server
	alias            map[string]string
	metric_name      string
	is_native        bool
	address          string
	managed_type     string
	managed_id       string
	query_paths      []P
	query_params     map[string]string
	body_reader      io.Reader
	body_instance    interface{}
	body_error       error
	body_unmarshaled bool

	mo            map[string]interface{}
	local         map[string]map[string]interface{}
	metrics_cache map[string]interface{}
}

func (self *context) init() error {
	if nil == self.query_params {
		self.query_params = map[string]string{}
	}

	if self.is_native {
		self.query_params["uid"] = self.address
		self.query_params["@address"] = self.address
		self.query_params["metric-name"] = self.metric_name
		self.mo = map[string]interface{}{}
	} else {
		self.query_params["type"] = self.managed_type
		self.query_params["id"] = self.managed_id
		self.query_params["uid"] = self.managed_id
		self.query_params["metric-name"] = self.metric_name

		var e error
		self.mo, e = self.srv.GetMOCahce(self.managed_id)
		if nil != e {
			return e
		}
		if nil == self.mo {
			return commons.RecordNotFoundWithType(self.managed_type, self.managed_id)
		}
		switch self.mo["type"] {
		case "network_device_port":
			device_id := fmt.Sprint(self.mo["device_id"])
			ifIndex := fmt.Sprint(self.mo["if_index"])
			self.mo, e = self.srv.GetMOCahce(device_id)
			if nil != e {
				return e
			}

			if nil == self.mo {
				return commons.RecordNotFoundWithType("device", self.managed_id)
			}
			self.query_paths = []P{{"port", ifIndex}}
		}
		// if "managed_object" == managed_type && "137" == managed_id {
		// 	bs, _ := json.MarshalIndent(mo, "", "  ")
		// 	fmt.Println(string(bs))
		// }
	}

	return nil
}

func (self *context) CreateCtx(metric_name string, managed_type, managed_id string) (MContext, error) {
	ctx := &context{srv: self.srv,
		alias:        alias_names,
		is_native:    false,
		managed_type: managed_type,
		managed_id:   managed_id}
	return ctx, ctx.init()
}

func (self *context) Body() (interface{}, error) {
	if self.body_unmarshaled {
		return self.body_instance, self.body_error
	}

	if nil == self.body_reader {
		return nil, errors.New("'body' is nil")
	}

	self.body_error = json.NewDecoder(self.body_reader).Decode(&self.body_instance)
	self.body_unmarshaled = true
	return self.body_instance, self.body_error
}

func (self *context) Read() Sampling {
	return self.srv
}

func (self *context) CopyTo(copy map[string]interface{}) {
	for k, v := range self.query_params {
		copy[k] = v
	}
}

func (self *context) Set(key string, value interface{}) {
	switch key[0] {
	// case '@': // thread safe?
	// 	self.mo[key[1:]] = value
	case '&', '!':
		if nil == self.metrics_cache {
			self.metrics_cache = make(map[string]interface{})
		}
		self.metrics_cache[key[1:]] = value
	default:
		if s, ok := value.(string); ok {
			self.query_params[key] = s
		} else {
			self.query_params[key] = fmt.Sprint(value)
		}
	}
}

func (self *context) cache(t, field string) (interface{}, error) {
	var m map[string]interface{}
	var n []map[string]interface{}
	var ok bool

	if nil != self.local {
		m, ok = self.local[t]
		if ok {
			goto ok
		}
	}

	if tn, ok := self.alias[t]; ok {
		t = tn

		if nil != self.local {
			if m, ok = self.local[tn]; ok {
				goto ok
			}
		}
	}

	n = ds.GetChildrenForm(self.mo["$attributes"],
		map[string]commons.Matcher{"type": commons.EqualString(t)})
	if nil == n || 0 == len(n) {
		return nil, TableNotExists
	}

	m = n[0]

	if nil == self.local {
		self.local = make(map[string]map[string]interface{})
	}

	self.local[t] = m
ok:
	if v, ok := m[field]; ok {
		return v, nil
	}
	return nil, NotFound
}

func (self *context) Contains(key string) bool {
	if _, ok := self.query_params[key]; ok {
		return ok
	}

	switch key[0] {
	case '@':
		_, ok := self.mo[key[1:]]
		return ok
	case '$':
		new_key := key[1:]
		if _, ok := self.query_params[new_key]; ok {
			return true
		}
		if _, ok := self.query_params["@"+new_key]; ok {
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
	if s, ok := self.query_params[key]; ok {
		return s, nil
	}

	switch key[0] {
	case '@':
		if v, ok := self.mo[key[1:]]; ok {
			return v, nil
		}
		return nil, NotFound
	case '$':
		new_key := key[1:]
		if s, ok := self.query_params[new_key]; ok {
			return s, nil
		}
		if s, ok := self.query_params["@"+new_key]; ok {
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

		v, e := self.srv.Get(new_key, nil, self)
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
		return nil, NotFound
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

func (self *context) GetFloatWithDefault(key string, defaultValue float64) float64 {
	v, e := self.Get(key)
	if nil != e {
		return defaultValue
	}

	res, e := commons.AsFloat64(v)
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

func (self *context) GetFloat(key string) (float64, error) {
	v, e := self.Get(key)
	if nil != e {
		return 0, e
	}

	return commons.AsFloat64(v)
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
