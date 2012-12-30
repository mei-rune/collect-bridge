package mdb

import (
	"commons"
	"commons/as"
	"errors"
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strings"
)

var (
	assocationOps = make([]*assocationOp, 4)
	operators     = make(map[string]func(pr *PropertyDefinition, s string) (interface{}, error))
)

type assocationOp struct {
	deleteOp func(s *mdb_server, assoc Assocation, id interface{}) error
	//createOp func(s *mdb_server, assoc *Assocation, id interface{}) error
}

func init() {
	assocationOps[BELONGS_TO] = &assocationOp{}
	assocationOps[HAS_MANG] = &assocationOp{deleteOp: deleteChildren}
	assocationOps[HAS_AND_BELONGS_TO_MANY] = &assocationOp{deleteOp: deleteMany2Many}

	operators["exists"] = op_exist
	operators["in"] = op_in
	operators["nin"] = op_nin
	operators["gt"] = op_gt
	operators["gte"] = op_gte
	operators["eq"] = op_eq
	operators["ne"] = op_ne
	operators["lt"] = op_lt
	operators["lte"] = op_lte
}

func op_exist(pr *PropertyDefinition, s string) (interface{}, error) {
	switch s {
	case "true":
		return bson.M{"$exists": true}, nil
	case "false":
		return bson.M{"$exists": false}, nil
	}
	return nil, errors.New("'exist' of '" + pr.Name + "' require one bool operand - " + s)
}
func op_in(pr *PropertyDefinition, s string) (interface{}, error) {
	return nil, errors.New("not implemented")
}
func op_nin(pr *PropertyDefinition, s string) (interface{}, error) {
	return nil, errors.New("not implemented")
}
func op_gt(pr *PropertyDefinition, s string) (interface{}, error) {
	return nil, errors.New("not implemented")
}
func op_gte(pr *PropertyDefinition, s string) (interface{}, error) {
	return nil, errors.New("not implemented")
}
func op_eq(pr *PropertyDefinition, s string) (interface{}, error) {
	return nil, errors.New("not implemented")
}
func op_ne(pr *PropertyDefinition, s string) (interface{}, error) {
	return nil, errors.New("not implemented")
}
func op_lt(pr *PropertyDefinition, s string) (interface{}, error) {
	return nil, errors.New("not implemented")
}
func op_lte(pr *PropertyDefinition, s string) (interface{}, error) {
	return nil, errors.New("not implemented")
}

type mdb_server struct {
	session     *mgo.Database
	restrict    bool
	definitions *ClassDefinitions
}

func checkValue(pr *PropertyDefinition, attributes map[string]interface{}, value interface{}, errs []error) (interface{}, []error, bool) {

	new_value, err := pr.Type.Convert(value)
	if nil != err {
		errs = append(errs, errors.New("'"+pr.Name+"' convert to internal value failed, "+err.Error()))
		return nil, errs, false
	}

	if nil != pr.Restrictions && 0 != len(pr.Restrictions) {
		is_failed := false
		for _, r := range pr.Restrictions {
			if ok, err := r.Validate(new_value, attributes); !ok {
				errs = append(errs, errors.New("'"+pr.Name+"' is validate failed, "+err.Error()))
				is_failed = true
			}
		}

		if is_failed {
			return nil, errs, false
		}
	}
	return new_value, errs, true
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
	for k, pr := range cls.Properties {
		var new_value interface{}
		value, ok := attributes[k]
		if !ok {
			if COLLECTION_UNKNOWN == pr.Collection {
				continue
			}

			if is_update {
				continue
			}

			if pr.IsRequired {
				errs = append(errs, errors.New("'"+k+"' is required"))
				continue
			}
			new_value = pr.DefaultValue
		} else {
			if self.restrict {
				delete(attributes, k)
			}

			is_failed := false

			if COLLECTION_UNKNOWN == pr.Collection {
				new_value, errs, is_failed = checkValue(pr, uattributes, value, errs)
			} else {
				array, _ := as.AsArray(value)
				if nil == array {
					errs = append(errs, fmt.Errorf("'"+k+"' must is a collection, actual is %v", value))
					continue
				}
				new_array := make([]interface{}, 0, len(array))
				var nv interface{} = nil
				failed := false
				for _, v := range array {
					nv, errs, failed = checkValue(pr, uattributes, v, errs)
					if !failed {
						new_array = append(new_array, nv)
					} else {
						is_failed = true
					}
				}
				new_value = new_array
			}

			if is_failed {
				continue
			}
		}

		new_attributes[k] = new_value
	}

	if 0 != len(errs) {
		return nil, &MutiErrors{msg: "validate failed", errs: errs}
	}

	if self.restrict && 0 != len(attributes) {
		for k, _ := range attributes {
			errs = append(errs, errors.New("'"+k+"' is useless"))
		}
		return nil, &MutiErrors{msg: "validate failed", errs: errs}
	}
	return new_attributes, nil
}

