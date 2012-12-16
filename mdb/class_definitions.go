package mdb

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

type ClassDefinitions struct {
	clsDefinitions map[string]*ClassDefinition
}

func (self *ClassDefinitions) loadParentProperties(cls *ClassDefinition, errs []error) []error {
	if nil != cls.properties {
		return errs
	}
	cls.properties = make(map[string]*PropertyDefinition, 2*len(cls.ownProperties))
	if nil != cls.Super {
		errs = self.loadParentProperties(cls.Super, errs)
		for k, v := range cls.Super.properties {
			cls.properties[k] = v
		}
	}

	for k, v := range cls.ownProperties {
		old, ok := cls.properties[k]
		if ok {
			if v.Type != old.Type {
				errs = append(errs, errors.New("The property with '"+k+
					"' override failed, type is not same, own is '"+
					v.Type.Name()+"', super is '"+old.Type.Name()+"'"))
			}
			for _, r := range v.restriction {
				old.restriction = append(old.restriction, r)
			}
			if nil != v.defaultValue {
				old.defaultValue = v.defaultValue
			}
		} else {
			cls.properties[k] = v
		}
	}
	return errs
}

func (self *ClassDefinitions) loadOwnProperties(xmlCls *XMLClassDefinition,
	cls *ClassDefinition, errs []error) []error {
	cls.ownProperties = make(map[string]*PropertyDefinition)
	for _, pr := range xmlCls.Properties {
		var cpr *PropertyDefinition = nil
		cpr, errs = self.loadOwnProperty(xmlCls, &pr, errs)
		if nil != cpr {
			cls.ownProperties[cpr.Name] = cpr
		}
	}
	return errs
}

func (self *ClassDefinitions) loadOwnProperty(xmlCls *XMLClassDefinition,
	pr *XMLPropertyDefinition, errs []error) (*PropertyDefinition, []error) {

	cpr := &PropertyDefinition{Name: pr.Name,
		Type:        GetTypeDefinition(pr.Restrictions.Type),
		restriction: make([]Validator, 0, 4)}

	if "" != pr.Restrictions.DefaultValue {
		var err error
		cpr.defaultValue, err = cpr.Type.ConvertFrom(pr.Restrictions.DefaultValue)
		if nil != err {
			errs = append(errs, errors.New("load property '"+pr.Name+"' of class '"+
				xmlCls.Name+"' failed, parse defaultValue '"+
				pr.Restrictions.DefaultValue+"' failed, "+err.Error()))
		}
	}

	if nil != pr.Restrictions.Enumerations && 0 != len(*pr.Restrictions.Enumerations) {
		validator, err := cpr.Type.CreateEnumerationValidator(*pr.Restrictions.Enumerations)
		if nil != err {
			errs = append(errs, errors.New("load property '"+pr.Name+"' of class '"+
				xmlCls.Name+"' failed, parse Enumerations '"+
				strings.Join(*pr.Restrictions.Enumerations, ",")+"' failed, "+err.Error()))
		} else {
			cpr.restriction = append(cpr.restriction, validator)
		}
	}
	if "" != pr.Restrictions.Pattern {
		validator, err := cpr.Type.CreatePatternValidator(pr.Restrictions.Pattern)
		if nil != err {
			errs = append(errs, errors.New("load property '"+pr.Name+"' of class '"+
				xmlCls.Name+"' failed, parse Pattern '"+
				pr.Restrictions.Pattern+"' failed, "+err.Error()))
		} else {
			cpr.restriction = append(cpr.restriction, validator)
		}
	}
	if "" != pr.Restrictions.MinValue || "" != pr.Restrictions.MaxValue {
		validator, err := cpr.Type.CreateRangeValidator(pr.Restrictions.MinValue,
			pr.Restrictions.MaxValue)
		if nil != err {
			errs = append(errs, errors.New("load property '"+pr.Name+"' of class '"+
				xmlCls.Name+"' failed, parse Range of Value '"+
				pr.Restrictions.MinValue+","+pr.Restrictions.MaxValue+
				"' failed, "+err.Error()))
		} else {
			cpr.restriction = append(cpr.restriction, validator)
		}
	}
	if "" != pr.Restrictions.Length {
		validator, err := cpr.Type.CreateLengthValidator(pr.Restrictions.Length,
			pr.Restrictions.Length)
		if nil != err {
			errs = append(errs, errors.New("load property '"+pr.Name+"' of class '"+
				xmlCls.Name+"' failed, parse Length '"+
				pr.Restrictions.Length+"' failed, "+err.Error()))
		} else {
			cpr.restriction = append(cpr.restriction, validator)
		}
	}
	if "" != pr.Restrictions.MinLength || "" != pr.Restrictions.MaxLength {
		validator, err := cpr.Type.CreateLengthValidator(pr.Restrictions.MinLength,
			pr.Restrictions.MaxLength)
		if nil != err {
			errs = append(errs, errors.New("load property '"+pr.Name+"' of class '"+
				xmlCls.Name+"' failed, parse Range of Length '"+
				pr.Restrictions.MinLength+","+pr.Restrictions.MaxLength+
				"' failed, "+err.Error()))
		} else {
			cpr.restriction = append(cpr.restriction, validator)
		}
	}

	if nil == cpr.Type {
		errs = append(errs, errors.New("load property '"+pr.Name+"' of class '"+
			xmlCls.Name+"' failed, '"+pr.Restrictions.Type+
			"' is unsupported type"))
		return nil, errs
	}

	return cpr, errs
}

