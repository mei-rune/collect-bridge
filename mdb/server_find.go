package mdb

import (
	"commons"
	"commons/as"
	"commons/errutils"
	"commons/stringutils"
	"errors"
	"fmt"
	"labix.org/v2/mgo"
	"strconv"
	"strings"
)

func (self *mdb_server) postRead(raw_cls *ClassDefinition,
	attributes map[string]interface{}) (map[string]interface{}, error) {

	new_attributes := make(map[string]interface{}, len(attributes))

	cls := raw_cls
	if raw_cls.IsInheritance() {
		t, ok := attributes["type"]
		if !ok {
			return nil, fmt.Errorf("result is not contains 'type' - %v", attributes)
		}

		nm, ok := t.(string)
		if !ok {
			return nil, fmt.Errorf("'type' is not a string type, it is  '%T:%s' is not found - %v",
				t, t, attributes)
		}

		cls = self.definitions.FindByUnderscoreName(nm)
		if nil == cls {
			return nil, fmt.Errorf("class '%s' is not found  - %v", nm, attributes)
		}
		if !cls.InheritanceFrom(raw_cls) {
			return nil, fmt.Errorf("class '%s' is not child of class %s  - %v", cls.Name,
				raw_cls.Name, attributes)
		}
		new_attributes["type"] = nm
	}
	new_attributes["_id"] = attributes["_id"]

	errs := make([]error, 0, 10)
	for k, pr := range cls.Properties {
		value, ok := attributes[k]
		if !ok {
			if nil != pr.DefaultValue {
				new_attributes[k] = pr.DefaultValue
			}
			continue
		}

		if COLLECTION_UNKNOWN == pr.Collection {
			new_value, err := pr.Type.Convert(value)
			if nil != err {
				errs = append(errs, fmt.Errorf("value '%v' of key '%s' convert to internal value failed, %s",
					value, k, err.Error()))
			} else {
				new_attributes[k] = new_value
			}
			continue
		}
		array, _ := as.AsArray(value)
		if nil == array {
			errs = append(errs, fmt.Errorf("'"+k+"' must is a collection, actual is %v", value))
			continue
		}

		new_array := make([]interface{}, 0, len(array))
		is_failed := false
		for _, v := range array {
			nv, err := pr.Type.Convert(v)
			if nil != err {
				errs = append(errs, fmt.Errorf("value '%v' of key '%s' convert to internal value failed, %s",
					value, k, err.Error()))
				is_failed = true
			} else {
				new_array = append(new_array, nv)
			}
		}

		if is_failed {
			continue
		}
		new_attributes[k] = new_array
	}

	if t, ok := attributes["type"]; ok {
		new_attributes["type"] = t
	}

	if 0 != len(errs) {
		return nil, commons.NewMutiErrors("validate failed", errs)
	}
	return new_attributes, nil
}

func (self *mdb_server) collectIncludes(cls *ClassDefinition, params map[string]string) (
	map[*ClassDefinition]Assocation, error) {
	includes, ok := params["includes"]
	if !ok || 0 == len(includes) {
		return nil, nil
	}
	assocations := make(map[*ClassDefinition]Assocation, 10)
	for _, s := range strings.Split(includes, ",") {
		peer := self.definitions.FindByUnderscoreName(s)
		if nil == peer {
			return nil, errors.New("class '" + s + "' is not found in the includes.")
		}
		assoc := cls.GetAssocationByCollectionName(peer.CollectionName())
		if nil == assoc {
			return nil, errors.New("assocation that to '" + s + "' is not found in the includes.")
		}
		assocations[peer] = assoc
	}
	return assocations, nil
}

