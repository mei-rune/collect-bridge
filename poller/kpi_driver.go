package poller

import (
	"commons"
)

type KPIDriver struct {
	Address string
}

func NewKPIDriver(address string) (commons.Driver, error) {
	return &KPIDriver{Address: address}, nil
}

func (self *KPIDriver) Get(params map[string]string) (commons.Result, commons.RuntimeError) {
	return nil, commons.NotImplemented
}

func (self *KPIDriver) Put(params map[string]string) (commons.Result, commons.RuntimeError) {
	return nil, commons.NotImplemented
}

func (self *KPIDriver) Create(params map[string]string) (commons.Result, commons.RuntimeError) {
	return nil, commons.NotImplemented
}

func (self *KPIDriver) Delete(params map[string]string) (commons.Result, commons.RuntimeError) {
	return nil, commons.NotImplemented
}
