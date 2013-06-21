package ds

import (
	"commons/types"
	"database/sql"
	"errors"
	"fmt"
)

var (
	assocationOps = make([]*assocationOp, 5)
)

type assocationOp struct {
	deleteById func(s *session, assoc types.Assocation, cls *types.TableDefinition, id interface{}) (int64, error)
	deleteAll  func(s *session, assoc types.Assocation, cls *types.TableDefinition) (int64, error)
	// findOp      func(s *session, assoc types.Assocation, cls *types.TableDefinition, id interface{},
	// 	peer *types.TableDefinition) ([]map[string]interface{}, error)
	//createOp func(s *session, assoc *Assocation, id interface{}) error
}

func init() {
	assocationOps[types.BELONGS_TO] = &assocationOp{}
	assocationOps[types.HAS_ONE] = &assocationOp{deleteById: deleteByIdWithHasOne,
		deleteAll: deleteAllWithHasOne}
	assocationOps[types.HAS_MANY] = &assocationOp{deleteById: deleteByIdWithHasMany,
		deleteAll: deleteAllWithHasMany}
	assocationOps[types.HAS_AND_BELONGS_TO_MANY] = &assocationOp{deleteById: deleteByIdWithMany2Many,
		deleteAll: deleteAllWithMany2Many}
}

// // func findHasOne(s *session, assoc types.Assocation, cls *types.TableDefinition, id interface{},
// // 	peer *types.TableDefinition) ([]map[string]interface{}, error) {
// // 	hasOne, ok := assoc.(*HasOne)
// // 	if !ok {
// // 		panic(fmt.Sprintf("it is a %T, please ensure it is a HasOne.", assoc))
// // 	}
// // 	return s.findBy(peer, map[string]string{"@" + hasOne.ForeignKey: IdString(id)})
// // }

// // func findHasMany(s *session, assoc types.Assocation, cls *types.TableDefinition, id interface{},
// // 	peer *types.TableDefinition) ([]map[string]interface{}, error) {
// // 	hasMany, ok := assoc.(*HasMany)
// // 	if !ok {
// // 		panic(fmt.Sprintf("it is a %T, please ensure it is a HasMay.", assoc))
// // 	}
// // 	if hasMany.Polymorphic {
// // 		return s.findBy(peer, map[string]string{
// // 			"@parent_type": stringutils.Underscore(cls.Name), "@parent_id": IdString(id)})
// // 	}
// // 	return s.findBy(peer, map[string]string{"@" + hasMany.ForeignKey: IdString(id)})
// // }

func deleteByIdWithHasOne(s *session, assoc types.Assocation, cls *types.TableDefinition, id interface{}) (int64, error) {
	hasOne, ok := assoc.(*types.HasOne)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasOne.", assoc))
	}
	if hasOne.Polymorphic {
		return s.delete(hasOne.Target(), map[string]string{"@parent_type": cls.UnderscoreName, "@parent_id": fmt.Sprint(id)})
	} else {
		return s.delete(hasOne.Target(), map[string]string{"@" + hasOne.ForeignKey: fmt.Sprint(id)})
	}
}

func deleteByIdWithHasMany(s *session, assoc types.Assocation, cls *types.TableDefinition, id interface{}) (int64, error) {
	hasMany, ok := assoc.(*types.HasMany)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasMay.", assoc))
	}
	if hasMany.Polymorphic {
		return s.delete(hasMany.Target(), map[string]string{"@parent_type": cls.UnderscoreName, "@parent_id": fmt.Sprint(id)})
	} else {
		return s.delete(hasMany.Target(), map[string]string{"@" + hasMany.ForeignKey: fmt.Sprint(id)})
	}
}

func deleteByIdWithMany2Many(s *session, assoc types.Assocation, cls *types.TableDefinition, id interface{}) (int64, error) {
	habtm, ok := assoc.(*types.HasAndBelongsToMany)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasAndBelongsToMany.", assoc))
	}

	return s.delete(habtm.Through, map[string]string{"@" + habtm.ForeignKey: fmt.Sprint(id)})
}

