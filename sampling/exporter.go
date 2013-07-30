package sampling

import (
	"encoding/json"
)

type exportable interface {
	stats() interface{}
}

type exporter struct {
	Var exportable
}

func (self *exporter) String() string {
	if nil == self.Var {
		return ""
	}
	v := self.Var.stats()
	if nil == v {
		return "null"
	}

	bs, e := json.MarshalIndent(v, "", "  ")
	if nil != e {
		return e.Error()
	}
	return string(bs)
}
