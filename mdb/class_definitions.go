package mdb

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
)

type PropertyDefinition struct {
	Name        string
	Type        TypeDefinition
	restriction []Validatable
}

type MutiErrorsError struct {
	msg  string
	errs []error
}

func (self *MutiErrorsError) Error() string {
	return self.msg
}
func (self *MutiErrorsError) Errors() []error {
	return self.errs
}

func (self *PropertyDefinition) Validate(obj interface{}) (bool, error) {
	if nil == self.restriction {
		return true, nil
	}

	var result bool = true
	var errs []error = make([]error, 0, len(self.restriction))
	for _, validatable := range self.restriction {
		if ok, err := validatable.Validate(obj); !ok {
			result = false
			errs = append(errs, err)
		}
	}

	if result {
		return true, nil
	}
	return false, &MutiErrorsError{errs: errs, msg: "property '" + self.Name + "' is error"}
}

type ClassDefinition struct {
	Super      *ClassDefinition
	Name       string
	properties map[string]*PropertyDefinition
}

func (self *ClassDefinition) CollectionName() string {
	if nil == self.Super {
		return self.Name
	}

	return self.Super.CollectionName()
}

type ClassDefinitions struct {
	clsDefinitions map[string]*ClassDefinition
}

func (self *ClassDefinitions) LoadProperty(pr *XMLPropertyDefinition, errs []error) (*PropertyDefinition, []error) {
	// TODO
	panic("not implemented")
}

func (self *ClassDefinitions) LoadFromXml(nm string) error {
	bytes, err := ioutil.ReadFile("test/test1.xml")
	if nil != err {
		return fmt.Errorf("read file '%s' failed, %s", nm, err.Error())
	}

	var xml_definitions XMLClassDefinitions
	err = xml.Unmarshal(bytes, &xml_definitions)
	if nil != err {
		return fmt.Errorf("unmarshal xml '%s' failed, %s", nm, err.Error())
	}

	if nil == xml_definitions.Definitions || 0 == len(xml_definitions.Definitions) {
		return fmt.Errorf("unmarshal xml '%s' error, class definition is empty", nm)
	}

	var errs = make([]error, 0, 20)
	for _, xmlDefinition := range xml_definitions.Definitions {
		_, ok := self.clsDefinitions[xmlDefinition.Name]
		if ok {
			errs = append(errs, errors.New("class '"+xmlDefinition.Name+
				"' is aleady exists."))
			continue
		}

		var super *ClassDefinition = nil
		if "" != xmlDefinition.Base {
			super, ok := self.clsDefinitions[xmlDefinition.Base]
			if !ok || nil == super {
				errs = append(errs, errors.New("Base '"+xmlDefinition.Base+
					"' of class '"+xmlDefinition.Name+"' is not found."))
				continue
			}
		}

		cls := &ClassDefinition{Name: xmlDefinition.Name, Super: super}
		cls.properties = make(map[string]*PropertyDefinition)
		for _, pr := range xmlDefinition.Properties {
			var cpr *PropertyDefinition = nil
			cpr, errs = self.LoadProperty(&pr, errs)
			if nil != cpr {
				cls.properties[cpr.Name] = cpr
			}
		}
		self.clsDefinitions[cls.Name] = cls
	}

	if 0 == len(errs) {
		return nil
	}
	return &MutiErrorsError{errs: errs, msg: "load file '" + nm + "' failed."}
}

func (self *ClassDefinitions) Find(nm string) *ClassDefinition {
	if cls, ok := self.clsDefinitions[nm]; ok {
		return cls
	}
	return nil
}