func deleteAllWithHasOne(s *session, assoc types.Assocation, cls *types.TableDefinition) (int64, error) {
	hasOne, ok := assoc.(*types.HasOne)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasOne.", assoc))
	}
	if hasOne.Polymorphic {
		return s.delete(hasOne.Target(), map[string]string{"@parent_type": cls.UnderscoreName})
	} else {
		return s.delete(hasOne.Target(), nil)
	}
}

func deleteAllWithHasMany(s *session, assoc types.Assocation, cls *types.TableDefinition) (int64, error) {
	hasMany, ok := assoc.(*types.HasMany)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasMay.", assoc))
	}
	if hasMany.Polymorphic {
		return s.delete(hasMany.Target(), map[string]string{"@parent_type": cls.UnderscoreName})
	} else {
		return s.delete(hasMany.Target(), nil)
	}
}

func deleteAllWithMany2Many(s *session, assoc types.Assocation, cls *types.TableDefinition) (int64, error) {
	habtm, ok := assoc.(*types.HasAndBelongsToMany)
	if !ok {
		panic(fmt.Sprintf("it is a %T, please ensure it is a HasAndBelongsToMany.", assoc))
	}

	return s.delete(habtm.Through, nil)
}

// func deleteHasMany(s *session, assoc types.Assocation, cls *types.TableDefinition, id interface{}) error {
// 	hasMany, ok := assoc.(*types.HasMany)
// 	if !ok {
// 		panic(fmt.Sprintf("it is a %T, please ensure it is a HasMay.", assoc))
// 	}
// }

// func deleteByParent(s *session,
// 	parent_table *types.TableDefinition,
// 	parent_id interface{},
// 	polymorphic bool,
// 	table *types.TableDefinition) error {

// 	qc := map[string]string{}

// 	if polymorphic {
// 		qc["parent_type"] = "parent_id": id}
// 	} else {
// 		qc = bson.M{hasMany.ForeignKey: id}
// 	}

// 	it := s.session.C(hasMany.Target().CollectionName()).Find(qc).Select(bson.M{"_id": 1}).Iter()
// 	var result map[string]interface{}
// 	for it.Next(&result) {
// 		o, ok := result["_id"]
// 		if !ok {
// 			continue
// 		}

// 		_, err := s.removeById(assoc.Target(), o)
// 		if nil != err && "not found" == err.Error() {
// 			return err
// 		}
// 	}
// 	return it.Err()
// }

// func deleteMany2Many(s *session, assoc types.Assocation, cls *types.TableDefinition, id interface{}) error {
// 	habtm, ok := assoc.(*types.HasAndBelongsToMany)
// 	if !ok {
// 		panic(fmt.Sprintf("it is a %T, please ensure it is a HasAndBelongsToMany.", assoc))
// 	}
// 	it := s.session.C(habtm.CollectionName).Find(bson.M{habtm.ForeignKey: id}).Select(bson.M{"_id": 1}).Iter()

// 	var result map[string]interface{}
// 	for it.Next(&result) {
// 		o, ok := result["_id"]
// 		if !ok {
// 			continue
// 		}
// 		_, err := s.removeById(assoc.Target(), o)
// 		if nil != err && "not found" == err.Error() {
// 			return err
// 		}
// 	}

// 	return it.Err()
// }

// func deleteAllHasOne(s *session, assoc types.Assocation, cls *types.TableDefinition) error {
// 	hasOne, ok := assoc.(*types.HasOne)
// 	if !ok {
// 		panic(fmt.Sprintf("it is a %T, please ensure it is a HasOne.", assoc))
// 	}
// 	cn := hasOne.Target().CollectionName()
// 	_, err := s.removeAll(hasOne.Target(), map[string]string{})
// 	if nil != err {
// 		if !collectionExists(s, cn) {
// 			return nil
// 		}
// 		return fmt.Errorf("delete '%s' collection failed, %v", cn, err)
// 	}
// 	return nil
// }

