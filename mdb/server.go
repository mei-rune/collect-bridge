package mdb

import (
	"errors"
)

type MdbServer struct {
	driver      Driver
	definitions ClassDefinitions
}

func (self *MdbServer) validate(cls *ClassDefinition, attributes map[string]interface{}) (map[string]interface{}, error) {
	//new_attributes := make(map[string]interface{}, len(attributes))
	return nil, errors.New("not implemented")
}

func (self *MdbServer) Create(cls *ClassDefinition, attributes map[string]interface{}) (interface{}, error) {
	//attributes, errs := self.validate(cls, attributes)
	return nil, errors.New("not implemented")

}
func (self *MdbServer) FindById(cls *ClassDefinition, id interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (self *MdbServer) Update(cls *ClassDefinition, id interface{}, attributes map[string]interface{}) error {
	return errors.New("not implemented")
}

func (self *MdbServer) RemoveById(cls *ClassDefinition, id interface{}) error {
	return errors.New("not implemented")
}
