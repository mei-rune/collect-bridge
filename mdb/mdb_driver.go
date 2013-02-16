package mdb

import (
	"commons"
	"commons/errutils"
	"encoding/json"
	"errors"
	"labix.org/v2/mgo"
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

func (self *MdbDriver) Create(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	objectType, _ := params["mdb.type"]
	if "" == objectType {
		return nil, errutils.IsRequired("mdb.type")
	}
	definition := self.definitions.FindByUnderscoreName(objectType)
	if nil == definition {
		return nil, commons.NewRuntimeError(commons.InternalErrorCode, "class '"+objectType+"' is not found")
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

	instance_id, err := self.mdb_server.Create(definition, result)
	if err != nil {
		return nil, commons.NewRuntimeError(commons.InternalErrorCode, "insert object to db, "+err.Error())
	}

	return commons.Return(instance_id), nil
}

func (self *MdbDriver) Put(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	objectType, _ := params["mdb.type"]
	if "" == objectType {
		return nil, errutils.IsRequired("mdb.type")
	}
	id, _ := params["id"]
	if "" == id {
		return nil, commons.IdNotExists
	}

	oid, err := parseObjectIdHex(id)
	if nil != err {
		return nil, errutils.BadRequest("id is not a objectId")
	}

	var result map[string]interface{}
	definition := self.definitions.FindByUnderscoreName(objectType)
	if nil == definition {
		return nil, commons.NewRuntimeError(commons.InternalErrorCode, "class '"+objectType+"' is not found")
	}

	body, ok := params["body"]
	if !ok {
		return nil, commons.BodyNotExists
	}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return nil, commons.NewRuntimeError(commons.InternalErrorCode, "unmarshal object from request failed, "+err.Error())
	}

	err = self.mdb_server.Update(definition, oid, result)
	if err != nil {
		return nil, commons.NewRuntimeError(commons.InternalErrorCode, "update object to db, "+err.Error())
	}

	return commons.ReturnOK(), nil
}

func (self *MdbDriver) Delete(params map[string]string) (bool, commons.RuntimeError) {
	objectType, _ := params["mdb.type"]
	if "" == objectType {
		return false, errutils.IsRequired("mdb.type")
	}
	id, _ := params["id"]
	if "" == id {
		return false, commons.IdNotExists
	}

	oid, err := parseObjectIdHex(id)
	if nil != err {
		return false, errutils.BadRequest("id is not a objectId")
	}

	definition := self.definitions.FindByUnderscoreName(objectType)
	if nil == definition {
		return false, commons.NewRuntimeError(commons.InternalErrorCode, "class '"+objectType+"' is not found")
	}

	ok, err := self.RemoveById(definition, oid)
	if !ok {
		return false, commons.NewRuntimeError(commons.InternalErrorCode, "remove object from db failed, "+err.Error())
	}

	return true, nil
}

func (self *MdbDriver) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	objectType, _ := params["mdb.type"]
	if "" == objectType {
		return nil, errutils.IsRequired("mdb.type")
	}
	definition := self.definitions.FindByUnderscoreName(objectType)
	if nil == definition {
		return nil, commons.NewRuntimeError(commons.InternalErrorCode, "class '"+objectType+"' is not found")
	}

	id, _ := params["id"]
	if "" == id {
		results, err := self.FindBy(definition, params)
		if err != nil {
			return nil, commons.NewRuntimeError(commons.InternalErrorCode, "query result from db, "+err.Error())
		}
		return commons.Return(results), nil
	}

	oid, err := parseObjectIdHex(id)
	if nil != err {
		return nil, errutils.BadRequest("id is not a objectId")
	}
	result, err := self.FindById(definition, oid)
	if err != nil {
		return nil, commons.NewRuntimeError(commons.InternalErrorCode, "query result from db, "+err.Error())
	}

	return commons.Return(result), nil
}
