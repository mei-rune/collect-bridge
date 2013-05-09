package types

import (
	"net"
)

// type ObjectId struct {
// 	definition *ClassDefinition
// 	id         string
// }

// type Iter interface {
// 	Next(result interface{}) bool
// 	Err() error
// }

// type Query interface {
// 	Select(fields ...string) Query
// 	Count() (n int, err error)
// 	Skip(offest int) Query
// 	Limit(n int) Query
// 	Sort(fields ...string) Query
// 	Iter() Iter
// 	Explain(ers interface{}) error
// }

// type Driver interface {
// 	Insert(cls *ClassDefinition, attributes map[string]interface{}) (interface{}, error)
// 	Update(cls *ClassDefinition, id interface{}, updated_attributes map[string]interface{}) error
// 	Query(cls string, attributes map[string]interface{}) Query
// 	FindBy(cls *ClassDefinition, attributes map[string]interface{}) Query
// 	FindById(cls *ClassDefinition, id interface{}) (map[string]interface{}, error)
// 	Delete(cls *ClassDefinition, id interface{}) error
// }

// type SqlInteger32 int32
// type int64 int64
// type SqlDecimal float64
// type SqlString string
type SqlPassword string

//type SqlDateTime time.Time

type SqlIPAddress net.IP
type SqlPhysicalAddress net.HardwareAddr

// func (self *SqlDateTime) GetBSON() (interface{}, error) {
// 	return time.Time(*self), nil
// }

// func (self *SqlDateTime) String() string {
// 	return time.Time(*self).Format(time.RFC3339)
// }

func (self *SqlIPAddress) GetBSON() (interface{}, error) {
	return net.IP(*self).String(), nil
}

func (self *SqlIPAddress) String() string {
	return net.IP(*self).String()
}

func (self *SqlIPAddress) MarshalJSON() ([]byte, error) {
	return []byte("\"" + net.IP(*self).String() + "\""), nil
}

func (self *SqlPhysicalAddress) GetBSON() (interface{}, error) {
	return net.HardwareAddr(*self).String(), nil
}

func (self *SqlPhysicalAddress) String() string {
	return net.HardwareAddr(*self).String()
}

func (self *SqlPhysicalAddress) MarshalJSON() ([]byte, error) {
	return []byte("\"" + net.HardwareAddr(*self).String() + "\""), nil
}
