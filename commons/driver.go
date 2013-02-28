package commons

import (
	"commons/as"
	"fmt"
)

type DriverManager map[string]Driver

func NewDriverManager() *DriverManager {
	drv := make(DriverManager)
	return &drv
}

func (self *DriverManager) Stop(name string) {
	drv, ok := (*self)[name]
	if !ok {
		return
	}
	startable, ok := drv.(Startable)
	if ok && nil != startable {
		startable.Stop()
	}
}

func (self *DriverManager) Start(name string) error {
	drv, ok := (*self)[name]
	if !ok {
		return NotFound(name)
	}

	startable, ok := drv.(Startable)
	if ok && nil != startable {
		err := startable.Start()
		if nil != err {
			return err
		}
	}
	return nil
}

func (self *DriverManager) Reset(name string) error {
	drv, ok := (*self)[name]
	if !ok {
		return NotFound(name)
	}

	startable, ok := drv.(Startable)
	if ok && nil != startable {
		startable.Stop()
		err := startable.Start()
		if nil != err {
			return err
		}
	}
	return nil
}

func (self *DriverManager) Register(name string, driver Driver) {
	_, ok := (*self)[name]
	if ok {
		panic(fmt.Errorf("'%s' always registred.", name))
	}
	(*self)[name] = driver
}

func (self *DriverManager) Unregister(name string) {
	delete(*self, name)
}

func (self *DriverManager) Connect(name string) (Driver, bool) {
	driver, ok := (*self)[name]
	return driver, ok
}

func (self *DriverManager) Names() []string {
	names := make([]string, 0, 10)
	for k, _ := range *self {
		names = append(names, k)
	}
	return names
}

var (
	METRIC_DRVS = map[string]func(ctx map[string]interface{}) (Driver, RuntimeError){}
)

// func Register(name string, driver Driver) {
//	drivers.Register(name, driver)
// }
// func Unregister(name string) {
//	drivers.Unregister(name)
// }

// func Connect(name string) (Driver, bool) {
//	return drivers.Connect(name)
// }

type Startable interface {
	Start() error
	Stop()
}

type Result map[string]interface{}

type Driver interface {
	Get(map[string]string) (Result, RuntimeError)
	Put(map[string]string) (Result, RuntimeError)
	Create(map[string]string) (Result, RuntimeError)
	Delete(map[string]string) (Result, RuntimeError)
}

type DefaultDrv struct {
	GetValue, PutValue, CreateValue, DeleteValue interface{}
	GetErr, PutErr, CreateErr, DeleteErr         RuntimeError
}

func (self *DefaultDrv) Get(params map[string]string) (Result, RuntimeError) {
	return Return(self.GetValue), self.GetErr
}

func (self *DefaultDrv) Put(params map[string]string) (Result, RuntimeError) {
	return Return(self.PutValue), self.PutErr
}

func (self *DefaultDrv) Create(params map[string]string) (Result, RuntimeError) {
	return Return(self.CreateValue), self.CreateErr
}

func (self *DefaultDrv) Delete(params map[string]string) (Result, RuntimeError) {
	return Return(self.DeleteValue), self.DeleteErr
}

func GetReturn(params Result) interface{} {
	v, ok := params["value"]
	if ok {
		return v
	}
	return nil
}

func GetReturnCode(params Result) int {
	v, ok := params["code"]
	if ok {
		i, e := as.AsInt(v)
		if nil != e {
			panic(e.Error())
		}
		return i
	}
	return -1
}

func Return(value interface{}) Result {
	return Result{"value": value}
}

func (self Result) Return(value interface{}) Result {
	self["value"] = value
	return self
}

func (self Result) With(key string, value interface{}) Result {
	self[key] = value
	return self
}

func (self Result) Warnings(value interface{}) Result {
	if nil == value || "" == value {
		return self
	}
	self["warnings"] = value
	return self
}

func (self Result) Effected(effected int) Result {
	self["effected"] = effected
	return self
}

func (self Result) TryGetBool(key string) (bool, error) {
	return TryGetBool(self, key)
}

func (self Result) TryGetInt(key string) (int, error) {
	return TryGetInt(self, key)
}

func (self Result) TryGetInt64(key string) (int64, error) {
	return TryGetInt64(self, key)
}

func (self Result) TryGetUint(key string) (uint, error) {
	return TryGetUint(self, key)
}

func (self Result) TryGetUint64(key string) (uint64, error) {
	return TryGetUint64(self, key)
}

func (self Result) TryGetString(key string) (string, error) {
	return TryGetString(self, key)
}

func (self Result) GetEffected() int {
	i, e := self.TryGetInt("effected")
	if nil != e {
		panic(e)
	}
	return i
}

func (self Result) Get(key string) interface{} {
	return self[key]
}

func (self Result) GetWarnings() interface{} {
	return self["warnings"]
}
