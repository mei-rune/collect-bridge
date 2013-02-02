package metrics

import (
	"commons"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Base struct {
	drvMgr *commons.DriverManager
	drv    *commons.Driver
}

func (self *Base) Init(drvMgr *commons.Driver, drvName string) commons.RuntimeError {
	self.drvMgr = drvMgr
	self.drv, _ = drvMgr.Connect(drvMgr)
	if nil == self.drv {
		return commons.NotFound(drvName)
	}
	return nil
}

func (self *Base) GetCopyWithPrefix(params map[string]string, prefix string) map[string]string {
	cp := map[string]string{}
	for k, v := range params {
		cp[k] = v
	}
	return GetWithPrefix(cp)
}

func (self *Base) GetWithPrefix(params map[string]string, prefix string) map[string]string {
	for k, v := range params {
		if strings.HasPrefix(k, prefix) {
			params[k[len(prefix):]] = v
		}
	}
	return params
}

func (self *Base) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return nil, commons.NotImplemented
}

func (self *Base) Put(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return nil, commons.NotImplemented
}

func (self *Base) Create(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return nil, commons.NotImplemented
}

func (self *Base) Delete(params map[string]string) (bool, commons.RuntimeError) {
	return false, commons.NotImplemented
}

type SystemOid struct {
	Base
}

func (self *SystemOid) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	pa := self.GetWithPrefix(params, "snmp.")
	pa["oid"] = "1.3.6.1.2.1.1.1.0"
	pa["action"] = "get"
	return self.drv.Get(pa)
}

type SystemDescr struct {
	Base
}

func (self *SystemDescr) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	pa := self.GetWithPrefix(params, "snmp.")
	pa["oid"] = "1.3.6.1.2.1.1.2.0"
	pa["action"] = "get"
	return self.drv.Get(pa)
}

type Interface struct {
	Base
}

func (self *Interface) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	pa := self.GetWithPrefix(params, "snmp.")
	pa["oid"] = "1.3.6.1.2.1.2.2.1"
	pa["action"] = "table"
	return self.drv.Get(pa)
}

func init() {
	METRIC_DRVS["sys.oid"] = func(params map[string]string, drvMgr *commons.DriverManager) (commons.Driver, commons.RuntimeError) {
		drv := &SystemOid{}
		return drv, drv.Init(drvMgr, "snmp")
	}
	METRIC_DRVS["sys.descr"] = func(params map[string]string, drvMgr *commons.DriverManager) (commons.Driver, commons.RuntimeError) {
		drv := &SystemDescr{}
		return drv, drv.Init(drvMgr, "snmp")
	}
	METRIC_DRVS["interface"] = func(params map[string]string, drvMgr *commons.DriverManager) (commons.Driver, commons.RuntimeError) {
		drv := &Interface{}
		return drv, drv.Init(drvMgr, "snmp")
	}
}
