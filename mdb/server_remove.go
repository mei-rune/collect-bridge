package mdb

import (
	"commons"
	"commons/errutils"
	"commons/stringutils"
	"fmt"
	"labix.org/v2/mgo/bson"
)

func internalError(e error) commons.RuntimeError {
	return errutils.InternalError(e.Error())
}

func collectionExists(s *mdb_server, cn string) bool {
	names, e := s.session.CollectionNames()
	if nil != e {
		return false
	}
	for _, nm := range names {
		if nm == cn {
			return true
		}
	}
	return false
}

func (self *mdb_server) removeChildren(cls *ClassDefinition, id interface{}) (int, error) {
	deleted := 0
	for _, a := range cls.Assocations {
		op := assocationOps[a.Type()]
		if nil == op || nil == op.deleteOp {
			continue
		}
		err := op.deleteOp(self, a, cls, id)
		if nil != err {
			return deleted, err
		} else {
			deleted++
		}
	}
	return deleted, nil
}

func (self *mdb_server) removeAllChildren(cls *ClassDefinition) (int, error) {
	deleted := 0
	for _, a := range cls.Assocations {
		op := assocationOps[a.Type()]
		if nil == op || nil == op.deleteAllOp {
			continue
		}
		err := op.deleteAllOp(self, a, cls)
		if nil != err {
			return deleted, err
		} else {
			deleted++
		}
	}
	return deleted, nil
}

func (self *mdb_server) removeById(cls *ClassDefinition, id interface{}) (int, commons.RuntimeError) {
	deleted, err := self.removeChildren(cls, id)
	if nil != err {
		return deleted, internalError(err)
	}

	err = self.session.C(cls.CollectionName()).RemoveId(id)
	if nil != err {
		if "not found" == err.Error() {
			return 0, errutils.RecordNotFound(IdString(id))
		}

		return deleted, errutils.InternalError("delete " + stringutils.Underscore(cls.Name) + " fialed, " + err.Error())
	}
	return 1, nil
}

func (self *mdb_server) removeBy(cls *ClassDefinition, params map[string]string) (int, commons.RuntimeError) {
	s, err := self.buildQueryStatement(cls, params)
	if nil != err {
		return -1, internalError(err)
	}
	collection := self.session.C(cls.CollectionName())
	deleted := 0

	iter := collection.Find(s).Select(bson.M{"_id": 1}).Iter()
	var result map[string]interface{}
	for iter.Next(&result) {
		_, e := self.removeChildren(cls, result["_id"])
		if nil != e {
			return deleted, internalError(e)
		} else {
			deleted++
		}
	}

	err = iter.Err()
	if nil != err {
		return deleted, errutils.InternalError("delete " + cls.Name + "failed, " + err.Error())
	}

	changeInfo, err := collection.RemoveAll(s)
	if nil != err {
		return deleted, internalError(err)
	}

	return changeInfo.Removed, nil
}

func (self *mdb_server) removeAll(cls *ClassDefinition, params map[string]string) (int, commons.RuntimeError) {
	if cls.IsInheritance() && nil != cls.Super {
		return self.removeBy(cls, params)
	}

	deleted, err := self.removeAllChildren(cls)
	if nil != err {
		return deleted, internalError(err)
	}
	err = self.session.C(cls.CollectionName()).DropCollection()
	if nil != err {
		if !collectionExists(self, cls.CollectionName()) {
			return -1, nil
		}

		return deleted, internalError(err)
	}
	return -1, nil
}

func (self *mdb_server) Delete(cls *ClassDefinition, id string, params map[string]string) (int, string, commons.RuntimeError) {
	switch id {
	case "all":
		effected, e := self.removeAll(cls, params)
		if nil != e {
			return 0, fmt.Sprintf("delete some chilren, count is %d", effected), e
		} else {
			return effected, "", nil
		}
	case "query":
		effected, e := self.removeBy(cls, params)
		if nil != e {
			return 0, fmt.Sprintf("delete some chilren, count is %d", effected), e
		} else {
			return effected, "", nil
		}
	}

	oid, err := parseObjectIdHex(id)
	if nil != err {
		return 0, "", errutils.BadRequest("id is not a objectId")
	}

	effected, e := self.removeById(cls, oid)
	if nil != e {
		return 0, fmt.Sprintf("delete some chilren, count is %d", effected), e
	}
	return effected, "", nil
}
