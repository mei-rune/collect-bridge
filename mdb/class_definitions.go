package mdb

import (
	"errors"
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

func (self *MutiErrorsError) Error() {
	return self.msg
}
func (self *MutiErrorsError) Errors() []error {
	return self.errs
}

func (self *PropertyDefinition) Validate(obj interface{}) (bool, []error) {
	if nil == self.restriction {
		return true, nil
	}

	var result bool = true
	var errs []error = make([]error, 0, len(self.restriction))
	for validatable := range self.restriction {
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
	super      ClassDefinition
	Name       string
	properties map[string]PropertyDefinition
}

func (self *ClassDefinition) CollectionName() string {
	if nil == super {
		return self.Name
	}

	return self.super.CollectionName()
}

type ClassDefinitions struct {
	clsDefinitions map[string]ClassDefinition
}

func (self *ClassDefinitions) LoadFromXml(nm string) error {
	var errs = make([]error, 0, 20)

	for clsDefinition := range readXml(nm) {
		super, ok := self.clsDefinitions[clsDefinition.Name]
		if ok {
			errs = append(errs, errors.New("class '"+clsDefinition.Name+
				"' is aleady exists."))
			continue
		}

		if "" != clsDefinition.Base {
			base, ok := self.clsDefinitions[clsDefinition.Base]
			if !ok || nil == base {
				errs = append(errs, errors.New("Base '"+clsDefinition.Base+
					"' of class '"+clsDefinition.Name+"' is not found."))
				continue
			}
		}

		cls = &ClassDefinition{Name: clsDefinition.Name}

		for pr := range clsDefinition.Properties {

		}
		self.clsDefinitions[cls.Name] = cls
	}

	if 0 == len(errs) {
		return nil
	}
	return &MutiErrorsError{errs: errs, msg: "load file '" + nm + "' failed."}
}

func (self *ClassDefinitions) Find(nm string) {
	if cls, ok := self.clsDefinitions[nm]; ok {
		return cls
	}
	return nil
}
