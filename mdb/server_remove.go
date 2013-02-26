package mdb

import (
	"commons"
	"errors"
	"labix.org/v2/mgo/bson"
)

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

func (self *mdb_server) RemoveById(cls *ClassDefinition, id interface{}) (bool, error) {
	err := self.session.C(cls.CollectionName()).RemoveId(id)
	if nil != err {
		return false, err
	}
	_, err = self.removeChildren(cls, id)
	return true, err
}

func (self *mdb_server) removeChildren(cls *ClassDefinition, id interface{}) (bool, error) {
	errs := make([]error, 0)
	for _, a := range cls.Assocations {
		op := assocationOps[a.Type()]
		if nil == op || nil == op.deleteOp {
			continue
		}
		err := op.deleteOp(self, a, cls, id)
		if nil != err {
			errs = append(errs, err)
		}
	}
	if 0 == len(errs) {
		return true, nil
	}
	return false, commons.NewMutiErrors("parameters is error.", errs)
}

func (self *mdb_server) removeAllChildren(cls *ClassDefinition) error {
	errs := make([]error, 0)
	for _, a := range cls.Assocations {
		op := assocationOps[a.Type()]
		if nil == op || nil == op.deleteAllOp {
			continue
		}
		err := op.deleteAllOp(self, a, cls)
		if nil != err {
			errs = append(errs, err)
		}
	}
	if 0 == len(errs) {
		return nil
	}
	return commons.NewMutiErrors("parameters is error.", errs)
}

func (self *mdb_server) RemoveBy(cls *ClassDefinition, params map[string]string) (bool, error) {
	s, err := self.buildQueryStatement(cls, params)
	if nil != err {
		return false, err
	}
	c := self.session.C(cls.CollectionName())
	q := c.Find(s)
	if nil == q {
		return false, errors.New("return nil result")
	}
	iter := q.Select(bson.M{"_id": 1}).Iter()
	var result map[string]interface{}
	for iter.Next(&result) {
		self.removeChildren(cls, result["_id"])
	}

	err = iter.Err()
	if nil != err {
		return false, err
	}
	_, err = c.RemoveAll(s)
	return (nil == err), err
}

func (self *mdb_server) RemoveAll(cls *ClassDefinition, params map[string]string) (bool, error) {
	err := self.removeAllChildren(cls)
	if nil != err {
		return false, err
	}
	err = self.session.C(cls.CollectionName()).DropCollection()
	if nil != err {
		if !collectionExists(self, cls.CollectionName()) {
			return true, nil
		}

		return false, err
	}
	return true, nil
}
