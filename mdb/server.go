package mdb

import (
	"errors"
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

var (
	assocationOps = make([]*assocationOp, 4)
)

type assocationOp struct {
	deleteOp func(s *mdb_server, assoc Assocation, id interface{}) error
	//createOp func(s *mdb_server, assoc *Assocation, id interface{}) error
}

func init() {
	assocationOps[BELONGS_TO] = &assocationOp{}
	assocationOps[HAS_MANG] = &assocationOp{deleteOp: deleteChildren}
	assocationOps[HAS_AND_BELONGS_TO_MANY] = &assocationOp{deleteOp: deleteMany2Many}
}

type mdb_server struct {
	session     *mgo.Database
	restrict    bool
	definitions *ClassDefinitions
}

func (self *mdb_server) preWrite(cls *ClassDefinition, attributes map[string]interface{},
	is_update bool) (map[string]interface{}, error) {

	new_attributes := make(map[string]interface{}, len(attributes))
	errs := make([]error, 0, 10)
	for k, pr := range cls.Properties {
		var new_value interface{}
		value, ok := attributes[k]
		if !ok {
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
			var err error
			new_value, err = pr.Type.Convert(value)
			if nil != err {
				errs = append(errs, errors.New("'"+k+"' convert to internal value failed, "+err.Error()))
				continue
			}
		}

		if nil != pr.Restrictions && 0 != len(pr.Restrictions) {
			is_failed := false
			for _, r := range pr.Restrictions {
				if ok, err := r.Validate(new_value, attributes); !ok {
					errs = append(errs, errors.New("'"+k+"' is validate failed, "+err.Error()))
					is_failed = true
				}
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
			var err error
			new_value, err = pr.Type.Convert(value)
			if nil != err {
				errs = append(errs, errors.New("'"+k+"' convert to internal value failed, "+err.Error()))
				continue
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
