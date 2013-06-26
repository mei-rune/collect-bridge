package metrics

type metric_proxy struct {
	*dispatcher
}

func (self *metric_proxy) GetWithDefault(metric_name string, params commons.Map, defaultValue interface{}) interface{} {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if nil == res.InterfaceValue() {
		return defaultValue
	}
	return res.InterfaceValue()
}

func (self *metric_proxy) GetBoolWithDefault(metric_name string, params commons.Map, defaultValue bool) bool {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if b, ok := res.Value().AsBool(); ok {
		return b
	}
	return defaultValue
}

func (self *metric_proxy) GetIntWithDefault(metric_name string, params commons.Map, defaultValue int) int {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if i, ok := res.Value().AsInt(); ok {
		return i
	}
	return defaultValue
}

func (self *metric_proxy) GetInt32WithDefault(metric_name string, params commons.Map, defaultValue int32) int32 {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if i, ok := res.Value().AsInt32(); ok {
		return i
	}
	return defaultValue
}

func (self *metric_proxy) GetInt64WithDefault(metric_name string, params commons.Map, defaultValue int64) int64 {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if i, ok := res.Value().AsInt64(); ok {
		return i
	}
	return defaultValue
}

func (self *metric_proxy) GetUintWithDefault(metric_name string, params commons.Map, defaultValue uint) uint {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if u, ok := res.Value().AsUint(); ok {
		return u
	}
	return defaultValue
}

func (self *metric_proxy) GetUint32WithDefault(metric_name string, params commons.Map, defaultValue uint32) uint32 {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if u, ok := res.Value().AsUint32(); ok {
		return u
	}
	return defaultValue
}

func (self *metric_proxy) GetUint64WithDefault(metric_name string, params commons.Map, defaultValue uint64) uint64 {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if u, ok := res.Value().AsUint64(); ok {
		return u
	}
	return defaultValue
}

func (self *metric_proxy) GetStringWithDefault(key, defaultValue string) string {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if s, ok := res.Value().AsString(); ok {
		return s
	}
	return defaultValue
}

func (self *metric_proxy) GetArrayWithDefault(metric_name string, params commons.Map, defaultValue []interface{}) []interface{} {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if a, ok := res.Value().AsArray(); ok {
		return a
	}
	return defaultValue
}

func (self *metric_proxy) GetObjectWithDefault(metric_name string, params commons.Map, defaultValue map[string]interface{}) map[string]interface{} {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if a, ok := res.Value().AsObject(); ok {
		return a
	}
	return defaultValue
}

func (self *metric_proxy) GetObjectsWithDefault(metric_name string, params commons.Map, defaultValue []map[string]interface{}) []map[string]interface{} {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if a, ok := res.Value().AsObjects(); ok {
		return a
	}
	return defaultValue
}

func (self *metric_proxy) GetBool(metric_name string, params commons.Map) (bool, commons.RuntimeError) {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return false, res.Error()
	}
	return res.Value().AsBool()
}

func (self *metric_proxy) GetInt(metric_name string, params commons.Map) (int, commons.RuntimeError) {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return 0, res.Error()
	}
	return res.Value().AsInt()
}

func (self *metric_proxy) GetInt32(metric_name string, params commons.Map) (int32, commons.RuntimeError) {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return 0, res.Error()
	}
	return res.Value().AsInt32()
}

func (self *metric_proxy) GetInt64(metric_name string, params commons.Map) (int64, commons.RuntimeError) {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return 0, res.Error()
	}
	return res.Value().AsInt64()
}

func (self *metric_proxy) GetUint(metric_name string, params commons.Map) (uint, commons.RuntimeError) {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return 0, res.Error()
	}
	return res.Value().AsUint()
}

func (self *metric_proxy) GetUint32(metric_name string, params commons.Map) (uint32, commons.RuntimeError) {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return 0, res.Error()
	}
	return res.Value().AsUint32()
}

func (self *metric_proxy) GetUint64(metric_name string, params commons.Map) (uint64, commons.RuntimeError) {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return 0, res.Error()
	}
	return res.Value().AsUint64()
}

func (self *metric_proxy) GetString(metric_name string, params commons.Map) (string, commons.RuntimeError) {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return "", res.Error()
	}
	return res.Value().AsString()
}

func (self *metric_proxy) GetObject(metric_name string, params commons.Map) (map[string]interface{}, commons.RuntimeError) {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return nil, res.Error()
	}
	return res.Value().AsObject()
}

func (self *metric_proxy) GetArray(metric_name string, params commons.Map) ([]interface{}, commons.RuntimeError) {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return nil, res.Error()
	}
	return res.Value().AsArray()
}

func (self *metric_proxy) GetObjects(metric_name string, params commons.Map) ([]map[string]interface{}, commons.RuntimeError) {
	res := self.Get(metric_name, params)
	if res.HasError() {
		return nil, res.Error()
	}
	return res.Value().AsObjects()
}
