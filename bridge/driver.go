package main

import (
	"fmt"
)

var (
	drivers = make(map[string]Driver)
)

func Register(name string, driver Driver) {
	_, ok := drivers[name]
	if ok {
		panic(fmt.Errorf("'%s' always registred.", name))
	}
	drivers[name] = driver
}
func Unregister(name string) {
	delete(drivers, name)
}

func Connect(name string) (Driver, bool) {
	driver, ok := drivers[name]
	return driver, ok
}

type Driver interface {
	Get(map[string]string) (interface{}, error)
	Put(map[string]string) (interface{}, error)
	Create(map[string]string) (bool, error)
	Delete(map[string]string) (bool, error)
}
