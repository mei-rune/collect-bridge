package mdb

import (
	"commons"
	"commons/errutils"
	"commons/stringutils"
	"errors"
	"fmt"
	"labix.org/v2/mgo/bson"
	"time"
)

var ChildrenUpdateError = errors.New("don`t save children while update object.")

func checkValue(pr *PropertyDefinition, attributes map[string]interface{}, value interface{}, errs *[]error) (interface{}, bool) {
	new_value, err := pr.Type.Convert(value)
	if nil != err {
		*errs = append(*errs, errors.New("'"+pr.Name+"' convert to internal value failed, "+err.Error()))
		return nil, true
	}

	if nil != pr.Restrictions && 0 != len(pr.Restrictions) {
		is_failed := false
		for _, r := range pr.Restrictions {
			if ok, err := r.Validate(new_value, attributes); !ok {
				*errs = append(*errs, errors.New("'"+pr.Name+"' is validate failed, "+err.Error()))
				is_failed = true
			}
		}

		if is_failed {
			return nil, true
		}
	}
	return new_value, false
}

func doCreatedAtAndUpdatedAt(k string, attributes, new_attributes map[string]interface{},
	is_update bool, errs *[]error) bool {
	if k == "updated_at" {
		_, ok := attributes[k]
		if ok {
			*errs = append(*errs, errors.New("'"+k+"' is magic property"))
			return true
		}

		new_attributes[k] = time.Now()
		return true
	}

	if k == "created_at" {
		_, ok := attributes[k]
		if ok {
			*errs = append(*errs, errors.New("'"+k+"' is magic property"))
			return true
		}

		if !is_update {
			new_attributes[k] = time.Now()
		}
		return true
	}
	return false
}

func (self *mdb_server) doInheritance(cls *ClassDefinition, attributes, new_attributes map[string]interface{},
	is_update bool, errs *[]error) {
	t, ok := attributes["type"]
	if ok {
		if !cls.IsInheritance() {
			*errs = append(*errs, errors.New("it is not inheritance and cannot contains 'type'"))
			return
		}
		cm := stringutils.Underscore(cls.Name)
		if cm != t {
			*errs = append(*errs, fmt.Errorf("'%t' of 'type' is not equal class '%s'", t, cm))
			return
		}
		new_attributes["type"] = t
	} else {
		if !is_update && cls.IsInheritance() {
			new_attributes["type"] = stringutils.Underscore(cls.Name)
		}
	}

	t, ok = attributes["parent_type"]
	// if ok {
	// 	if nm, ok := t.(string); ok {
	// 		pcls := self.definitions.FindByUnderscoreName(nm)
	// 		if nil == pcls {
	// 			*errs = append(*errs, fmt.Errorf("parent class '%s' is not found", nm))
	// 			return
	// 		}
	// 		attributes["parent_type"] = getRootClassName(pcls)
	// 	}
	// }

}

func (self *mdb_server) preWrite(cls *ClassDefinition, uattributes map[string]interface{},
	is_update bool) (map[string]interface{}, error) {
	attributes := uattributes
	if self.restrict {
		attributes = make(map[string]interface{}, len(attributes))
		for k, v := range uattributes {
			attributes[k] = v
		}
	}

	new_attributes := make(map[string]interface{}, len(attributes))
	errs := make([]error, 0, 10)

	self.doInheritance(cls, attributes, new_attributes, is_update, &errs)

	for k, pr := range cls.Properties {

		if doCreatedAtAndUpdatedAt(k, attributes, new_attributes, is_update, &errs) {
			continue
		}

		value, ok := attributes[k]
		if !ok {
			if COLLECTION_UNKNOWN != pr.Collection {
				continue
			}

			if is_update {
				continue
			}

			if pr.IsRequired {
				errs = append(errs, errors.New("'"+k+"' is required"))
				continue
			}
			if nil == pr.DefaultValue {
				continue
			}

			new_attributes[k] = pr.DefaultValue
			continue
		}

		if pr.IsReadOnly {
			errs = append(errs, errors.New("'"+k+"' is readonly"))
			continue
		}

		if self.restrict {
			delete(attributes, k)
		}

		if COLLECTION_UNKNOWN == pr.Collection {
			new_value, is_failed := checkValue(pr, uattributes, value, &errs)

			if is_failed {
				continue
			}

			new_attributes[k] = new_value
			continue
		}

		array, ok := value.([]interface{})
		if !ok {
			errs = append(errs, fmt.Errorf("'"+k+"' must is a collection, actual is %v", value))
			continue
		}
		is_failed := false
		new_array := make([]interface{}, 0, len(array))
		for _, v := range array {
			nv, failed := checkValue(pr, uattributes, v, &errs)
			if failed {
				is_failed = true
			} else {
				new_array = append(new_array, nv)
			}
		}

		if is_failed {
			continue
		}

		new_attributes[k] = new_array
	}

	if 0 != len(errs) {
		return nil, commons.NewMutiErrors("validate failed", errs)
	}

	if self.restrict && 0 != len(attributes) {
		for k, _ := range attributes {
			if '$' == k[0] { // it is child of the model
				continue
			}
			errs = append(errs, errors.New("'"+k+"' is useless"))
		}
		return nil, commons.NewMutiErrors("validate failed", errs)
	}
	return new_attributes, nil
}