func (self *ClassDefinitions) loadAssocations(cls *ClassDefinition, xmlDefinition *XMLClassDefinition, errs []error) []error {
	if nil != xmlDefinition.BelongsTo && 0 != len(xmlDefinition.BelongsTo) {
		for _, belongs_to := range xmlDefinition.BelongsTo {
			target, ok := self.clsDefinitions[belongs_to.Target]
			if !ok {
				errs = append(errs, errors.New("Target '"+belongs_to.Target+
					"' of belongs_to '"+belongs_to.Name+"' is not found."))
				continue
			}

			pr, ok := cls.ownProperties[belongs_to.Name]
			if !ok {
				errs = append(errs, errors.New("Property '"+belongs_to.Name+
					"' of belongs_to '"+belongs_to.Name+"' is not found."))
				continue
			}
			if nil == cls.assocations {
				cls.assocations = make([]Assocation, 0, 4)
			}

			cls.assocations = append(cls.assocations, &BelongsTo{TargetClass: target, Name: pr})
		}
	}
	if nil != xmlDefinition.HasMany && 0 != len(xmlDefinition.HasMany) {
		for _, belongs_to := range xmlDefinition.HasMany {
			target, ok := self.clsDefinitions[belongs_to.Target]
			if !ok {
				errs = append(errs, errors.New("Target '"+belongs_to.Target+
					"' of has_many is not found."))
				continue
			}
			if nil == cls.assocations {
				cls.assocations = make([]Assocation, 0, 4)
			}
			cls.assocations = append(cls.assocations, &HasMany{TargetClass: target})
		}
	}
	if nil != xmlDefinition.HasAndBelongsToMany && 0 != len(xmlDefinition.HasAndBelongsToMany) {
		for _, belongs_to := range xmlDefinition.HasAndBelongsToMany {
			target, ok := self.clsDefinitions[belongs_to.Target]
			if !ok {
				errs = append(errs, errors.New("Target '"+belongs_to.Target+
					"' of has_and_belongs_to_many is not found."))
				continue
			}
			if nil == cls.assocations {
				cls.assocations = make([]Assocation, 0, 4)
			}
			cls.assocations = append(cls.assocations, &HasAndBelongsToMany{TargetClass: target})
		}
	}
	return errs
}

func (self *ClassDefinitions) LoadXml(nm string) error {
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

	errs := make([]error, 0, 10)

	// load class definitions and own properties
	for _, xmlDefinition := range xml_definitions.Definitions {
		_, ok := self.clsDefinitions[xmlDefinition.Name]
		if ok {
			errs = append(errs, errors.New("class '"+xmlDefinition.Name+
				"' is aleady exists."))
			continue
		}

		cls := &ClassDefinition{Name: xmlDefinition.Name}
		errs = self.loadOwnProperties(&xmlDefinition, cls, errs)

		self.clsDefinitions[cls.Name] = cls
	}

	// load super class and own assocations
	for _, xmlDefinition := range xml_definitions.Definitions {
		cls, ok := self.clsDefinitions[xmlDefinition.Name]
		if !ok {
			continue
		}
		if "" == xmlDefinition.Base {
			super, ok := self.clsDefinitions[xmlDefinition.Base]
			if !ok || nil == super {
				errs = append(errs, errors.New("Base '"+xmlDefinition.Base+
					"' of class '"+xmlDefinition.Name+"' is not found."))
			} else {
				cls.Super = super
			}
		}
		errs = self.loadAssocations(cls, &xmlDefinition, errs)
	}

	// load the properties of super class
	for _, cls := range self.clsDefinitions {
		errs = self.loadParentProperties(cls, errs)
	}

	if 0 == len(errs) {
		return nil
	}
	return &MutiErrors{msg: "load file '" + nm + "' error", errs: errs}
}

func (self *ClassDefinitions) Find(nm string) *ClassDefinition {
	if cls, ok := self.clsDefinitions[nm]; ok {
		return cls
	}
	return nil
}
