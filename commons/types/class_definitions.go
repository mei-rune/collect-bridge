package types

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

func loadParentAttributes(self *ClassDefinitions, cls *ClassDefinition, errs *[]string) {
	if nil != cls.Attributes {
		return
	}

	cls.Attributes = make(map[string]*AttributeDefinition, 2*len(cls.OwnAttributes))
	if nil != cls.Super {
		loadParentAttributes(self, cls.Super, errs)
		for k, v := range cls.Super.Attributes {
			cls.Attributes[k] = v
		}
	}

	for k, v := range cls.OwnAttributes {
		old, ok := cls.Attributes[k]
		if ok {
			if v.Type != old.Type {
				*errs = append(*errs, "The property with '"+k+
					"' override failed, type is not same, own is '"+
					v.Type.Name()+"', super is '"+old.Type.Name()+"'")
			}

			// merge restrictions
			if nil != old.Restrictions {
				if nil == v.Restrictions {
					v.Restrictions = make([]Validator, 0)
				}

				for _, r := range old.Restrictions {
					v.Restrictions = append(v.Restrictions, r)
				}
			}

			// merge defaultValue
			if nil == v.DefaultValue {
				v.DefaultValue = old.DefaultValue
			}

			// merge isRequired
			if !v.IsRequired {
				v.IsRequired = old.IsRequired
			}
		}
		cls.Attributes[k] = v
	}
}

func mergeErrors(errs []string, title string, msgs []string) []string {
	if nil == msgs || 0 == len(msgs) {
		return errs
	}
	if "" == title {
		for _, msg := range msgs {
			errs = append(errs, msg)
		}
		return errs
	}

	errs = append(errs, title)
	for _, msg := range msgs {
		errs = append(errs, "    "+msg)
	}
	return errs
}

func loadOwnAttributes(xmlCls *XMLClassDefinition,
	cls *ClassDefinition) (errs []string) {
	cls.OwnAttributes = make(map[string]*AttributeDefinition)
	for _, pr := range xmlCls.Attributes {
		if "type" == pr.Name {
			errs = append(errs, "load property '"+pr.Name+"' of class '"+
				xmlCls.Name+"' failed, it is reserved")
			continue
		}

		var cpr *AttributeDefinition = nil
		cpr, msgs := loadOwnAttribute(&pr)
		if nil != cpr {
			cls.OwnAttributes[cpr.Name] = cpr
		}

		errs = mergeErrors(errs, "load property '"+pr.Name+"' of class '"+
			xmlCls.Name+"' failed", msgs)
	}
	return errs
}