// func deleteAllHasMany(s *session, assoc types.Assocation, cls *types.TableDefinition) error {
// 	hasMany, ok := assoc.(*types.HasMany)
// 	if !ok {
// 		panic(fmt.Sprintf("it is a %T, please ensure it is a HasMay.", assoc))
// 	}
// 	cn := hasMany.Target().CollectionName()
// 	if hasMany.Polymorphic {
// 		_, err := s.removeBy(hasMany.Target(), map[string]string{"@parent_type": stringutils.Underscore(cls.Name)})
// 		if nil != err {
// 			return fmt.Errorf("delete from '%s' collection failed, %v", cn, err)
// 		}
// 		return nil
// 	}
// 	_, err := s.removeAll(hasMany.Target(), map[string]string{})
// 	if nil != err {
// 		if !collectionExists(s, cn) {
// 			return nil
// 		}
// 		return fmt.Errorf("delete '%s' collection failed, %v", cn, err)
// 	}
// 	return nil
// }

// func deleteAllMany2Many(s *session, assoc types.Assocation, cls *types.TableDefinition) error {
// 	habtm, ok := assoc.(*types.HasAndBelongsToMany)
// 	if !ok {
// 		panic(fmt.Sprintf("it is a %T, please ensure it is a HasAndBelongsToMany.", assoc))
// 	}
// 	cn := habtm.CollectionName
// 	err := s.session.C(cn).DropCollection()
// 	if nil != err {
// 		if !collectionExists(s, cn) {
// 			return nil
// 		}
// 		return fmt.Errorf("delete '%s' collection failed, %v", cn, err)
// 	}
// 	return nil
// }

func (self *session) deleteCascadeAll(table *types.TableDefinition) (int64, error) {
	deleted_all := int64(0)
	for s := table; nil != s; s = s.Super {
		for _, a := range s.Assocations {
			op := assocationOps[a.Type()]
			if nil == op || nil == op.deleteAll {
				continue
			}
			deleted, err := op.deleteAll(self, a, s)
			if nil != err {
				return deleted_all, err
			} else {
				deleted_all += deleted
			}
		}
	}
	return deleted_all, nil
}

func (self *session) deleteCascadeByIdAndTrueTable(table *types.TableDefinition, id interface{}) (int64, error) {
	deleted_all := int64(0)
	for s := table; nil != s; s = s.Super {
		for _, a := range s.Assocations {
			op := assocationOps[a.Type()]
			if nil == op || nil == op.deleteById {
				continue
			}
			deleted, err := op.deleteById(self, a, s, id)
			if nil != err {
				return deleted_all, err
			} else {
				deleted_all += deleted
			}
		}
	}
	return deleted_all, nil
}

func (self *session) deleteCascadeByParams(table *types.TableDefinition, params map[string]string) (int64, error) {
	deleted_all := int64(0)
	e := self.drv.forEach(table, params, func(rtable *types.TableDefinition, id interface{}) error {
		deleted_single, e := self.deleteCascadeByIdAndTrueTable(rtable, id)
		if nil != e {
			return e
		}
		deleted_all += deleted_single
		return nil
	})
	if nil != e {
		return 0, e
	}
	return deleted_all, nil
}

func (self *session) deleteCascadeById(table *types.TableDefinition, id interface{}) (int64, error) {
	var first_table *types.TableDefinition

	e := self.drv.forEach(table, map[string]string{"@id": fmt.Sprint(id)},
		func(rtable *types.TableDefinition, id interface{}) error {
			if nil != first_table {
				return errors.New("rows that id is '" + fmt.Sprint(id) + "' is more than one.")
			}
			first_table = rtable
			return nil
		})
	if nil != e {
		return 0, e
	}
	if nil == first_table {
		return 0, sql.ErrNoRows
	}

	return self.deleteCascadeByIdAndTrueTable(first_table, id)
}
