package sampling

import (
	"commons"
)

func type_error(e error) error {
	return TypeError
}

type getter interface {
	Get(metric_name string, paths []P, params MContext) commons.Result
}

type proxy struct {
	srv getter
}

func (self *proxy) Get(metric_name string, paths []P, params MContext) (interface{}, error) {
	res := self.srv.Get(metric_name, paths, params)
	if res.HasError() {
		return nil, res.Error()
	}
	return res.InterfaceValue(), nil
}

func (self *proxy) GetBool(metric_name string, paths []P, params MContext) (bool, error) {
	res := self.srv.Get(metric_name, paths, params)
	if res.HasError() {
		return false, res.Error()
	}
	b, e := res.Value().AsBool()
	if nil != e {
		return false, type_error(e)
	}
	return b, nil
}

func (self *proxy) GetInt(metric_name string, paths []P, params MContext) (int, error) {
	res := self.srv.Get(metric_name, paths, params)
	if res.HasError() {
		return 0, res.Error()
	}
	i, e := res.Value().AsInt()
	if nil != e {
		return 0, type_error(e)
	}
	return i, nil
}

func (self *proxy) GetInt32(metric_name string, paths []P, params MContext) (int32, error) {
	res := self.srv.Get(metric_name, paths, params)
	if res.HasError() {
		return 0, res.Error()
	}
	i32, e := res.Value().AsInt32()
	if nil != e {
		return 0, type_error(e)
	}
	return i32, nil
}

func (self *proxy) GetInt64(metric_name string, paths []P, params MContext) (int64, error) {
	res := self.srv.Get(metric_name, paths, params)
	if res.HasError() {
		return 0, res.Error()
	}
	i64, e := res.Value().AsInt64()
	if nil != e {
		return 0, type_error(e)
	}
	return i64, nil
}

func (self *proxy) GetUint(metric_name string, paths []P, params MContext) (uint, error) {
	res := self.srv.Get(metric_name, paths, params)
	if res.HasError() {
		return 0, res.Error()
	}
	u, e := res.Value().AsUint()
	if nil != e {
		return 0, type_error(e)
	}
	return u, nil
}

func (self *proxy) GetUint32(metric_name string, paths []P, params MContext) (uint32, error) {
	res := self.srv.Get(metric_name, paths, params)
	if res.HasError() {
		return 0, res.Error()
	}
	u32, e := res.Value().AsUint32()
	if nil != e {
		return 0, type_error(e)
	}
	return u32, nil
}

func (self *proxy) GetUint64(metric_name string, paths []P, params MContext) (uint64, error) {
	res := self.srv.Get(metric_name, paths, params)
	if res.HasError() {
		return 0, res.Error()
	}
	u64, e := res.Value().AsUint64()
	if nil != e {
		return 0, type_error(e)
	}
	return u64, nil
}

func (self *proxy) GetString(metric_name string, paths []P, params MContext) (string, error) {
	res := self.srv.Get(metric_name, paths, params)
	if res.HasError() {
		return "", res.Error()
	}

	s, e := res.Value().AsString()
	if nil != e {
		return "", type_error(e)
	}
	return s, nil
}

func (self *proxy) GetObject(metric_name string, paths []P, params MContext) (map[string]interface{}, error) {
	res := self.srv.Get(metric_name, paths, params)
	if res.HasError() {
		return nil, res.Error()
	}

	o, e := res.Value().AsObject()
	if nil != e {
		return nil, type_error(e)
	}
	return o, nil
}

func (self *proxy) GetArray(metric_name string, paths []P, params MContext) ([]interface{}, error) {
	res := self.srv.Get(metric_name, paths, params)
	if res.HasError() {
		return nil, res.Error()
	}

	a, e := res.Value().AsArray()
	if nil != e {
		return nil, type_error(e)
	}
	return a, nil
}

func (self *proxy) GetObjects(metric_name string, paths []P, params MContext) ([]map[string]interface{}, error) {
	res := self.srv.Get(metric_name, paths, params)
	if res.HasError() {
		return nil, res.Error()
	}

	o, e := res.Value().AsObjects()
	if nil != e {
		return nil, type_error(e)
	}
	return o, nil
}
