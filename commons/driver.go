package commons

import (
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

type Driver interface {
	Get(map[string]string) (map[string]interface{}, RuntimeError)
	Put(map[string]string) (map[string]interface{}, RuntimeError)
	Create(map[string]string) (map[string]interface{}, RuntimeError)
	Delete(map[string]string) (bool, RuntimeError)
}

type DefaultDrv struct {
	GetValue, PutValue, CreateValue      interface{}
	DeleteValue                          bool
	GetErr, PutErr, CreateErr, DeleteErr RuntimeError
}

func (self *DefaultDrv) Get(params map[string]string) (map[string]interface{}, RuntimeError) {
	return Return(self.GetValue), self.GetErr
}

func (self *DefaultDrv) Put(params map[string]string) (map[string]interface{}, RuntimeError) {
	return Return(self.PutValue), self.PutErr
}

func (self *DefaultDrv) Create(params map[string]string) (map[string]interface{}, RuntimeError) {
	return Return(self.CreateValue), self.CreateErr
}

func (self *DefaultDrv) Delete(params map[string]string) (bool, RuntimeError) {
	return self.DeleteValue, self.DeleteErr
}
