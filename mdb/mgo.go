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
	session  *mgo.Database
	restrict bool
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

func (self *mgo_driver) writeObject(cd *ClassDefinition, attributes map[string]interface{}, parents []ObjectId, writeSytle WriteSytle) (id interface{}, err error) {

	if nil == parents || 0 == len(parents) {
		m, err := toM(cd, attributes)
		if nil != err {
			return nil, err
		}

		err = self.session.C(cd.CollectionName()).Insert(bson.M{"base": m})
	} else {
		// db.inventory.update({"_id" : ObjectId("50b5f5bb6456afaf3407800b"), "rules.bb":
		//      {$exists: false}},{$set:{"rules.bb":{"a1":"1"}}})

		// if cd.CollectionName() != parents[0][0] {
		//	return nil, errors.New("collectionName is error")
		// }
		id := parents[0].definition.CollectionName()
		m, err := toM(cd, attributes)
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

func (self *mgo_driver) pre(cls *ClassDefinition, attributes map[string]interface{},
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

func (self *mgo_driver) post(cls *ClassDefinition, attributes map[string]interface{}) (map[string]interface{}, error) {

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

func (self *mgo_driver) Insert(cls *ClassDefinition, attributes map[string]interface{}) (interface{}, error) {
	new_attributes, errs := self.pre(cls, attributes, false)
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

func (self *mgo_driver) Update(cls *ClassDefinition, id interface{}, updated_attributes map[string]interface{}) error {
	new_attributes, errs := self.pre(cls, updated_attributes, true)
	if nil != errs {
		return errs
	}

	return self.session.C(cls.CollectionName()).UpdateId(id, bson.M{"$set": new_attributes})
}

func (self *mgo_driver) FindById(cls *ClassDefinition, id interface{}) (result map[string]interface{}, err error) {
	var q *mgo.Query
	q = self.session.C(cls.CollectionName()).FindId(id)
	if nil == q {
		return nil, errors.New("return nil")
	}
	err = q.One(&result)
	if nil != err {
		return nil, err
	}

	return self.post(cls, result)
}

func (self *mgo_driver) Query(cls string, attributes map[string]interface{}) Query {
	return self.session.C(cls).Find(attributes)
}

func (self *mgo_driver) FindBy(cls *ClassDefinition, attributes map[string]interface{}) Query {
	return self.session.C(cls.CollectionName()).Find(attributes)
}

func (self *mgo_driver) Delete(cls *ClassDefinition, id interface{}) (err error) {
	return self.session.C(cls.CollectionName()).RemoveId(id)
}
