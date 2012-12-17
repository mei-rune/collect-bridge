package mdb

import (
	"net"
	"time"
)

type ObjectId struct {
	definition *ClassDefinition
	id         string
}
type Driver interface {
	Insert(cd *ClassDefinition, properties map[string]interface{}) (interface{}, error)
	Update(cd *ClassDefinition, id string, properties map[string]interface{}) error
	FindById(cd *ClassDefinition, id string) (interface{}, error)
	Delete(cd *ClassDefinition, id string) error
}

type SqlInteger32 int32
type SqlInteger64 int64
type SqlDecimal float64
type SqlString string
type SqlPassword string
type SqlDateTime time.Time

type SqlIPAddress net.IP
type SqlPhysicalAddress net.HardwareAddr

func (self *SqlDateTime) GetBSON() (interface{}, error) {
	return time.Time(*self), nil
}

func (self *SqlIPAddress) GetBSON() (interface{}, error) {
	return net.IP(*self).String(), nil
}

func (self *SqlPhysicalAddress) GetBSON() (interface{}, error) {
	return net.HardwareAddr(*self).String(), nil
}
