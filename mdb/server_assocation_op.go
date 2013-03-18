package mdb

import (
	"commons/stringutils"
	"fmt"
	"labix.org/v2/mgo/bson"
)

var (
	assocationOps = make([]*assocationOp, 5)
)

type assocationOp struct {
	deleteOp    func(s *mdb_server, assoc Assocation, cls *ClassDefinition, id interface{}) error
	deleteAllOp func(s *mdb_server, assoc Assocation, cls *ClassDefinition) error
	findOp      func(s *mdb_server, assoc Assocation, cls *ClassDefinition, id interface{},
		peer *ClassDefinition) ([]map[string]interface{}, error)
	//createOp func(s *mdb_server, assoc *Assocation, id interface{}) error
}

func init() {
	assocationOps[BELONGS_TO] = &assocationOp{}
	assocationOps[HAS_ONE] = &assocationOp{deleteOp: deleteHasOne,
		deleteAllOp: deleteAllHasOne, findOp: findHasOne}
	assocationOps[HAS_MANG] = &assocationOp{deleteOp: deleteHasMany,
		deleteAllOp: deleteAllHasMany, findOp: findHasMany}
	assocationOps[HAS_AND_BELONGS_TO_MANY] = &assocationOp{deleteOp: deleteMany2Many,
		deleteAllOp: deleteAllMany2Many, findOp: findMany2Many}
}

func findHasOne(s *mdb_server, assoc Assocation, cls *ClassDefinition, id interface{},
	peer *ClassDefinition) ([]map[string]interface{}, error) {
	hasOne, ok := assoc.(*HasOne)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasOne.", assoc))
	}
	return s.findBy(peer, map[string]string{"@" + hasOne.ForeignKey: IdString(id)})
}

func deleteHasOne(s *mdb_server, assoc Assocation, cls *ClassDefinition, id interface{}) error {
	hasOne, ok := assoc.(*HasOne)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasOne.", assoc))
	}
	qc := bson.M{hasOne.ForeignKey: id}
	it := s.session.C(hasOne.Target().CollectionName()).Find(qc).Select(bson.M{"_id": 1}).Iter()
	var result map[string]interface{}
	for it.Next(&result) {
		o, ok := result["_id"]
		if !ok {
			continue
		}

		_, err := s.removeById(assoc.Target(), o)
		if nil != err && "not found" == err.Error() {
			return err
		}
	}
	return it.Err()
}

func deleteAllHasOne(s *mdb_server, assoc Assocation, cls *ClassDefinition) error {
	hasOne, ok := assoc.(*HasOne)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasOne.", assoc))
	}
	cn := hasOne.Target().CollectionName()
	_, err := s.removeAll(hasOne.Target(), map[string]string{})
	if nil != err {
		if !collectionExists(s, cn) {
			return nil
		}
		return fmt.Errorf("delete '%s' collection failed, %v", cn, err)
	}
	return nil
}

func findHasMany(s *mdb_server, assoc Assocation, cls *ClassDefinition, id interface{},
	peer *ClassDefinition) ([]map[string]interface{}, error) {
	hasMany, ok := assoc.(*HasMany)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasMay.", assoc))
	}
	if hasMany.Polymorphic {
		return s.findBy(peer, map[string]string{
			"@parent_type": stringutils.Underscore(cls.Name), "@parent_id": IdString(id)})
	}
	return s.findBy(peer, map[string]string{"@" + hasMany.ForeignKey: IdString(id)})
}

func deleteHasMany(s *mdb_server, assoc Assocation, cls *ClassDefinition, id interface{}) error {
	hasMany, ok := assoc.(*HasMany)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasMay.", assoc))
	}
	var qc bson.M
	if hasMany.Polymorphic {
		qc = bson.M{"parent_type": buildClassQuery(cls), "parent_id": id}
	} else {
		qc = bson.M{hasMany.ForeignKey: id}
	}

	fmt.Println(qc)
	it := s.session.C(hasMany.Target().CollectionName()).Find(qc).Select(bson.M{"_id": 1}).Iter()
	var result map[string]interface{}
	for it.Next(&result) {
		o, ok := result["_id"]
		if !ok {
			continue
		}

		_, err := s.removeById(assoc.Target(), o)
		if nil != err && "not found" == err.Error() {
			return err
		}
	}
	return it.Err()
}
func deleteAllHasMany(s *mdb_server, assoc Assocation, cls *ClassDefinition) error {
	hasMany, ok := assoc.(*HasMany)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasMay.", assoc))
	}
	cn := hasMany.Target().CollectionName()
	if hasMany.Polymorphic {
		_, err := s.removeBy(hasMany.Target(), map[string]string{"@parent_type": stringutils.Underscore(cls.Name)})
		if nil != err {
			return fmt.Errorf("delete from '%s' collection failed, %v", cn, err)
		}
		return nil
	}
	_, err := s.removeAll(hasMany.Target(), map[string]string{})
	if nil != err {
		if !collectionExists(s, cn) {
			return nil
		}
		return fmt.Errorf("delete '%s' collection failed, %v", cn, err)
	}
	return nil
}

func findMany2Many(s *mdb_server, assoc Assocation, cls *ClassDefinition, id interface{},
	peer *ClassDefinition) ([]map[string]interface{}, error) {
	hasBelongsToMany1, ok := assoc.(*HasAndBelongsToMany)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasMay.", assoc))
	}
	hasBelongsToMany2 := hasBelongsToMany1.TargetClass.GetAssocationByCollectionName(
		hasBelongsToMany1.CollectionName).(*HasAndBelongsToMany)
	if nil == hasBelongsToMany2 {
		panic(fmt.Sprintf("xxx.", assoc))
	}

	it := s.session.C(hasBelongsToMany1.CollectionName).Find(bson.M{hasBelongsToMany1.ForeignKey: id}).
		Select(bson.M{hasBelongsToMany2.ForeignKey: 1}).Iter()
	idlist := make([]interface{}, 0, 10)
	var result map[string]interface{}
	for it.Next(&result) {
		o, ok := result[hasBelongsToMany2.ForeignKey]
		if !ok {
			continue
		}
		idlist = append(idlist, o)
	}

	if nil != it.Err() {
		return nil, it.Err()
	}

	results := make([]map[string]interface{}, 0, 10)
	for _, id := range idlist {
		o, e := s.findById(hasBelongsToMany1.TargetClass, id, map[string]string{})
		if nil != e {
			return nil, e
		}
		results = append(results, o)
	}
	return results, nil
}

func deleteMany2Many(s *mdb_server, assoc Assocation, cls *ClassDefinition, id interface{}) error {
	habtm, ok := assoc.(*HasAndBelongsToMany)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasAndBelongsToMany.", assoc))
	}
	it := s.session.C(habtm.CollectionName).Find(bson.M{habtm.ForeignKey: id}).Select(bson.M{"_id": 1}).Iter()

	var result map[string]interface{}
	for it.Next(&result) {
		o, ok := result["_id"]
		if !ok {
			continue
		}
		_, err := s.removeById(assoc.Target(), o)
		if nil != err && "not found" == err.Error() {
			return err
		}
	}

	return it.Err()
}

func deleteAllMany2Many(s *mdb_server, assoc Assocation, cls *ClassDefinition) error {
	habtm, ok := assoc.(*HasAndBelongsToMany)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasAndBelongsToMany.", assoc))
	}
	cn := habtm.CollectionName
	err := s.session.C(cn).DropCollection()
	if nil != err {
		if !collectionExists(s, cn) {
			return nil
		}
		return fmt.Errorf("delete '%s' collection failed, %v", cn, err)
	}
	return nil
}
