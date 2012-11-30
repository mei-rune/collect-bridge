package mdb

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

type PropertyDefinition struct {
	Name         string
	Type         TypeDefinition
	restriction  []Validator
	defaultValue interface{}
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
	for _, Validator := range self.restriction {
		if ok, err := Validator.Validate(obj); !ok {
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

	ownProperties map[string]*PropertyDefinition
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

func (self *ClassDefinitions) LoadProperty(xmlCls *XMLClassDefinition,
	pr *XMLPropertyDefinition, buffer bytes.Buffer) *PropertyDefinition {

	cpr := &PropertyDefinition{Name: pr.Name,
		Type:        GetTypeDefinition(pr.Restrictions.Type),
		restriction: make([]Validator, 0, 4)}

	if "" != pr.Restrictions.DefaultValue {
		var err error
		cpr.defaultValue, err = cpr.Type.ConvertFrom(pr.Restrictions.DefaultValue)
		if nil != err {
			buffer.WriteString("\tload property '" + pr.Name + "' of class '" +
				xmlCls.Name + "' failed, parse defaultValue '" +
				pr.Restrictions.DefaultValue + "' failed, " + err.Error() + "\n")
		}
	}
	if nil != pr.Restrictions.Enumerations && 0 != len(*pr.Restrictions.Enumerations) {
		validator, err := cpr.Type.CreateEnumerationValidator(*pr.Restrictions.Enumerations)
		if nil != err {
			buffer.WriteString("\tload property '" + pr.Name + "' of class '" +
				xmlCls.Name + "' failed, parse Enumerations '" +
				strings.Join(*pr.Restrictions.Enumerations, ",") + "' failed, " + err.Error() + "\n")
		} else {
			cpr.restriction = append(cpr.restriction, validator)
		}
	}
	if "" != pr.Restrictions.Pattern {
		validator, err := cpr.Type.CreatePatternValidator(pr.Restrictions.Pattern)
		if nil != err {
			buffer.WriteString("\tload property '" + pr.Name + "' of class '" +
				xmlCls.Name + "' failed, parse Pattern '" +
				pr.Restrictions.Pattern + "' failed, " + err.Error() + "\n")
		} else {
			cpr.restriction = append(cpr.restriction, validator)
		}
	}
	if "" != pr.Restrictions.MinValue || "" != pr.Restrictions.MaxValue {
		validator, err := cpr.Type.CreateRangeValidator(pr.Restrictions.MinValue,
			pr.Restrictions.MaxValue)
		if nil != err {
			buffer.WriteString("\tload property '" + pr.Name + "' of class '" +
				xmlCls.Name + "' failed, parse Range of Value '" +
				pr.Restrictions.MinValue + "," + pr.Restrictions.MaxValue +
				"' failed, " + err.Error() + "\n")
		} else {
			cpr.restriction = append(cpr.restriction, validator)
		}
	}
	if "" != pr.Restrictions.Length {
		validator, err := cpr.Type.CreateLengthValidator(pr.Restrictions.Length,
			pr.Restrictions.Length)
		if nil != err {
			buffer.WriteString("\tload property '" + pr.Name + "' of class '" +
				xmlCls.Name + "' failed, parse Length '" +
				pr.Restrictions.Length + "' failed, " + err.Error() + "\n")
		} else {
			cpr.restriction = append(cpr.restriction, validator)
		}
	}
	if "" != pr.Restrictions.MinLength || "" != pr.Restrictions.MaxLength {
		validator, err := cpr.Type.CreateLengthValidator(pr.Restrictions.MinLength,
			pr.Restrictions.MaxLength)
		if nil != err {
			buffer.WriteString("\tload property '" + pr.Name + "' of class '" +
				xmlCls.Name + "' failed, parse Range of Length '" +
				pr.Restrictions.MinLength + "," + pr.Restrictions.MaxLength +
				"' failed, " + err.Error() + "\n")
		} else {
			cpr.restriction = append(cpr.restriction, validator)
		}
	}

	return cpr
}

func (self *ClassDefinitions) LoadFromXml(nm string) error {
	f, err := ioutil.ReadFile("test/test1.xml")
	if nil != err {
		return fmt.Errorf("read file '%s' failed, %s", nm, err.Error())
	}

	var xml_definitions XMLClassDefinitions
	err = xml.Unmarshal(f, &xml_definitions)
	if nil != err {
		return fmt.Errorf("unmarshal xml '%s' failed, %s", nm, err.Error())
	}

	if nil == xml_definitions.Definitions || 0 == len(xml_definitions.Definitions) {
		return fmt.Errorf("unmarshal xml '%s' error, class definition is empty", nm)
	}

	var buffer bytes.Buffer
	for _, xmlDefinition := range xml_definitions.Definitions {
		_, ok := self.clsDefinitions[xmlDefinition.Name]
		if ok {
			buffer.WriteString("\tclass '" + xmlDefinition.Name +
				"' is aleady exists.\n")
			continue
		}

		var super *ClassDefinition = nil
		if "" != xmlDefinition.Base {
			super, ok := self.clsDefinitions[xmlDefinition.Base]
			if !ok || nil == super {
				buffer.WriteString("\tBase '" + xmlDefinition.Base +
					"' of class '" + xmlDefinition.Name + "' is not found.\n")
				continue
			}
		}

		cls := &ClassDefinition{Name: xmlDefinition.Name, Super: super}
		cls.ownProperties = make(map[string]*PropertyDefinition)
		for _, pr := range xmlDefinition.Properties {
			var cpr *PropertyDefinition = nil
			cpr = self.LoadProperty(&xmlDefinition, &pr, buffer)
			if nil != cpr {
				cls.ownProperties[cpr.Name] = cpr
			}
		}
		self.clsDefinitions[cls.Name] = cls
	}

	if 0 == buffer.Len() {
		return nil
	}
	return errors.New("load file '" + nm + "' error\n" + buffer.String())
}

func (self *ClassDefinitions) Find(nm string) *ClassDefinition {
	if cls, ok := self.clsDefinitions[nm]; ok {
		return cls
	}
	return nil
}
