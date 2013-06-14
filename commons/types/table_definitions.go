package types

import (
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

func loadParentColumns(definitions map[string]*TableDefinition, cls *TableDefinition, errs *[]string) {
	if nil != cls.Attributes {
		return
	}

	cls.Attributes = make(map[string]*ColumnDefinition, 2*len(cls.OwnAttributes))
	if nil != cls.Super {
		loadParentColumns(definitions, cls.Super, errs)
		for k, v := range cls.Super.Attributes {
			cls.Attributes[k] = v
		}

		cls.Id = cls.Super.Id
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

		if "id" == pr.Name {
			errs = append(errs, "load column '"+pr.Name+"' of class '"+
				xmlCls.Name+"' failed, it is reserved")
			continue
		}

		var cpr *AttributeDefinition = nil
		cpr, msgs := loadOwnAttribute(&pr)
		if nil != cpr {
			switch cpr.Name {
			case "type":
				if "string" != cpr.Type.Name() {
					errs = append(errs, "load column 'type' of class '"+
						xmlCls.Name+"' failed, it is reserved and must is a string")
					continue
				}

				if cpr.Collection.IsCollection() {
					errs = append(errs, "load column 'type' of class '"+
						xmlCls.Name+"' failed, it is reserved and must not is a collection")
					continue
				}
			case "record_version":
				if "integer" != cpr.Type.Name() {
					errs = append(errs, "load column 'record_version' of class '"+
						xmlCls.Name+"' failed, it is reserved and must is a integer")
					continue
				}

				if cpr.Collection.IsCollection() {
					errs = append(errs, "load column 'record_version' of class '"+
						xmlCls.Name+"' failed, it is reserved and must not is a collection")
					continue
				}

				errs = append(errs, "load column 'record_version' of class '"+
					xmlCls.Name+"' failed, it is reserved")
			}

			cls.OwnAttributes[cpr.Name] = &ColumnDefinition{*cpr}
		}

		errs = mergeErrors(errs, "load column '"+pr.Name+"' of class '"+
			xmlCls.Name+"' failed", msgs)
	}
	return errs
}

func makeAssocation(definitions map[string]*TableDefinition, cls *TableDefinition,
	t, tName, polymorphic, fKey string) (Assocation, error) {

	target, ok := definitions[tName]
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
			fKey = Underscore(cls.Name) + "_id"
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

func loadAssocations(definitions map[string]*TableDefinition, cls *TableDefinition, xmlDefinition *XMLClassDefinition, errs *[]string) {
	if nil != xmlDefinition.BelongsTo && 0 != len(xmlDefinition.BelongsTo) {
		for _, belongs_to := range xmlDefinition.BelongsTo {
			target, ok := definitions[belongs_to.Target]
			if !ok {
				*errs = append(*errs, "belongs_to Target '"+belongs_to.Target+
					"' of class '"+xmlDefinition.Name+"' is not found.")
				continue
			}

			if "" == belongs_to.Name {
				belongs_to.Name = Underscore(belongs_to.Target) + "_id"
			}

			pr, ok := cls.OwnAttributes[belongs_to.Name]
			if !ok {
				pr = &ColumnDefinition{AttributeDefinition{Name: belongs_to.Name, Type: GetTypeDefinition("objectId"),
					Collection: COLLECTION_UNKNOWN}}
				cls.OwnAttributes[belongs_to.Name] = pr
			}

			cls.Assocations = append(cls.Assocations, &BelongsTo{TargetTable: target, Name: pr})
		}
	}
	if nil != xmlDefinition.HasMany && 0 != len(xmlDefinition.HasMany) {
		for _, hasMany := range xmlDefinition.HasMany {
			ass, err := makeAssocation(definitions, cls, "has_many", hasMany.Target,
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
			ass, err := makeAssocation(definitions, cls, "has_one", hasOne.Target,
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
			target, ok := definitions[habtm.Target]
			if !ok {
				*errs = append(*errs, "Target '"+habtm.Target+
					"' of has_and_belongs_to_many is not found.")
				continue
			}

			foreignKey := habtm.ForeignKey
			if "" == foreignKey {
				foreignKey = Underscore(cls.Name) + "_id"
			}

			through, ok := definitions[habtm.Through]
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

	definitions := make(map[string]*TableDefinition)
	errs := make([]string, 0, 10)

	// load table definitions and own properties
	for _, xmlDefinition := range xml_definitions.Definitions {
		_, ok := definitions[xmlDefinition.Name]
		if ok {
			errs = append(errs, "table '"+xmlDefinition.Name+
				"' is aleady exists.")
			continue
		}

		cls := &TableDefinition{Name: xmlDefinition.Name,
			UnderscoreName: Underscore(xmlDefinition.Name),
			CollectionName: Tableize(xmlDefinition.Name)}
		msgs := loadOwnColumns(&xmlDefinition, cls)
		if nil != msgs && 0 != len(msgs) {
			errs = mergeErrors(errs, "", msgs)
		}

		definitions[cls.Name] = cls
	}

	// load super class
	for _, xmlDefinition := range xml_definitions.Definitions {
		cls, ok := definitions[xmlDefinition.Name]
		if !ok {
			continue
		}
		if "" == xmlDefinition.Base {
			continue
		}
		super, ok := definitions[xmlDefinition.Base]
		if !ok || nil == super {
			errs = append(errs, "Base '"+xmlDefinition.Base+
				"' of class '"+xmlDefinition.Name+"' is not found.")
		} else {
			cls.Super = super
		}
	}

	// load own assocations
	for _, xmlDefinition := range xml_definitions.Definitions {
		cls, ok := definitions[xmlDefinition.Name]
		if !ok {
			continue
		}

		loadAssocations(definitions, cls, &xmlDefinition, &errs)
	}

	// load the properties of super class
	for _, cls := range definitions {
		if nil != cls.Super {
			continue
		}
		cls.Id = makeIdColumn()
		cls.OwnAttributes[cls.Id.Name] = cls.Id
	}

	// load the properties of super class
	for _, cls := range definitions {
		loadParentColumns(definitions, cls, &errs)
	}

	// reset collection name
	for _, cls := range definitions {
		if !cls.IsSingleTableInheritance() {
			for s := cls.Super; nil != s; s = s.Super {
				if s.IsSingleTableInheritance() {
					errs = append(errs, "'"+cls.Name+"' is not simple table inheritance, but parent table '"+s.Name+"' is simple table inheritance")
					break
				}
			}

			//fmt.Printf("%v --> not sti\r\n %v\r\n\r\n", cls.Name, cls.String())
			continue
		}

		last := cls.CollectionName

		for s := cls.Super; nil != s; s = s.Super {
			if !s.IsSingleTableInheritance() {
				break
			}
			last = s.CollectionName
		}
		//fmt.Printf("%v --> %v\r\n", cls.Name, cls.CollectionName)
		cls.CollectionName = last
	}

	// load the properties of super class
	for _, cls := range definitions {
		if nil == cls.Super {
			continue
		}

		if nil == cls.Super.OwnChildren {
			cls.Super.OwnChildren = NewTableDefinitions()
		}
		cls.Super.OwnChildren.Register(cls)

		for s := cls.Super; nil != s; s = s.Super {

			if nil == s.Children {
				s.Children = NewTableDefinitions()
			}

			s.Children.Register(cls)
		}
	}

	// check id is exists.
	for _, cls := range definitions {
		if ok := cls.GetAttribute("id"); nil == ok {
			errs = append(errs, "'"+cls.Name+"' has not 'id'")
		}

		//fmt.Println(cls.Name, cls.UnderscoreName, cls.CollectionName)
	}

	// change collection name
	// for _, cls := range definitions {
	// 	SetCollectionName(self, cls, &errs)
	// }

	// // check hierarchical of type
	// for _, cls := range definitions {
	// 	errs = checkHierarchicalType(self, cls, errs)
	// }

	if 0 != len(errs) {
		errs = mergeErrors(nil, "load file '"+nm+"' error:", errs)
		return nil, errors.New(strings.Join(errs, "\r\n"))
	}

	res := NewTableDefinitions()
	for _, cls := range definitions {
		res.Register(cls)
	}
	return res, nil
}