func (self *mdb_server) postRead(cls *ClassDefinition, attributes map[string]interface{}) (map[string]interface{}, error) {

	new_attributes := make(map[string]interface{}, len(attributes))
	errs := make([]error, 0, 10)
	for k, pr := range cls.Properties {
		var new_value interface{}
		value, ok := attributes[k]
		if !ok {
			new_value = pr.DefaultValue
		} else {

			if COLLECTION_UNKNOWN == pr.Collection {
				var err error
				new_value, err = pr.Type.Convert(value)
				if nil != err {
					errs = append(errs, errors.New("'"+k+"' convert to internal value failed, "+err.Error()))
					continue
				}
			} else {
				array, _ := as.AsArray(value)
				if nil == array {
					errs = append(errs, fmt.Errorf("'"+k+"' must is a collection, actual is %v", value))
					continue
				}

				new_array := make([]interface{}, 0, len(array))
				is_failed := false
				for _, v := range array {
					nv, err := pr.Type.Convert(v)
					if nil != err {
						errs = append(errs, errors.New("'"+k+"' convert to internal value failed, "+err.Error()))
						is_failed = true
					} else {
						new_array = append(new_array, nv)
					}
				}

				if is_failed {
					continue
				}
				new_value = new_array
			}
		}

		new_attributes[k] = new_value
	}

	if 0 != len(errs) {
		return nil, &MutiErrors{msg: "validate failed", errs: errs}
	}

	return new_attributes, nil
}

func (self *mdb_server) Create(cls *ClassDefinition, attributes map[string]interface{}) (interface{}, error) {
	new_attributes, errs := self.preWrite(cls, attributes, false)
	if nil != errs {
		return nil, errs
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

func (self *mdb_server) FindById(cls *ClassDefinition, id interface{}) (map[string]interface{}, error) {
	var q *mgo.Query
	var result map[string]interface{}

	q = self.session.C(cls.CollectionName()).FindId(id)
	if nil == q {
		return nil, errors.New("return nil result")
	}
	err := q.One(&result)
	if nil != err {
		return nil, err
	}

	return self.postRead(cls, result)
}

func (self *mdb_server) Update(cls *ClassDefinition, id interface{}, updated_attributes map[string]interface{}) error {
	new_attributes, errs := self.preWrite(cls, updated_attributes, true)
	if nil != errs {
		return errs
	}

	return self.session.C(cls.CollectionName()).UpdateId(id, bson.M{"$set": new_attributes})
}

func deleteChildren(s *mdb_server, assoc Assocation, id interface{}) error {
	hasMany, ok := assoc.(*HasMany)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasMay.", assoc))
	}
	it := s.session.C(hasMany.Target().CollectionName()).Find(bson.M{hasMany.ForeignKey: id}).Select("_id").Iter()

	var result map[string]interface{}
	for it.Next(&result) {
		o, ok := result["_id"]
		if !ok {
			continue
		}
		s.RemoveById(assoc.Target(), o)
	}

	return it.Err()
}

func deleteMany2Many(s *mdb_server, assoc Assocation, id interface{}) error {
	habtm, ok := assoc.(*HasAndBelongsToMany)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasAndBelongsToMany.", assoc))
	}
	it := s.session.C(habtm.CollectionName).Find(bson.M{habtm.ForeignKey: id}).Select("_id").Iter()

	var result map[string]interface{}
	for it.Next(&result) {
		o, ok := result["_id"]
		if !ok {
			continue
		}
		s.RemoveById(assoc.Target(), o)
	}

	return it.Err()
}

