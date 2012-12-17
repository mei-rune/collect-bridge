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
	Insert(cd *ClassDefinition, properties map[string]interface{}, parents []ObjectId) (interface{}, error)
	Update(cd *ClassDefinition, id string, properties map[string]interface{}, parents []ObjectId) error
	FindById(cd *ClassDefinition, id string, parents []ObjectId) (interface{}, error)
	Delete(cd *ClassDefinition, id string, parents []ObjectId) error
}

type SqlInteger32 int32
type SqlInteger64 int64
type SqlDecimal float64
type SqlString string
type SqlDateTime time.Time
type SqlIPAddress net.IP
type SqlPhysicalAddress net.HardwareAddr
type SqlPassword string
