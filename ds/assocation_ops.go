package ds

// import (
// 	"commons/stringutils"
// 	"fmt"
// )

// var (
// 	assocationOps = make([]*assocationOp, 5)
// )

// type assocationOp struct {
// 	deleteOp    func(s *object_session, assoc types.Assocation, cls *types.TableDefinition, id interface{}) error
// 	deleteAllOp func(s *object_session, assoc types.Assocation, cls *types.TableDefinition) error
// 	// findOp      func(s *object_session, assoc types.Assocation, cls *types.TableDefinition, id interface{},
// 	// 	peer *types.TableDefinition) ([]map[string]interface{}, error)
// 	//createOp func(s *object_session, assoc *Assocation, id interface{}) error
// }

// func init() {
// 	assocationOps[BELONGS_TO] = &assocationOp{}
// 	assocationOps[HAS_ONE] = &assocationOp{deleteOp: deleteHasOne,
// 		deleteAllOp: deleteAllHasOne}
// 	assocationOps[HAS_MANG] = &assocationOp{deleteOp: deleteHasMany,
// 		deleteAllOp: deleteAllHasMany}
// 	assocationOps[HAS_AND_BELONGS_TO_MANY] = &assocationOp{deleteOp: deleteMany2Many,
// 		deleteAllOp: deleteAllMany2Many}
// }

// // func findHasOne(s *object_session, assoc types.Assocation, cls *types.TableDefinition, id interface{},
// // 	peer *types.TableDefinition) ([]map[string]interface{}, error) {
// // 	hasOne, ok := assoc.(*HasOne)
// // 	if !ok {
// // 		panic(fmt.Sprintf("it is a %T, please ensure it is a HasOne.", assoc))
// // 	}
// // 	return s.findBy(peer, map[string]string{"@" + hasOne.ForeignKey: IdString(id)})
// // }

// // func findHasMany(s *object_session, assoc types.Assocation, cls *types.TableDefinition, id interface{},
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

// func deleteHasOne(s *object_session, assoc types.Assocation, cls *types.TableDefinition, id interface{}) (int64, error) {
// 	hasOne, ok := assoc.(*types.HasOne)
// 	if !ok {
// 		panic(fmt.Sprintf("it is a %T, please ensure it is a HasOne.", assoc))
// 	}

// 	s.deleteByParams(hasOne.Target(), id)
// 	qc := bson.M{hasOne.ForeignKey: id}
// 	it := s.session.C(hasOne.Target().CollectionName()).Find(qc).Select(bson.M{"_id": 1}).Iter()
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

// func deleteHasMany(s *object_session, assoc types.Assocation, cls *types.TableDefinition, id interface{}) error {
// 	hasMany, ok := assoc.(*types.HasMany)
// 	if !ok {
// 		panic(fmt.Sprintf("it is a %T, please ensure it is a HasMay.", assoc))
// 	}
// }

// func deleteByParent(s *object_session,
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

// func deleteMany2Many(s *object_session, assoc types.Assocation, cls *types.TableDefinition, id interface{}) error {
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

// func deleteAllHasOne(s *object_session, assoc types.Assocation, cls *types.TableDefinition) error {
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

// func deleteAllHasMany(s *object_session, assoc types.Assocation, cls *types.TableDefinition) error {
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

// func deleteAllMany2Many(s *object_session, assoc types.Assocation, cls *types.TableDefinition) error {
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
