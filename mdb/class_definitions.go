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

func loadParentProperties(self *ClassDefinitions, cls *ClassDefinition, errs []error) []error {
	if nil != cls.Properties {
		return errs
	}
	cls.Properties = make(map[string]*PropertyDefinition, 2*len(cls.OwnProperties))
	if nil != cls.Super {
		errs = loadParentProperties(self, cls.Super, errs)
		for k, v := range cls.Super.Properties {
			cls.Properties[k] = v
		}
	}

	for k, v := range cls.OwnProperties {
		old, ok := cls.Properties[k]
		if ok {
			if v.Type != old.Type {
				errs = append(errs, errors.New("The property with '"+k+
					"' override failed, type is not same, own is '"+
					v.Type.Name()+"', super is '"+old.Type.Name()+"'"))
			}

			if nil != old.Restrictions {

				if nil == v.Restrictions {
					v.Restrictions = make([]Validator, 0)
				}

				for _, r := range old.Restrictions {
					v.Restrictions = append(v.Restrictions, r)
				}
			}
			if nil == v.DefaultValue {
				v.DefaultValue = old.DefaultValue
			}
		}
		cls.Properties[k] = v
	}
	return errs
}

func loadOwnProperties(self *ClassDefinitions, xmlCls *XMLClassDefinition,
	cls *ClassDefinition, errs []error) []error {
	cls.OwnProperties = make(map[string]*PropertyDefinition)
	for _, pr := range xmlCls.Properties {
		var cpr *PropertyDefinition = nil
		cpr, errs = loadOwnProperty(self, xmlCls, &pr, errs)
		if nil != cpr {
			cls.OwnProperties[cpr.Name] = cpr
		}
	}
	return errs
}

func loadOwnProperty(self *ClassDefinitions, xmlCls *XMLClassDefinition,
	pr *XMLPropertyDefinition, errs []error) (*PropertyDefinition, []error) {

	cpr := &PropertyDefinition{Name: pr.Name,
		Type:         GetTypeDefinition(pr.Restrictions.Type),
		Restrictions: make([]Validator, 0, 4)}

	if "" != pr.Restrictions.DefaultValue {
		var err error
		cpr.DefaultValue, err = cpr.Type.ConvertFrom(pr.Restrictions.DefaultValue)
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
			cpr.Restrictions = append(cpr.Restrictions, validator)
		}
	}
	if "" != pr.Restrictions.Pattern {
		validator, err := cpr.Type.CreatePatternValidator(pr.Restrictions.Pattern)
		if nil != err {
			errs = append(errs, errors.New("load property '"+pr.Name+"' of class '"+
				xmlCls.Name+"' failed, parse Pattern '"+
				pr.Restrictions.Pattern+"' failed, "+err.Error()))
		} else {
			cpr.Restrictions = append(cpr.Restrictions, validator)
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
			cpr.Restrictions = append(cpr.Restrictions, validator)
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
			cpr.Restrictions = append(cpr.Restrictions, validator)
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
			cpr.Restrictions = append(cpr.Restrictions, validator)
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

func loadAssocations(self *ClassDefinitions, cls *ClassDefinition, xmlDefinition *XMLClassDefinition, errs []error) []error {
	if nil != xmlDefinition.BelongsTo && 0 != len(xmlDefinition.BelongsTo) {
		for _, belongs_to := range xmlDefinition.BelongsTo {
			target, ok := self.clsDefinitions[belongs_to.Target]
			if !ok {
				errs = append(errs, errors.New("Target '"+belongs_to.Target+
					"' of belongs_to '"+belongs_to.Name+"' is not found."))
				continue
			}

			pr, ok := cls.OwnProperties[belongs_to.Name]
			if !ok {
				errs = append(errs, errors.New("Property '"+belongs_to.Name+
					"' of belongs_to '"+belongs_to.Name+"' is not found."))
				continue
			}
			if nil == cls.Assocations {
				cls.Assocations = make([]Assocation, 0, 4)
			}

			cls.Assocations = append(cls.Assocations, &BelongsTo{TargetClass: target, Name: pr})
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
			if nil == cls.Assocations {
				cls.Assocations = make([]Assocation, 0, 4)
			}
			cls.Assocations = append(cls.Assocations, &HasMany{TargetClass: target})
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
			if nil == cls.Assocations {
				cls.Assocations = make([]Assocation, 0, 4)
			}
			cls.Assocations = append(cls.Assocations, &HasAndBelongsToMany{TargetClass: target})
		}
	}
	return errs
}

func LoadXml(nm string) (*ClassDefinitions, error) {
	f, err := ioutil.ReadFile(nm)
	if nil != err {
		return nil, fmt.Errorf("read file '%s' failed, %s", nm, err.Error())
	}

	var xml_definitions XMLClassDefinitions
	err = xml.Unmarshal(f, &xml_definitions)
	if nil != err {
		return nil, fmt.Errorf("unmarshal xml '%s' failed, %s", nm, err.Error())
	}

	if nil == xml_definitions.Definitions || 0 == len(xml_definitions.Definitions) {
		return nil, fmt.Errorf("unmarshal xml '%s' error, class definition is empty", nm)
	}

	self := &ClassDefinitions{clsDefinitions: make(map[string]*ClassDefinition, 100)}
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
		errs = loadOwnProperties(self, &xmlDefinition, cls, errs)

		self.clsDefinitions[cls.Name] = cls
	}

	// load super class and own assocations
	for _, xmlDefinition := range xml_definitions.Definitions {
		cls, ok := self.clsDefinitions[xmlDefinition.Name]
		if !ok {
			continue
		}
		if "" != xmlDefinition.Base {
			super, ok := self.clsDefinitions[xmlDefinition.Base]
			if !ok || nil == super {
				errs = append(errs, errors.New("Base '"+xmlDefinition.Base+
					"' of class '"+xmlDefinition.Name+"' is not found."))
			} else {
				cls.Super = super
			}
		}
		errs = loadAssocations(self, cls, &xmlDefinition, errs)
	}

	// load the properties of super class
	for _, cls := range self.clsDefinitions {
		errs = loadParentProperties(self, cls, errs)
	}

	if 0 == len(errs) {
		return self, nil
	}
	return self, &MutiErrors{msg: "load file '" + nm + "' error", errs: errs}
}

func (self *ClassDefinitions) Find(nm string) *ClassDefinition {
	if cls, ok := self.clsDefinitions[nm]; ok {
		return cls
	}
	return nil
}
