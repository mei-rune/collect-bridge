package types

import (
	"commons/stringutils"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

func makeIdColumn() *ColumnDefinition {
	return &ColumnDefinition{AttributeDefinition{Name: "id", Type: GetTypeDefinition("objectId"),
		Collection: COLLECTION_UNKNOWN}}
}

func loadParentColumns(self *TableDefinitions, cls *TableDefinition, errs *[]string) {
	if nil != cls.Attributes {
		return
	}

	cls.Attributes = make(map[string]*ColumnDefinition, 2*len(cls.OwnAttributes))
	if nil != cls.Super {
		loadParentColumns(self, cls.Super, errs)
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

func loadOwnColumns(xmlCls *XMLClassDefinition, cls *TableDefinition) (errs []string) {
	cls.OwnAttributes = make(map[string]*ColumnDefinition)
	for _, pr := range xmlCls.Attributes {
		if "type" == pr.Name {
			errs = append(errs, "load column '"+pr.Name+"' of class '"+
				xmlCls.Name+"' failed, it is reserved")
			continue
		}

		if "id" == pr.Name {
			errs = append(errs, "load column '"+pr.Name+"' of class '"+
				xmlCls.Name+"' failed, it is reserved")
			continue
		}

		var cpr *AttributeDefinition = nil
		cpr, msgs := loadOwnAttribute(&pr)
		if nil != cpr {
			cls.OwnAttributes[cpr.Name] = &ColumnDefinition{*cpr}
		}

		errs = mergeErrors(errs, "load column '"+pr.Name+"' of class '"+
			xmlCls.Name+"' failed", msgs)
	}
	return errs
}

func makeAssocation(self *TableDefinitions, cls *TableDefinition,
	t, tName, polymorphic, fKey string) (Assocation, error) {

	target, ok := self.definitions[tName]
	if !ok {
		return nil, errors.New("'" + tName + "' is not found.")
	}
	// if nil != target.Super {
	// 	*errs = append(*errs, "'"+tName+"' is a sub class.")
	// 	return nil
	// }
	is_polymorphic := false
	if "" == polymorphic {
		is_polymorphic = false
	} else if "true" == polymorphic {
		is_polymorphic = true
	} else {
		return nil, errors.New("attribute 'polymorphic' is unrecorign.")
	}

	if is_polymorphic {
		if "" != fKey {
			return nil, errors.New("'foreignKey' must is not present .")
		}
		fKey = "parent_id"
		pr, ok := target.OwnAttributes["parent_id"]
		if !ok {
			pr = &ColumnDefinition{AttributeDefinition{Name: "parent_id", Type: GetTypeDefinition("objectId"),
				Collection: COLLECTION_UNKNOWN}}

			target.OwnAttributes["parent_id"] = pr
		} else {
			if "objectId" != pr.Type.Name() {
				return nil, errors.New("'parent_id' is not objectId type")
			}
		}

		pr, ok = target.OwnAttributes["parent_type"]
		if !ok {
			pr = &ColumnDefinition{AttributeDefinition{Name: "parent_type", Type: GetTypeDefinition("string"),
				Collection: COLLECTION_UNKNOWN}}
			target.OwnAttributes["parent_type"] = pr
		} else {
			if "string" != pr.Type.Name() {
				return nil, errors.New("'parent_type' is reserved and must is a string type")
			}
			if pr.Collection.IsCollection() {
				return nil, errors.New("'parent_type' is reserved and is a collection")
			}
		}
	} else {
		if "" == fKey {
			fKey = stringutils.Underscore(cls.Name) + "_id"
		}
		pr, ok := target.OwnAttributes[fKey]
		if !ok {
			pr = &ColumnDefinition{AttributeDefinition{Name: fKey, Type: GetTypeDefinition("objectId"),
				Collection: COLLECTION_UNKNOWN}}
			target.OwnAttributes[fKey] = pr
		} else {
			if "objectId" != pr.Type.Name() {
				return nil, errors.New("'foreignKey' is not objectId type")
			}
			if pr.Collection.IsCollection() {
				return nil, errors.New("'foreignKey' is a collection")
			}
		}
	}
	if "has_many" == t {
		return &HasMany{TargetTable: target, Polymorphic: is_polymorphic, ForeignKey: fKey}, nil
	}
	return &HasOne{TargetTable: target, Polymorphic: is_polymorphic, ForeignKey: fKey}, nil
}

func loadAssocations(self *TableDefinitions, cls *TableDefinition, xmlDefinition *XMLClassDefinition, errs *[]string) {
	if nil != xmlDefinition.BelongsTo && 0 != len(xmlDefinition.BelongsTo) {
		for _, belongs_to := range xmlDefinition.BelongsTo {
			target, ok := self.definitions[belongs_to.Target]
			if !ok {
				*errs = append(*errs, "belongs_to Target '"+belongs_to.Target+
					"' of class '"+xmlDefinition.Name+"' is not found.")
				continue
			}
			if "" == belongs_to.Name {
				belongs_to.Name = stringutils.Underscore(belongs_to.Target) + "_id"
			}

			pr, ok := cls.OwnAttributes[belongs_to.Name]
			if !ok {
				*errs = append(*errs, "Property '"+belongs_to.Name+
					"' of belongs_to '"+belongs_to.Target+"' is not found.")
				continue
			}

			cls.Assocations = append(cls.Assocations, &BelongsTo{TargetTable: target, Name: pr})
		}
	}
	if nil != xmlDefinition.HasMany && 0 != len(xmlDefinition.HasMany) {
		for _, hasMany := range xmlDefinition.HasMany {
			ass, err := makeAssocation(self, cls, "has_many", hasMany.Target,
				hasMany.Polymorphic, hasMany.ForeignKey)
			if nil != err {
				*errs = append(*errs, "load has_many '"+hasMany.Target+"' failed, "+err.Error())
			}
			if nil == ass {
				continue
			}

			cls.Assocations = append(cls.Assocations, ass)
		}
	}
	if nil != xmlDefinition.HasOne && 0 != len(xmlDefinition.HasOne) {
		for _, hasOne := range xmlDefinition.HasOne {
			ass, err := makeAssocation(self, cls, "has_one", hasOne.Target,
				"", hasOne.ForeignKey)
			if nil != err {
				*errs = append(*errs, "load has_one '"+hasOne.Target+"' failed, "+err.Error())
			}
			if nil == ass {
				continue
			}

			cls.Assocations = append(cls.Assocations, ass)
		}
	}
	if nil != xmlDefinition.HasAndBelongsToMany && 0 != len(xmlDefinition.HasAndBelongsToMany) {
		for _, habtm := range xmlDefinition.HasAndBelongsToMany {
			target, ok := self.definitions[habtm.Target]
			if !ok {
				*errs = append(*errs, "Target '"+habtm.Target+
					"' of has_and_belongs_to_many is not found.")
				continue
			}

			foreignKey := habtm.ForeignKey
			if "" == foreignKey {
				foreignKey = stringutils.Underscore(cls.Name) + "_id"
			}

			through, ok := self.definitions[habtm.Through]
			if !ok {
				*errs = append(*errs, "Through '"+habtm.Through+
					"' of has_and_belongs_to_many is not found.")
				continue
			}

			cls.Assocations = append(cls.Assocations, &HasAndBelongsToMany{TargetTable: target,
				Through: through, ForeignKey: foreignKey})
		}
	}
}

func LoadTableDefinitions(nm string) (*TableDefinitions, error) {
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
		return nil, fmt.Errorf("unmarshal xml '%s' error, definitions is empty", nm)
	}

	self := &TableDefinitions{definitions: make(map[string]*TableDefinition, 100),
		underscore2Definitions: make(map[string]*TableDefinition, 100)}
	errs := make([]string, 0, 10)

	// load table definitions and own properties
	for _, xmlDefinition := range xml_definitions.Definitions {
		_, ok := self.definitions[xmlDefinition.Name]
		if ok {
			errs = append(errs, "table '"+xmlDefinition.Name+
				"' is aleady exists.")
			continue
		}

		cls := &TableDefinition{Name: xmlDefinition.Name,
			UnderscoreName: stringutils.Underscore(xmlDefinition.Name),
			CollectionName: stringutils.Tableize(xmlDefinition.Name)}
		msgs := loadOwnColumns(&xmlDefinition, cls)
		if nil != msgs && 0 == len(msgs) {
			errs = mergeErrors(errs, "", msgs)
		}

		self.definitions[cls.Name] = cls
		self.underscore2Definitions[cls.UnderscoreName] = cls
	}

	// load super class
	for _, xmlDefinition := range xml_definitions.Definitions {
		cls, ok := self.definitions[xmlDefinition.Name]
		if !ok {
			continue
		}
		if "" == xmlDefinition.Base {
			continue
		}
		super, ok := self.definitions[xmlDefinition.Base]
		if !ok || nil == super {
			errs = append(errs, "Base '"+xmlDefinition.Base+
				"' of class '"+xmlDefinition.Name+"' is not found.")
		} else {
			cls.Super = super

			if nil == super.Children {
				super.Children = make([]*TableDefinition, 0, 3)
			}
			super.Children = append(super.Children, cls)
		}
	}

	// load own assocations
	for _, xmlDefinition := range xml_definitions.Definitions {
		cls, ok := self.definitions[xmlDefinition.Name]
		if !ok {
			continue
		}

		loadAssocations(self, cls, &xmlDefinition, &errs)
	}

	// load the properties of super class
	for _, cls := range self.definitions {
		if nil != cls.Super {
			continue
		}
		cls.Id = makeIdColumn()
		cls.OwnAttributes[cls.Id.Name] = cls.Id
	}

	// load the properties of super class
	for _, cls := range self.definitions {
		loadParentColumns(self, cls, &errs)
	}

	// change collection name
	// for _, cls := range self.definitions {
	// 	SetCollectionName(self, cls, &errs)
	// }

	// // check hierarchical of type
	// for _, cls := range self.definitions {
	// 	errs = checkHierarchicalType(self, cls, errs)
	// }

	if 0 == len(errs) {
		return self, nil
	}
	errs = mergeErrors(nil, "load file '"+nm+"' error:", errs)
	return self, errors.New(strings.Join(errs, "\r\n"))
}