func (self *mdb_server) loadIncludes(cls *ClassDefinition, id interface{}, res map[string]interface{},
	includes map[*ClassDefinition]Assocation) error {
	if nil == includes {
		return nil
	}

	for peer, include := range includes {
		findOp := assocationOps[include.Type()].findOp
		if nil == findOp {
			continue
		}
		peerName := stringutils.Underscore(peer.Name)
		results, err := findOp(self, include, cls, id, peer)
		if nil != err {
			return errors.New("query includes '" + peerName + "' failed, " + err.Error())
		}
		res["$"+peerName] = results
	}
	return nil
}

func (self *mdb_server) findById(cls *ClassDefinition, id interface{}, params map[string]string) (
	map[string]interface{}, error) {
	var q *mgo.Query
	var result map[string]interface{}

	includes, err := self.collectIncludes(cls, params)
	if nil != err {
		return nil, err
	}

	q = self.session.C(cls.CollectionName()).FindId(id)
	if nil == q {
		return nil, errors.New("return nil result")
	}
	err = q.One(&result)
	if nil != err {
		return nil, err
	}

	res, err := self.postRead(cls, result)
	if nil != err {
		return nil, err
	}
	err = self.loadIncludes(cls, id, res, includes)
	if nil != err {
		return nil, err
	}
	return res, err
}

func (self *mdb_server) findBy(cls *ClassDefinition, params map[string]string) (
	[]map[string]interface{}, error) {
	s, err := self.buildQueryStatement(cls, params)
	if nil != err {
		return nil, err
	}

	includes, err := self.collectIncludes(cls, params)
	if nil != err {
		return nil, err
	}

	q := self.session.C(cls.CollectionName()).Find(s)
	if nil == q {
		return nil, errors.New("return nil result")
	}
	offset := params["offset"]
	if "" != offset {
		n, err := strconv.Atoi(offset)
		if nil != err {
			return nil, errors.New("'offset' is not a integer - " + offset)
		}
		q.Skip(n)
	}
	limit := params["limit"]
	if "" != limit {
		n, err := strconv.Atoi(limit)
		if nil != err {
			return nil, errors.New("'limit' is not a integer - " + limit)
		}
		q.Limit(n)
	}

	results := make([]map[string]interface{}, 0, 10)
	it := q.Iter()
	attributes := make(map[string]interface{})
	for it.Next(&attributes) {
		attributes, err = self.postRead(cls, attributes)
		if nil != err {
			return nil, err
		}
		err := self.loadIncludes(cls, attributes["_id"], attributes, includes)
		if nil != err {
			return nil, err
		}
		results = append(results, attributes)
		attributes = make(map[string]interface{})
	}

	err = it.Err()
	if nil != err {
		return nil, err
	}

	return results, nil
}

func (self *mdb_server) count(cls *ClassDefinition, params map[string]string) (int, error) {
	s, err := self.buildQueryStatement(cls, params)
	if nil != err {
		return -1, err
	}
	q := self.session.C(cls.CollectionName()).Find(s)
	if nil == q {
		return -1, errors.New("return nil result")
	}
	return q.Count()
}

func (self *mdb_server) Get(cls *ClassDefinition, id string, params map[string]string) (interface{}, commons.RuntimeError) {

	switch id {
	case "", "query":
		results, err := self.findBy(cls, params)
		if err != nil {
			return nil, commons.NewRuntimeError(commons.InternalErrorCode, "query result from db, "+err.Error())
		}
		return results, nil
	case "count":
		count, err := self.count(cls, params)
		if err != nil {
			return nil, commons.NewRuntimeError(commons.InternalErrorCode, "query result from db, "+err.Error())
		}
		return count, nil
	}
	oid, err := parseObjectIdHex(id)
	if nil != err {
		return nil, errutils.BadRequest("id is not a objectId")
	}
	result, err := self.findById(cls, oid, params)
	if err != nil {
		if "not found" == err.Error() {
			return nil, errutils.RecordNotFound(id)
		}

		return nil, commons.NewRuntimeError(commons.InternalErrorCode, "query result from db, "+err.Error())
	}

	return result, nil
}
