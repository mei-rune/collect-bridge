package mdb

import (
	"errors"
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

var (
	operators = make(map[string]func(pr *PropertyDefinition, s string) (interface{}, error))
)

func init() {

	operators["exists"] = op_exist
	operators["in"] = op_in
	operators["nin"] = op_nin
	operators["gt"] = op_gt
	operators["gte"] = op_gte
	operators["eq"] = op_eq
	operators["ne"] = op_ne
	operators["lt"] = op_lt
	operators["lte"] = op_lte

	errors.New("")
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
	v, e := pr.Type.Convert(s)
	if nil != e {
		return nil, e
	}

	return bson.M{"$gt": v}, nil
}
func op_gte(pr *PropertyDefinition, s string) (interface{}, error) {
	v, e := pr.Type.Convert(s)
	if nil != e {
		return nil, e
	}

	return bson.M{"$gte": v}, nil
}
func op_eq(pr *PropertyDefinition, s string) (interface{}, error) {
	v, e := pr.Type.Convert(s)
	if nil != e {
		return nil, e
	}

	return v, nil
}
func op_ne(pr *PropertyDefinition, s string) (interface{}, error) {
	v, e := pr.Type.Convert(s)
	if nil != e {
		return nil, e
	}

	return bson.M{"$ne": v}, nil
}
func op_lt(pr *PropertyDefinition, s string) (interface{}, error) {
	v, e := pr.Type.Convert(s)
	if nil != e {
		return nil, e
	}

	return bson.M{"$lt": v}, nil
}
func op_lte(pr *PropertyDefinition, s string) (interface{}, error) {
	v, e := pr.Type.Convert(s)
	if nil != e {
		return nil, e
	}

	return bson.M{"$lte": v}, nil
}

type mdb_server struct {
	session     *mgo.Database
	restrict    bool
	definitions *ClassDefinitions
}

func (self *mdb_server) findClassByAttributes(attributes map[string]interface{}) (*ClassDefinition, error) {
	if nil == attributes {
		return nil, nil
	}
	objectType, ok := attributes["type"]
	if !ok {
		return nil, nil
	}

	t, ok := objectType.(string)
	if !ok {
		return nil, errors.New(fmt.Sprintf("type '%v' in body is not a string type", objectType))
	}

	definition := self.definitions.FindByUnderscoreName(t)
	if nil == definition {
		return nil, errors.New("class '" + t + "' is not found")
	}
	return definition, nil
}

func (self *mdb_server) findClass(t string, attributes map[string]interface{}) (*ClassDefinition, error) {
	cls, e := self.findClassByAttributes(attributes)
	if nil != e {
		return nil, e
	}
	if nil != cls {
		return cls, nil
	}
	definition := self.definitions.FindByUnderscoreName(t)
	if nil == definition {
		return nil, errors.New("class '" + t + "' is not found")
	}
	return definition, nil
}

func IdString(id interface{}) string {
	if js, ok := id.(bson.ObjectId); ok {
		return js.Hex()
	}
	return fmt.Sprint(id)
}