func (self *mdb_server) createChildren(cls *ClassDefinition, id interface{}, attributes map[string]interface{}) []string {
	warnings := []string{}
	for k, v := range attributes {
		if '$' != k[0] {
			continue
		}
		ccls := self.definitions.FindByUnderscoreName(k[1:])
		if nil == ccls {
			warnings = append(warnings, fmt.Sprintf("class '%s' of '%s' is not found", k[1:], k))
			continue
		}

		assoc := cls.RootClass().GetAssocationByCollectionName(ccls.CollectionName())
		if nil == assoc {
			warnings = append(warnings, fmt.Sprintf("class '%s' is not contains child that name is '%s' and collection is %s at the '%s'",
				cls.Name, k[1:], ccls.CollectionName(), k))
			continue
		}

		foreignKey := ""
		is_polymorphic := false
		switch a := assoc.(type) {
		case *HasMany:
			foreignKey = a.ForeignKey
			is_polymorphic = a.Polymorphic
		case *HasOne:
			foreignKey = a.ForeignKey
		default:
			warnings = append(warnings, fmt.Sprintf("class '%s' is not contains child that name is '%s' at the '%s'", cls.Name, k[1:], k))
		}

		commons.Each(v, func(ck interface{}, r interface{}) {
			attrs, ok := r.(map[string]interface{})
			if !ok {
				warnings = append(warnings, fmt.Sprintf("value of '%s.%s' is not map[string]interface{}", k, ck))
				return
			}

			lcls, err := self.findClassByAttributes(attrs)
			if nil != err {
				warnings = append(warnings, fmt.Sprintf("class '%s' of '%s.%s' is not found, %v - %v", k[1:], k, ck, err, attrs))
				return
			}
			if nil == lcls {
				lcls = ccls
			}
			if is_polymorphic {
				attrs["parent_type"] = getRootClassName(cls)
				attrs["parent_id"] = id
			} else {
				attrs[foreignKey] = id
			}
			sid, err := self.Create(lcls, map[string]string{}, attrs)
			if nil != err {
				warnings = append(warnings, fmt.Sprintf("save '%v.%v' failed, %v - %v", k, ck, err, attrs))
			}
			child_warnings := self.createChildren(lcls, sid, attrs)
			if nil != child_warnings && 0 != len(child_warnings) {
				for _, s := range child_warnings {
					warnings = append(warnings, "    "+s)
				}
			}
		}, func(instance interface{}) {
			warnings = append(warnings, fmt.Sprintf("value of '%s' is not []interface{} or map[string]interface{}", k))
		})
	}
	if 0 == len(warnings) {
		return nil
	}
	return warnings
}

func (self *mdb_server) checkChildren(attributes map[string]interface{}) error {
	for k, _ := range attributes {
		if 0 == len(k) {
			continue
		}

		if '$' == k[0] {
			return ChildrenUpdateError
		}
	}
	return nil
}

func (self *mdb_server) Create(cls *ClassDefinition, params map[string]string, attributes map[string]interface{}) (interface{}, error) {
	new_attributes, errs := self.preWrite(cls, attributes, false)
	if nil != errs {
		return nil, errs
	}
	if "true" == attributes["save"] {
		err := self.checkChildren(attributes)
		if nil != err {
			return nil, err
		}
		s, err := self.buildQueryStatement(cls, params)
		if nil != err {
			return nil, err
		}
		changeInfo, err := self.session.C(cls.CollectionName()).Upsert(s, new_attributes)
		if nil != err {
			return nil, err
		}
		return changeInfo.UpsertedId, nil
	}

	id, ok := new_attributes["_id"]
	if !ok {
		id = bson.NewObjectId()
		new_attributes["_id"] = id
	}
	err := self.session.C(cls.CollectionName()).Insert(new_attributes)
	if nil != err {
		return nil, err
	}

	return id, nil
}

func (self *mdb_server) Update(cls *ClassDefinition, id string, params map[string]string,
	updated_attributes map[string]interface{}) (int, error) {

	new_attributes, errs := self.preWrite(cls, updated_attributes, true)
	if nil != errs {
		return -1, errs
	}

	err := self.checkChildren(updated_attributes)
	if nil != err {
		return -1, err
	}

	switch id {
	case "all", "query":
		s, err := self.buildQueryStatement(cls, params)
		if nil != err {
			return -1, err
		}

		changeInfo, err := self.session.C(cls.CollectionName()).UpdateAll(s, bson.M{"$set": new_attributes})
		if nil != err {
			return -1, err
		}
		return changeInfo.Updated, nil
	}

	oid, err := parseObjectIdHex(id)
	if nil != err {
		return -1, errutils.BadRequest("id is not a objectId")
	}

	err = self.session.C(cls.CollectionName()).UpdateId(oid, bson.M{"$set": new_attributes})
	if nil != err {
		return -1, err
	}

	return 1, nil
}
