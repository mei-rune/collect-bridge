package mdb

import (
	"fmt"
	q "labix.org/v2/mgo/bson"
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
	restrict    bool
	driver      Driver
	definitions *ClassDefinitions
}

func (self *mdb_server) Create(cls *ClassDefinition, attributes map[string]interface{}) (interface{}, error) {
	return self.driver.Insert(cls, attributes)
}

func (self *mdb_server) FindById(cls *ClassDefinition, id interface{}) (map[string]interface{}, error) {
	return self.driver.FindById(cls, id)
}

func (self *mdb_server) Update(cls *ClassDefinition, id interface{}, attributes map[string]interface{}) error {
	return self.driver.Update(cls, id, attributes)
}

func deleteChildren(s *mdb_server, assoc Assocation, id interface{}) error {
	hasMany, ok := assoc.(*HasMany)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasMay.", assoc))
	}
	it := s.driver.FindBy(assoc.Target(), q.M{hasMany.ForeignKey: id}).Select("_id").Iter()

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
	it := s.driver.Query(habtm.CollectionName, q.M{habtm.ForeignKey: id}).Select("_id").Iter()

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
	err := self.driver.Delete(cls, id)
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
