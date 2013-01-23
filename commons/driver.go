package commons

import (
	"fmt"
)

type DriverManager map[string]Driver

func NewDriverManager() *DriverManager {
	drv := make(DriverManager)
	return &drv
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

// var (
//	drivers = NewDriverManager()
// )

// func Register(name string, driver Driver) {
//	drivers.Register(name, driver)
// }
// func Unregister(name string) {
//	drivers.Unregister(name)
// }

// func Connect(name string) (Driver, bool) {
//	return drivers.Connect(name)
// }

type Driver interface {
	Get(map[string]string) (map[string]interface{}, error)
	Put(map[string]string) (map[string]interface{}, error)
	Create(map[string]string) (bool, error)
	Delete(map[string]string) (bool, error)
}