func loadOwnAttribute(pr *XMLAttributeDefinition) (cpr *AttributeDefinition, errs []string) {

	cpr = &AttributeDefinition{Name: pr.Name,
		IsRequired:   false,
		Type:         GetTypeDefinition(pr.Restrictions.Type),
		Restrictions: make([]Validator, 0, 4)}

	switch pr.Restrictions.Collection {
	case "array":
		cpr.Collection = COLLECTION_ARRAY
	case "set":
		cpr.Collection = COLLECTION_SET
	default:
		cpr.Collection = COLLECTION_UNKNOWN
	}

	if "created_at" == pr.Name || "updated_at" == pr.Name {
		if "datetime" != cpr.Type.Name() {
			fmt.Println("-->", cpr.Type.Name())
			errs = append(errs, "it is reserved and must is a datetime")
		}

		if cpr.Collection.IsCollection() {
			errs = append(errs, "it is reserved and must not is a collection")
		}
	}

	if nil != pr.Restrictions.Required {
		cpr.IsRequired = true
	}

	if nil != pr.Restrictions.ReadOnly {
		cpr.IsReadOnly = true
	}

	if nil != pr.Restrictions.Unique {
		cpr.IsUniquely = true
	}

	if "" != pr.Restrictions.DefaultValue {
		if COLLECTION_UNKNOWN != cpr.Collection {
			errs = append(errs, "collection has not defaultValue ")
		} else {
			var err error
			cpr.DefaultValue, err = cpr.Type.ToInternal(pr.Restrictions.DefaultValue)
			if nil != err {
				errs = append(errs, "parse defaultValue '"+
					pr.Restrictions.DefaultValue+"' failed, "+err.Error())
			}
		}
	}

	if nil != pr.Restrictions.Enumerations && 0 != len(*pr.Restrictions.Enumerations) {
		validator, err := cpr.Type.CreateEnumerationValidator(*pr.Restrictions.Enumerations)
		if nil != err {
			errs = append(errs, "parse Enumerations '"+
				strings.Join(*pr.Restrictions.Enumerations, ",")+"' failed, "+err.Error())
		} else {
			cpr.Restrictions = append(cpr.Restrictions, validator)
		}
	}
	if "" != pr.Restrictions.Pattern {
		validator, err := cpr.Type.CreatePatternValidator(pr.Restrictions.Pattern)
		if nil != err {
			errs = append(errs, "parse Pattern '"+
				pr.Restrictions.Pattern+"' failed, "+err.Error())
		} else {
			cpr.Restrictions = append(cpr.Restrictions, validator)
		}
	}
	if "" != pr.Restrictions.MinValue || "" != pr.Restrictions.MaxValue {
		validator, err := cpr.Type.CreateRangeValidator(pr.Restrictions.MinValue,
			pr.Restrictions.MaxValue)
		if nil != err {
			errs = append(errs, "parse Range of Value '"+
				pr.Restrictions.MinValue+","+pr.Restrictions.MaxValue+
				"' failed, "+err.Error())
		} else {
			cpr.Restrictions = append(cpr.Restrictions, validator)
		}
	}
	if "" != pr.Restrictions.Length {
		validator, err := cpr.Type.CreateLengthValidator(pr.Restrictions.Length,
			pr.Restrictions.Length)
		if nil != err {
			errs = append(errs, "parse Length '"+
				pr.Restrictions.Length+"' failed, "+err.Error())
		} else {
			cpr.Restrictions = append(cpr.Restrictions, validator)
		}
	}
	if "" != pr.Restrictions.MinLength || "" != pr.Restrictions.MaxLength {
		validator, err := cpr.Type.CreateLengthValidator(pr.Restrictions.MinLength,
			pr.Restrictions.MaxLength)
		if nil != err {
			errs = append(errs, "parse Range of Length '"+
				pr.Restrictions.MinLength+","+pr.Restrictions.MaxLength+
				"' failed, "+err.Error())
		} else {
			cpr.Restrictions = append(cpr.Restrictions, validator)
		}
	}

	if nil == cpr.Type {
		errs = append(errs, "'"+pr.Restrictions.Type+
			"' is unsupported type")
		return nil, errs
	}

	if nil == errs || 0 == len(errs) {
		return cpr, nil
	}
	return nil, errs
}

func LoadClassDefinitions(nm string) (*ClassDefinitions, error) {
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

	self := &ClassDefinitions{clsDefinitions: make(map[string]*ClassDefinition, 100),
		underscore2Definitions: make(map[string]*ClassDefinition, 100)}
	errs := make([]string, 0, 10)

	// load class definitions and own properties
	for _, xmlDefinition := range xml_definitions.Definitions {
		_, ok := self.clsDefinitions[xmlDefinition.Name]
		if ok {
			errs = append(errs, "class '"+xmlDefinition.Name+
				"' is duplicated.")
			continue
		}

		cls := &ClassDefinition{Name: xmlDefinition.Name,
			UnderscoreName: Underscore(xmlDefinition.Name)}

		msgs := loadOwnAttributes(&xmlDefinition, cls)
		switch xmlDefinition.Abstract {
		case "true", "":
			cls.IsAbstract = true
		case "false":
			cls.IsAbstract = false
		default:
			msgs = append(msgs, "'abstract' value is invalid, it must is 'true' or 'false', actual is '"+xmlDefinition.Abstract+"'")
		}
		errs = mergeErrors(errs, "load class '"+xmlDefinition.Name+"' failed", msgs)

		self.Register(cls)
	}

	// load super class
	for _, xmlDefinition := range xml_definitions.Definitions {
		cls, ok := self.clsDefinitions[xmlDefinition.Name]
		if !ok {
			continue
		}
		if "" == xmlDefinition.Base {
			continue
		}
		super, ok := self.clsDefinitions[xmlDefinition.Base]
		if !ok || nil == super {
			errs = append(errs, "Base '"+xmlDefinition.Base+
				"' of class '"+xmlDefinition.Name+"' is not found.")
			continue
		}

		cls.Super = super
		if nil == super.OwnChildren {
			super.OwnChildren = make([]*ClassDefinition, 0, 3)
		}
		super.OwnChildren = append(super.OwnChildren, cls)
	}

	// load the properties of super class
	for _, cls := range self.clsDefinitions {
		loadParentAttributes(self, cls, &errs)
	}

	if 0 == len(errs) {
		return self, nil
	}
	errs = mergeErrors(nil, "load file '"+nm+"' error:", errs)
	return self, errors.New(strings.Join(errs, "\r\n"))
}
