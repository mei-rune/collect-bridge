package sampling

type proxy struct {
	*dispatcher
}

func (self *proxy) GetWithDefault(metric_name string, params MContext, defaultValue interface{}) interface{} {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if nil == res.InterfaceValue() {
		return defaultValue
	}
	return res.InterfaceValue()
}

func (self *proxy) GetBoolWithDefault(metric_name string, params MContext, defaultValue bool) bool {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if b, ok := res.Value().AsBool(); nil == ok {
		return b
	}
	return defaultValue
}

func (self *proxy) GetIntWithDefault(metric_name string, params MContext, defaultValue int) int {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if i, ok := res.Value().AsInt(); nil == ok {
		return i
	}
	return defaultValue
}

func (self *proxy) GetInt32WithDefault(metric_name string, params MContext, defaultValue int32) int32 {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if i, ok := res.Value().AsInt32(); nil == ok {
		return i
	}
	return defaultValue
}

func (self *proxy) GetInt64WithDefault(metric_name string, params MContext, defaultValue int64) int64 {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if i, ok := res.Value().AsInt64(); nil == ok {
		return i
	}
	return defaultValue
}

func (self *proxy) GetUintWithDefault(metric_name string, params MContext, defaultValue uint) uint {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if u, ok := res.Value().AsUint(); nil == ok {
		return u
	}
	return defaultValue
}

func (self *proxy) GetUint32WithDefault(metric_name string, params MContext, defaultValue uint32) uint32 {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if u, ok := res.Value().AsUint32(); nil == ok {
		return u
	}
	return defaultValue
}

func (self *proxy) GetUint64WithDefault(metric_name string, params MContext, defaultValue uint64) uint64 {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if u, ok := res.Value().AsUint64(); nil == ok {
		return u
	}
	return defaultValue
}

func (self *proxy) GetStringWithDefault(metric_name string, params MContext, defaultValue string) string {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if s, ok := res.Value().AsString(); nil == ok {
		return s
	}
	return defaultValue
}

func (self *proxy) GetArrayWithDefault(metric_name string, params MContext, defaultValue []interface{}) []interface{} {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if a, ok := res.Value().AsArray(); nil == ok {
		return a
	}
	return defaultValue
}

func (self *proxy) GetObjectWithDefault(metric_name string, params MContext,
	defaultValue map[string]interface{}) map[string]interface{} {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if a, ok := res.Value().AsObject(); nil == ok {
		return a
	}
	return defaultValue
}

func (self *proxy) GetObjectsWithDefault(metric_name string, params MContext,
	defaultValue []map[string]interface{}) []map[string]interface{} {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return defaultValue
	}
	if a, ok := res.Value().AsObjects(); nil == ok {
		return a
	}
	return defaultValue
}

func (self *proxy) Get(metric_name string, params MContext) (interface{}, error) {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return nil, res.Error()
	}
	return res.InterfaceValue(), nil
}

func (self *proxy) GetBool(metric_name string, params MContext) (bool, error) {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return false, res.Error()
	}
	return res.Value().AsBool()
}

func (self *proxy) GetInt(metric_name string, params MContext) (int, error) {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return 0, res.Error()
	}
	return res.Value().AsInt()
}

func (self *proxy) GetInt32(metric_name string, params MContext) (int32, error) {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return 0, res.Error()
	}
	return res.Value().AsInt32()
}

func (self *proxy) GetInt64(metric_name string, params MContext) (int64, error) {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return 0, res.Error()
	}
	return res.Value().AsInt64()
}

func (self *proxy) GetUint(metric_name string, params MContext) (uint, error) {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return 0, res.Error()
	}
	return res.Value().AsUint()
}

func (self *proxy) GetUint32(metric_name string, params MContext) (uint32, error) {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return 0, res.Error()
	}
	return res.Value().AsUint32()
}

func (self *proxy) GetUint64(metric_name string, params MContext) (uint64, error) {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return 0, res.Error()
	}
	return res.Value().AsUint64()
}

func (self *proxy) GetString(metric_name string, params MContext) (string, error) {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return "", res.Error()
	}
	return res.Value().AsString()
}

func (self *proxy) GetObject(metric_name string, params MContext) (map[string]interface{}, error) {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return nil, res.Error()
	}
	return res.Value().AsObject()
}

func (self *proxy) GetArray(metric_name string, params MContext) ([]interface{}, error) {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return nil, res.Error()
	}
	return res.Value().AsArray()
}

func (self *proxy) GetObjects(metric_name string, params MContext) ([]map[string]interface{}, error) {
	res := self.dispatcher.Get(metric_name, params)
	if res.HasError() {
		return nil, res.Error()
	}
	return res.Value().AsObjects()
}
