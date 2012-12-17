package mdb

import (
	"bytes"
	"errors"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strconv"
)

type WriteSytle int

const (
	WRITE_SAVE   WriteSytle = 0
	WRITE_INSERT WriteSytle = 1
	WRITE_UPDATE WriteSytle = 2
)

type mgo_driver struct {
	session *mgo.Database
}

func toM(cd *ClassDefinition, properties map[string]interface{}) (values bson.M, err error) {
	values = make(map[string]interface{})
	for k, v := range properties {
		value := v
		pr, ok := cd.GetProperty(k)
		if ok {
			value, err = pr.Type.Convert(v)
			if nil != err {
				return nil, err
			}
		}
		values[k] = value
	}
	return values, nil
}

func toPath(parents []ObjectId) string {
	var buf bytes.Buffer
	for i, ss := range parents {
		if 0 != i {
			buf.WriteString(ss.definition.CollectionName())
			buf.WriteString(".")
			buf.WriteString(ss.id)
		}
	}
	return buf.String()
}

func (self *mgo_driver) removeObject(cd *ClassDefinition, id string, parents []ObjectId) (err error) {

	if nil == parents || 0 == len(parents) {
		err = self.session.C(cd.CollectionName()).RemoveId(id)
	} else {
		// db.inventory.update({"_id" : ObjectId("50b5f5bb6456afaf3407800b"), "rules.bb":
		//      {$exists: false}},{$unset:{"rules.bb":0}})

		if cd.CollectionName() != parents[0].definition.CollectionName() {
			return errors.New("collectionName is error")
		}

		path := toPath(parents)
		query := bson.M{"_id": bson.ObjectIdHex(id), path: bson.M{"$exists": "true"}}

		err = self.session.C(cd.CollectionName()).Update(query, bson.M{"$unset": bson.M{path: 0}})
	}
	if nil != err {
		e, ok := err.(*mgo.LastError)
		if !ok {
			return err
		}
		if "" != e.Err {
			return err
		}
	}

	return nil
}

func (self *mgo_driver) writeObject(cd *ClassDefinition, properties map[string]interface{}, parents []ObjectId, writeSytle WriteSytle) (id interface{}, err error) {

	if nil == parents || 0 == len(parents) {
		m, err := toM(cd, properties)
		if nil != err {
			return nil, err
		}

		err = self.session.C(cd.CollectionName()).Insert(bson.M{"base": m})
	} else {
		// db.inventory.update({"_id" : ObjectId("50b5f5bb6456afaf3407800b"), "rules.bb":
		//      {$exists: false}},{$set:{"rules.bb":{"a1":"1"}}})

		// if cd.CollectionName() != parents[0][0] {
		// 	return nil, errors.New("collectionName is error")
		// }
		id := parents[0].definition.CollectionName()
		m, err := toM(cd, properties)
		if nil != err {
			return nil, err
		}
		var query bson.M
		path := toPath(parents)
		switch writeSytle {
		case WRITE_SAVE:
			query = bson.M{"_id": bson.ObjectIdHex(id)}
		case WRITE_INSERT:
			query = bson.M{"_id": bson.ObjectIdHex(id), path: bson.M{"$exists": "false"}}
		case WRITE_UPDATE:
			query = bson.M{"_id": bson.ObjectIdHex(id), path: bson.M{"$exists": "true"}}
		default:
			return nil, errors.New("unsupported write style - " + strconv.Itoa(int(writeSytle)))
		}
		err = self.session.C(cd.CollectionName()).Update(query, bson.M{"$set": bson.M{path: bson.M{"base": m}}})
	}
	if nil != err {
		e, ok := err.(*mgo.LastError)
		if !ok {
			return nil, err
		}
		if "" != e.Err {
			return nil, err
		}
		if 1 != e.N {
			return nil, errors.New("number of excepted change, actual is " + strconv.Itoa(e.N))
		}
		id = e.UpsertedId
	}

	return id, nil
}

func (self *mgo_driver) Insert(cls *ClassDefinition, attributes map[string]interface{}) (interface{}, error) {
	id, ok := attributes["_id"]
	if !ok {
		id = bson.NewObjectId()
		attributes["_id"] = id
	}
	err := self.session.C(cls.CollectionName()).Insert(attributes)
	if nil == err {
		return id, nil
	}
	return nil, err
}

func (self *mgo_driver) Update(cls *ClassDefinition, id string, attributes map[string]interface{}) error {
	err := self.session.C(cls.CollectionName()).UpdateId(id, bson.M{"$set": attributes})

	if nil != err {
		e, ok := err.(*mgo.LastError)
		if !ok {
			return err
		}
		if "" != e.Err {
			return err
		}
		if 1 != e.N {
			return errors.New("number of excepted change, actual is " + strconv.Itoa(e.N))
		}
		return nil
	}

	return errors.New("update failed, return nil.")
}

func (self *mgo_driver) FindById(cls *ClassDefinition, id string) (result interface{}, err error) {
	var q *mgo.Query
	q = self.session.C(cls.CollectionName()).FindId(id)
	if nil == q {
		return nil, errors.New("return nil")
	}
	err = q.One(&result)

	return result, err
}

func (self *mgo_driver) Delete(cd *ClassDefinition, id string) (err error) {
	err = self.session.C(cd.CollectionName()).RemoveId(id)
	if nil != err {
		e, ok := err.(*mgo.LastError)
		if !ok {
			return err
		}
		if "" != e.Err {
			return err
		}
	}
	return nil
}