func (self *mdb_server) RemoveById(cls *ClassDefinition, id interface{}) (bool, error) {
	err := self.session.C(cls.CollectionName()).RemoveId(id)
	if nil != err {
		return false, err
	}

	errs := make([]error, 0)
	for _, a := range cls.Assocations {
		op := assocationOps[a.Type()]
		if nil == op || nil == op.deleteOp {
			continue
		}
		err = op.deleteOp(self, a, id)
		if nil != err {
			errs = append(errs, err)
		}
	}
	if 0 == len(errs) {
		return true, nil
	}
	return true, &MutiErrors{msg: "", errs: errs}
}

func collectOwnProperties(cls *ClassDefinition, properties map[string]*PropertyDefinition) {

	for k, p := range cls.OwnProperties {
		properties[k] = p
	}

	if nil != cls.Children || 0 == len(cls.Children) {
		return
	}

	for _, child := range cls.Children {
		collectOwnProperties(child, properties)
	}
}

func collectProperties(cls *ClassDefinition) map[string]*PropertyDefinition {
	if nil != cls.Children || 0 == len(cls.Children) {
		return cls.Properties
	}

	properties := make(map[string]*PropertyDefinition, len(cls.Properties))

	for k, p := range cls.Properties {
		properties[k] = p
	}

	for _, child := range cls.Children {
		collectOwnProperties(child, properties)
	}
	return properties
}

func parseObjectIdHex(s string) (id bson.ObjectId, err error) {
	defer func() {
		if e := recover(); nil != e {
			err = commons.NewError(e)
		}
	}()

	v := bson.ObjectIdHex(s)
	return v, nil
}
func buildQueryStatement(cls *ClassDefinition, params map[string]string) (bson.M, error) {
	if nil == params || 0 == len(params) {
		return nil, nil
	}

	is_all := nil != cls.Children || 0 == len(cls.Children)
	properties := cls.Properties
	q := bson.M{}
	for nm, exp := range params {
		pr, _ := properties[nm]
		if nil == pr {
			if "_id" == nm {
				var err error
				if strings.HasPrefix(exp, "[eq]") {
					q["_id"], err = parseObjectIdHex(exp[4:])
				} else {
					q["_id"], err = parseObjectIdHex(exp)
				}
				if nil != err {
					return nil, errors.New("_id is a invalid ObjectId")
				}
				continue
			}
			if nil == pr {
				pos := strings.LastIndex(nm, ".")
				if -1 != pos {
					pr, _ := properties[nm[0:pos]]
				}

				if nil == pr {

					if is_all {
						return nil, errors.New("'" + nm + "' is not a property.")
					}
					properties = collectProperties(cls)
					is_all = true
					pr, _ = properties[nm]

					if nil == pr {
						if -1 != pos {
							pr, _ := properties[nm[0:pos]]
						}

						if nil == pr {
							return nil, errors.New("'" + nm + "' is not a property.")
						}
					}
				}
			}
		}

		var ss []string
		if '[' == exp[0] {
			ss := strings.SplitN(exp[1:], "]", 2)
		} else {
			ss = nil
		}

		if nil == ss || 2 != len(ss) {
			v, err := pr.Type.Convert(exp)
			if nil != err {
				return nil, errors.New("'" + nm + "' convert to " +
					pr.Type.Name() + ", failed, " + err.Error())
			}
			q[nm] = v
			continue
		}

		f, _ := operators[ss[0]]
		if nil == f {
			return nil, errors.New(" '" + ss[0] + "' is unsupported operator for '" +
				nm + "'.")
		}
		value, err := f(pr, ss[1])
		if nil != err {
			return nil, errors.New("'" + nm + "' convert to " +
				pr.Type.Name() + ", failed, " + err.Error())
		}
		q[nm] = value
	}
	return q, nil
}

func (self *mdb_server) FindBy(cls *ClassDefinition, params map[string]string) ([]map[string]interface{}, error) {

	s, err := buildQueryStatement(cls, params)
	if nil != err {
		return nil, err
	}

	q := self.session.C(cls.CollectionName()).Find(s)
	if nil == q {
		return nil, errors.New("return nil result")
	}

	results := make([]map[string]interface{}, 0, 10)
	it := q.Iter()
	var attributes map[string]interface{}
	for it.Next(&attributes) {
		attributes, err = self.postRead(cls, attributes)
		if nil != err {
			return nil, err
		}
		results = append(results, attributes)
	}

	err = it.Err()
	if nil != err {
		return nil, err
	}

	return results, nil
}
