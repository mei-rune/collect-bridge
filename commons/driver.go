package commons

import (
	"fmt"
)

type DriverManager map[string]Driver

func NewDriverManager() DriverManager {
	return make(map[string]Driver)
}

func (self DriverManager) Register(name string, driver Driver) {
	_, ok := drivers[name]
	if ok {
		panic(fmt.Errorf("'%s' always registred.", name))
	}
	drivers[name] = driver
}
func (self DriverManager) Unregister(name string) {
	delete(drivers, name)
}

func (self DriverManager) Connect(name string) (Driver, bool) {
	driver, ok := drivers[name]
	return driver, ok
}

var (
	drivers = NewDriverManager()
)

func Register(name string, driver Driver) {
	drivers.Register(name, driver)
}
func Unregister(name string) {
	drivers.Unregister(name)
}

func Connect(name string) (Driver, bool) {
	return drivers.Connect(name)
}

type Driver interface {
	Get(map[string]string) (interface{}, error)
	Put(map[string]string) (interface{}, error)
	Create(map[string]string) (bool, error)
	Delete(map[string]string) (bool, error)
}
