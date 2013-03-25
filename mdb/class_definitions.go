package mdb

import (
	"commons"
	"commons/stringutils"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

type ClassDefinitions struct {
	underscore2Definitions map[string]*ClassDefinition
	clsDefinitions         map[string]*ClassDefinition
	id2Definitions         map[int]*ClassDefinition
}

func (self *ClassDefinitions) FindByUnderscoreName(nm string) *ClassDefinition {
	if cls, ok := self.underscore2Definitions[nm]; ok {
		return cls
	}
	return nil
}

func (self *ClassDefinitions) Find(nm string) *ClassDefinition {
	if cls, ok := self.clsDefinitions[nm]; ok {
		return cls
	}
	return nil
}

func checkHierarchicalType(self *ClassDefinitions, cls *ClassDefinition, errs []error) []error {
	if nil == cls.HierarchicalType {
		if nil != cls.Super && nil != cls.Super.HierarchicalType {
			errs = append(errs, errors.New("parent '"+cls.Super.Name+"' is define hierarchical, child '"+
				cls.Name+"' is not defined"))
		}
	} else if nil != cls.Super {
		if nil == cls.Super.HierarchicalType {
			errs = append(errs, errors.New("child '"+cls.Name+"' is define hierarchical and parent '"+
				cls.Super.Name+"' is not defined"))
		} else {

			if cls.HierarchicalType.MinValue < cls.Super.HierarchicalType.MinValue {
				errs = append(errs, errors.New("'minValue' of child '"+cls.Name+
					"' is less than 'minValue' of parent '"+
					cls.Super.Name+"'"))
			}

			if cls.HierarchicalType.MaxValue > cls.Super.HierarchicalType.MaxValue {
				errs = append(errs, errors.New("'maxValue' of child '"+cls.Name+
					"' is greater than 'maxValue' of parent '"+cls.Super.Name+"'"))
			}

			if cls.HierarchicalType.MinValue < cls.Super.HierarchicalType.Value &&
				cls.Super.HierarchicalType.Value > cls.HierarchicalType.MaxValue {
				errs = append(errs, errors.New("'Value' of child '"+cls.Name+
					"' is between with 'minValue' and 'maxValue' of parent '"+cls.Super.Name+"'"))
			}
		}
	}

	if nil != cls.Children {
		for _, c1 := range cls.Children {
			if nil == c1.HierarchicalType {
				continue
			}
			for _, c2 := range cls.Children {
				if c1 == c2 {
					continue
				}
				if nil == c2.HierarchicalType {
					continue
				}

				if c1.HierarchicalType.MinValue < c2.HierarchicalType.MinValue {
					if c1.HierarchicalType.MaxValue >= c2.HierarchicalType.MinValue {
						errs = append(errs, errors.New("hierarchical range of '"+c1.Name+
							"' and '"+c2.Name+"' is overlapping"))
					}
				} else {
					if c1.HierarchicalType.MinValue <= c2.HierarchicalType.MaxValue {
						errs = append(errs, errors.New("hierarchical range of '"+c1.Name+
							"' and '"+c2.Name+"' is overlapping"))
					}
				}
			}
		}
	}
	return errs
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

			if !v.IsRequired {
				v.IsRequired = old.IsRequired
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
		if "type" == pr.Name {
			cls.HierarchicalType, errs = loadOwnHierarchicalType(self, &pr, errs)
		} else {
			var cpr *PropertyDefinition = nil
			cpr, errs = loadOwnProperty(self, xmlCls, &pr, errs)
			if nil != cpr {
				cls.OwnProperties[cpr.Name] = cpr
			}
		}
	}
	return errs
}

func loadOwnHierarchicalType(self *ClassDefinitions, pr *XMLPropertyDefinition, errs []error) (*HierarchicalEnumeration, []error) {
	ok := true
	if integerType.Name() != pr.Restrictions.Type {
		errs = append(errs, errors.New("'type' is not a number - "+pr.Restrictions.Type))
		ok = false
	}
	if "" == pr.Restrictions.MinValue {
		errs = append(errs, errors.New(" 'minValue' of 'type' is empty"))
		ok = false
	}
	if "" == pr.Restrictions.MaxValue {
		errs = append(errs, errors.New(" 'maxValue' of 'type' is empty"))
		ok = false
	}
	if "" == pr.Restrictions.DefaultValue {
		errs = append(errs, errors.New(" 'defaultValue' of 'type' is empty"))
		ok = false
	}

	min, err := strconv.Atoi(pr.Restrictions.MinValue)
	if nil != err {
		if "" == pr.Restrictions.MinValue {
			errs = append(errs, errors.New(" 'minValue' of 'type' is empty"))
		} else {
			errs = append(errs, errors.New(" 'minValue' of 'type' is not a number - "+pr.Restrictions.MinValue))
		}
		ok = false
	}

	max, err := strconv.Atoi(pr.Restrictions.MaxValue)
	if nil != err {
		if "" == pr.Restrictions.MaxValue {
			errs = append(errs, errors.New(" 'maxValue' of 'type' is empty"))
		} else {
			errs = append(errs, errors.New(" 'maxValue' of 'type' is not a number - "+pr.Restrictions.MaxValue))
		}
		ok = false
	}

	dvalue, err := strconv.Atoi(pr.Restrictions.DefaultValue)
	if nil != err {
		if "" == pr.Restrictions.DefaultValue {
			errs = append(errs, errors.New(" 'defaultValue' of 'type' is empty"))
		} else {
			errs = append(errs, errors.New(" 'defaultValue' of 'type' is not a number - "+pr.Restrictions.DefaultValue))
		}
		ok = false
	}

	if min > max {
		errs = append(errs, errors.New(" 'minValue' and 'maxValue' of 'type' is error - "+pr.Restrictions.MinValue+
			" < "+pr.Restrictions.MaxValue))
		ok = false
	}

	if min > dvalue || dvalue > max {
		errs = append(errs, errors.New(" 'defaultValue' of 'type' is error - "+pr.Restrictions.MinValue+
			" < "+pr.Restrictions.DefaultValue+"< "+pr.Restrictions.MaxValue))
		ok = false
	}

	if ok {
		return &HierarchicalEnumeration{Value: dvalue, MinValue: min, MaxValue: max}, errs
	}
	return nil, errs
}

func loadOwnProperty(self *ClassDefinitions, xmlCls *XMLClassDefinition,
	pr *XMLPropertyDefinition, errs []error) (*PropertyDefinition, []error) {

	cpr := &PropertyDefinition{Name: pr.Name,
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
		if _, ok := cpr.Type.(*DateTimeTypeDefinition); !ok {
			errs = append(errs, errors.New("load property '"+pr.Name+"' of class '"+
				xmlCls.Name+"' failed, it must is a dateTime"))
		}

		if COLLECTION_UNKNOWN != cpr.Collection {
			errs = append(errs, errors.New("load property '"+pr.Name+"' of class '"+
				xmlCls.Name+"' failed, it must not is a collection"))
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
			errs = append(errs, errors.New("load property '"+pr.Name+"' of class '"+
				xmlCls.Name+"' failed, collection has not defaultValue "))
		} else {
			var err error
			cpr.DefaultValue, err = cpr.Type.Convert(pr.Restrictions.DefaultValue)
			if nil != err {
				errs = append(errs, errors.New("load property '"+pr.Name+"' of class '"+
					xmlCls.Name+"' failed, parse defaultValue '"+
					pr.Restrictions.DefaultValue+"' failed, "+err.Error()))
			}
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

func makeAssocation(self *ClassDefinitions, cls *ClassDefinition, errs *[]error, t, tName, em, polymorphic_s, fKey, attrName string) Assocation {
	target, ok := self.clsDefinitions[tName]
	if !ok {
		*errs = append(*errs, fmt.Errorf("process %s target '%s' of class '%s' failed, '%s' is not found.",
			t, tName, cls.Name, tName))
		return nil
	}
	if nil != target.Super {
		*errs = append(*errs, fmt.Errorf("process %s target '%s' of class '%s' failed, '%s' is a sub class.",
			t, tName, cls.Name, tName))
		return nil
	}
	polymorphic := false
	if "" == polymorphic_s {
		polymorphic = false
	} else if "true" == polymorphic_s {
		polymorphic = true
	} else {
		*errs = append(*errs, fmt.Errorf("process %s target '%s' of class '%s' failed, attribute 'polymorphic' is unrecorign.",
			t, tName, cls.Name, tName))
		return nil
	}

	embedded := false
	if "" == em {
		embedded = false
	} else if "true" == em {
		embedded = true
	} else {
		*errs = append(*errs, fmt.Errorf("process %s target '%s' of class '%s' failed, attribute 'embedded' is unrecorign.",
			t, tName, cls.Name, tName))
		return nil
	}
	if embedded {
		if "" != fKey {
			*errs = append(*errs, fmt.Errorf("process %s target '%s' of class '%s' failed,  assocations is embedded, "+
				"'foreignKey' must is not present .", t, tName, cls.Name))
			return nil
		}
		if "" == attrName {
			attrName = stringutils.Underscore(cls.Name)
		}
	} else {
		if "" != attrName {
			*errs = append(*errs, fmt.Errorf("process %s target '%s' of class '%s' failed,  assocations is embedded, "+
				" 'attributeName' must is not present .", t, tName, cls.Name))
			return nil
		}
		if "has_many" == t && polymorphic {
			if "" != fKey {
				*errs = append(*errs, fmt.Errorf("process %s target '%s' of class '%s' failed,  assocations is polymorphic, "+
					" 'foreignKey' must is not present .", t, tName, cls.Name))
				return nil
			}
			fKey = "parent_id"

			pr, ok := target.OwnProperties["parent_id"]
			if !ok {
				pr = &PropertyDefinition{Name: "parent_id", Type: &objectIdType, Collection: COLLECTION_UNKNOWN}
				target.OwnProperties["parent_id"] = pr
			} else {
				if _, ok := pr.Type.(*ObjectIdTypeDefinition); !ok {
					*errs = append(*errs, fmt.Errorf("process %s target '%s' of class '%s' failed, 'parent_id' is not objectId type",
						t, tName, cls.Name))
					return nil
				}
			}
			pr, ok = target.OwnProperties["parent_type"]
			if !ok {
				pr = &PropertyDefinition{Name: "parent_type", Type: &stringType, Collection: COLLECTION_UNKNOWN}
				target.OwnProperties["parent_type"] = pr
			} else {
				if _, ok := pr.Type.(*StringTypeDefinition); !ok {
					*errs = append(*errs, fmt.Errorf("process %s target '%s' of class '%s' failed, 'parent_type' is not string type",
						t, tName, cls.Name))
					return nil
				}
				if COLLECTION_UNKNOWN != pr.Collection {
					*errs = append(*errs, fmt.Errorf("process %s target '%s' of class '%s' failed, 'parent_type' is array or set",
						t, tName, cls.Name))
					return nil
				}
			}
		} else {
			if "" == fKey {
				fKey = stringutils.Underscore(cls.Name) + "_id"
			}
			pr, ok := target.OwnProperties[fKey]
			if !ok {
				pr = &PropertyDefinition{Name: fKey, Type: &objectIdType, Collection: COLLECTION_UNKNOWN}
				target.OwnProperties[fKey] = pr
			} else {
				if _, ok := pr.Type.(*ObjectIdTypeDefinition); !ok {
					*errs = append(*errs, fmt.Errorf("process %s target '%s' of class '%s' failed, foreignKey is not objectId type",
						t, tName, cls.Name))
					return nil
				}
			}
		}
	}
	if "has_many" == t {
		return &HasMany{TargetClass: target, Embedded: embedded, Polymorphic: polymorphic,
			AttributeName: attrName, ForeignKey: fKey}
	}
	return &HasOne{TargetClass: target, Embedded: embedded,
		AttributeName: attrName, ForeignKey: fKey}
}

func loadAssocations(self *ClassDefinitions, cls *ClassDefinition, xmlDefinition *XMLClassDefinition, errs []error) []error {
	if nil != xmlDefinition.BelongsTo && 0 != len(xmlDefinition.BelongsTo) {
		for _, belongs_to := range xmlDefinition.BelongsTo {
			target, ok := self.clsDefinitions[belongs_to.Target]
			if !ok {
				errs = append(errs, errors.New("belongs_to Target '"+belongs_to.Target+
					"' of class '"+xmlDefinition.Name+"' is not found."))
				continue
			}
			if "" == belongs_to.Name {
				belongs_to.Name = stringutils.Underscore(belongs_to.Target) + "_id"
			}

			pr, ok := cls.OwnProperties[belongs_to.Name]
			if !ok {
				errs = append(errs, errors.New("Property '"+belongs_to.Name+
					"' of belongs_to '"+belongs_to.Target+"' is not found."))
				continue
			}
			if nil == cls.Assocations {
				cls.Assocations = make([]Assocation, 0, 4)
			}

			cls.Assocations = append(cls.Assocations, &BelongsTo{TargetClass: target, Name: pr})
		}
	}
	if nil != xmlDefinition.HasMany && 0 != len(xmlDefinition.HasMany) {
		for _, hasMany := range xmlDefinition.HasMany {
			ass := makeAssocation(self, cls, &errs, "has_many", hasMany.Target, hasMany.Embedded,
				hasMany.Polymorphic, hasMany.ForeignKey, hasMany.AttributeName)
			if nil == ass {
				continue
			}
			if nil == cls.Assocations {
				cls.Assocations = make([]Assocation, 0, 4)
			}
			cls.Assocations = append(cls.Assocations, ass)
		}
	}
	if nil != xmlDefinition.HasOne && 0 != len(xmlDefinition.HasOne) {
		for _, hasOne := range xmlDefinition.HasOne {
			ass := makeAssocation(self, cls, &errs, "has_one", hasOne.Target, hasOne.Embedded,
				"", hasOne.ForeignKey, hasOne.AttributeName)
			if nil == ass {
				continue
			}
			if nil == cls.Assocations {
				cls.Assocations = make([]Assocation, 0, 4)
			}
			cls.Assocations = append(cls.Assocations, ass)
		}
	}
	if nil != xmlDefinition.HasAndBelongsToMany && 0 != len(xmlDefinition.HasAndBelongsToMany) {
		for _, habtm := range xmlDefinition.HasAndBelongsToMany {
			target, ok := self.clsDefinitions[habtm.Target]
			if !ok {
				errs = append(errs, errors.New("Target '"+habtm.Target+
					"' of has_and_belongs_to_many is not found."))
				continue
			}
			if nil == cls.Assocations {
				cls.Assocations = make([]Assocation, 0, 4)
			}
			foreignKey := habtm.ForeignKey
			if "" == foreignKey {
				foreignKey = stringutils.Underscore(cls.Name) + "_id"
			}
			var collectionName string
			through, ok := self.clsDefinitions[habtm.Through]
			if !ok {
				t1 := stringutils.Tableize(target.Name)
				t2 := stringutils.Tableize(cls.Name)

				if t1[0] > t2[0] {
					collectionName = t1 + "_" + t2
				} else {
					collectionName = t2 + "_" + t1
				}
			} else {
				through = nil
				collectionName = stringutils.Tableize(habtm.Through)
			}
			cls.Assocations = append(cls.Assocations, &HasAndBelongsToMany{TargetClass: target,
				Through: through, CollectionName: collectionName, ForeignKey: foreignKey})
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

	self := &ClassDefinitions{clsDefinitions: make(map[string]*ClassDefinition, 100),
		underscore2Definitions: make(map[string]*ClassDefinition, 100)}
	errs := make([]error, 0, 10)

	// load class definitions and own properties
	for _, xmlDefinition := range xml_definitions.Definitions {
		_, ok := self.clsDefinitions[xmlDefinition.Name]
		if ok {
			errs = append(errs, errors.New("class '"+xmlDefinition.Name+
				"' is aleady exists."))
			continue
		}

		cls := &ClassDefinition{Name: xmlDefinition.Name,
			UnderscoreName: stringutils.Underscore(xmlDefinition.Name),
			collectionName: stringutils.Tableize(xmlDefinition.Name)}
		errs = loadOwnProperties(self, &xmlDefinition, cls, errs)

		self.clsDefinitions[cls.Name] = cls
		self.underscore2Definitions[cls.UnderscoreName] = cls
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

				if nil == super.Children {
					super.Children = make([]*ClassDefinition, 0, 3)
				}
				super.Children = append(super.Children, cls)
			}
		}
		errs = loadAssocations(self, cls, &xmlDefinition, errs)
	}

	// load the properties of super class
	for _, cls := range self.clsDefinitions {
		errs = loadParentProperties(self, cls, errs)
	}

	// check hierarchical of type
	for _, cls := range self.clsDefinitions {
		errs = checkHierarchicalType(self, cls, errs)
	}

	if 0 == len(errs) {
		return self, nil
	}
	return self, commons.NewMutiErrors("load file '"+nm+"' error", errs)
}

// func LoadHierarchyFromXml(self *ClassDefinitions, nm string) error {
// 	f, err := ioutil.ReadFile(nm)
// 	if nil != err {
// 		return nil, fmt.Errorf("read file '%s' failed, %s", nm, err.Error())
// 	}

// 	var xml_definitions XMLClassIdentifiers
// 	err = xml.Unmarshal(f, &xml_definitions)
// 	if nil != err {
// 		return nil, fmt.Errorf("unmarshal xml '%s' failed, %s", nm, err.Error())
// 	}

// 	if nil == xml_definitions.Definitions || 0 == len(xml_definitions.Definitions) {
// 		return nil, fmt.Errorf("unmarshal xml '%s' error, class definition is empty", nm)
// 	}

// 	self.id2Definitions = make(map[int]*ClassDefinition, 100)}
// 	errs := make([]error, 0, 10)

// 	for _, xmlDefinition := range xml_definitions.Definitions {

// 	}

// 	if 0 == len(errs) {
// 		return nil
// 	}
// 	return &MutiErrors{msg: "load file '" + nm + "' error", errs: errs}
// }
