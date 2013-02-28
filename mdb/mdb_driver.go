package mdb

import (
	"commons"
	"commons/errutils"
	"encoding/json"
	"errors"
	"fmt"
	"labix.org/v2/mgo"
	"strings"
)

type MdbDriver struct {
	drvMgr *commons.DriverManager
	mdb_server
}

func NewMdbDriver(mgo_url, mgo_db string, drvMgr *commons.DriverManager) (*MdbDriver, error) {
	nm := commons.SearchFile("etc/mj_models.xml")
	if "" == nm {
		return nil, errors.New("'etc/mj_models.xml' is not found.")
	}
	definitions, e := LoadXml(nm)
	if nil != e {
		return nil, errors.New("load 'etc/mj_models.xml' failed, " + e.Error())
	}

	sess, err := mgo.Dial(mgo_url)
	if nil != err {
		return nil, errors.New("connect to mongo server '" + mgo_url + "' failed, " + err.Error())
	}

	sess.SetSafe(&mgo.Safe{W: 1, FSync: true, J: true})

	return &MdbDriver{drvMgr, mdb_server{session: sess.DB(mgo_db), definitions: definitions}}, nil
}

func (self *MdbDriver) findClassByBody(body map[string]interface{}) (*ClassDefinition, commons.RuntimeError) {
	if nil == body {
		return nil, nil
	}
	objectType, ok := body["type"]
	if !ok {
		return nil, nil
	}

	t, ok := objectType.(string)
	if !ok {
		return nil, errutils.BadRequest(fmt.Sprintf("type '%v' in body is not a string type", objectType))
	}

	definition := self.definitions.FindByUnderscoreName(t)
	if nil == definition {
		return nil, errutils.BadRequest("class '" + t + "' is not found")
	}
	return definition, nil
}

func (self *MdbDriver) findClass(params map[string]string, body map[string]interface{}) (*ClassDefinition, commons.RuntimeError) {
	definition, e := self.findClassByBody(body)
	if nil != definition || nil != e {
		return definition, e
	}

	objectType, _ := params["mdb.type"]
	if "" == objectType {
		return nil, errutils.IsRequired("mdb.type")
	}
	definition = self.definitions.FindByUnderscoreName(objectType)
	if nil == definition {
		return nil, errutils.BadRequest("class '" + objectType + "' is not found")
	}
	return definition, nil
}

func (self *MdbDriver) Create(params map[string]string) (commons.Result, commons.RuntimeError) {
	body, ok := params["body"]
	if !ok {
		return nil, commons.BodyNotExists
	}

	var attributes map[string]interface{}
	err := json.Unmarshal([]byte(body), &attributes)
	if err != nil {
		return nil, commons.NewRuntimeError(commons.InternalErrorCode, "unmarshal object from request failed, "+err.Error())
	}

	definition, e := self.findClass(params, attributes)
	if nil != e {
		return nil, e
	}

	instance_id, err := self.mdb_server.Create(definition, params, attributes)
	if err != nil {
		return nil, commons.NewRuntimeError(commons.InternalErrorCode, "insert object to db, "+err.Error())
	}
	warnings := self.mdb_server.createChildren(definition, instance_id, attributes)

	res := commons.Return(instance_id)
	if nil != warnings {
		res.Warnings(strings.Join(warnings, "\n"))
	}
	return res, nil
}

func (self *MdbDriver) Put(params map[string]string) (commons.Result, commons.RuntimeError) {
	id, _ := params["id"]
	if "" == id {
		return nil, commons.IdNotExists
	}

	body, ok := params["body"]
	if !ok {
		return nil, commons.BodyNotExists
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		return nil, commons.NewRuntimeError(commons.InternalErrorCode, "unmarshal object from request failed, "+err.Error())
	}

	definition, e := self.findClass(params, result)
	if nil != e {
		return nil, e
	}

	effected, err := self.mdb_server.Update(definition, id, params, result)
	if err != nil {
		if "not found" == err.Error() {
			return nil, errutils.RecordNotFound(id)
		}

		return nil, commons.NewRuntimeError(commons.InternalErrorCode, "update object to db, "+err.Error())
	}

	return commons.Return(effected), nil
}

func (self *MdbDriver) Delete(params map[string]string) (commons.Result, commons.RuntimeError) {
	objectType, _ := params["mdb.type"]
	if "" == objectType {
		return nil, errutils.IsRequired("mdb.type")
	}

	definition := self.definitions.FindByUnderscoreName(objectType)
	if nil == definition {
		return nil, errutils.BadRequest("class '" + objectType + "' is not found")
	}
	id, _ := params["id"]

	if "" == id {
		return nil, commons.IdNotExists
	}

	effected, warnings, err := self.mdb_server.Delete(definition, id, params)
	if nil != err {
		return nil, err
	}
	return commons.Return(effected).Effected(effected).Warnings(warnings), nil
}

func (self *MdbDriver) Get(params map[string]string) (commons.Result, commons.RuntimeError) {
	objectType, _ := params["mdb.type"]
	if "" == objectType {
		return nil, errutils.IsRequired("mdb.type")
	}
	definition := self.definitions.FindByUnderscoreName(objectType)
	if nil == definition {
		return nil, errutils.BadRequest("class '" + objectType + "' is not found")
	}

	id, _ := params["id"]
	switch id {
	case "", "query":
		results, err := self.FindBy(definition, params)
		if err != nil {
			return nil, commons.NewRuntimeError(commons.InternalErrorCode, "query result from db, "+err.Error())
		}
		return commons.Return(results), nil
	case "count":
		count, err := self.Count(definition, params)
		if err != nil {
			return nil, commons.NewRuntimeError(commons.InternalErrorCode, "query result from db, "+err.Error())
		}
		return commons.Return(count), nil
	}
	oid, err := parseObjectIdHex(id)
	if nil != err {
		return nil, errutils.BadRequest("id is not a objectId")
	}
	result, err := self.FindById(definition, oid, params)
	if err != nil {
		if "not found" == err.Error() {
			return nil, errutils.RecordNotFound(id)
		}

		return nil, commons.NewRuntimeError(commons.InternalErrorCode, "query result from db, "+err.Error())
	}

	return commons.Return(result), nil
}
