package poller

import (
	"commons"
	"commons/errutils"
	"errors"
	"fmt"
	"mdb"
)

type KPIDriver struct {
	managedObjects map[string]map[string]interface{}
	Warnings       interface{}
	client         *commons.HttpClient
}

func NewKPIDriver(baseUrl string, client *mdb.Client) (commons.Driver, error) {
	res, err := client.FindByWithIncludes("device", nil, "snmp_params")
	if nil != err {
		return nil, errors.New("load device failed, " + err.Error())
	}
	managedObjects := make(map[string]map[string]interface{})
	for _, mo := range res {
		managedObjects[fmt.Sprintf("device-%v", mo["_id"])] = mo
	}

	return &KPIDriver{managedObjects: managedObjects, client: &commons.HttpClient{Url: baseUrl}}, nil
}

func makeId(t, id string) string {
	return t + "-" + id
}

func (self *KPIDriver) Get(params map[string]string) (commons.Result, commons.RuntimeError) {
	id := makeId(params["managed_type"], params["managed_id"])
	mo := self.managedObjects[id]
	if nil == mo {
		return nil, commons.NotFound(id)
	}

	access_params, e := commons.TryGetObjects(mo, "$snmp_params")
	if nil != e {
		return nil, errutils.InternalError(fmt.Sprintf("fetch access params failed - %v", e))
	}
	if nil == access_params || 0 == len(access_params) {
		return nil, errutils.InternalError("access params is not exists.")
	}

	snmp_params := access_params[0]
	if "snmp_params" != snmp_params["type"] {
		return nil, errutils.InternalError("get access params failed - it is not a snmp params")
	}

	if charset := params["charset"]; "" == charset {
		params["charset"] = "gb18030"
	}
	url := self.client.CreateUrl().Concat("metric", params["metric"], id).WithQueries(params, "").WithAnyQueries(snmp_params, "snmp.").ToUrl()
	return self.client.Invoke("GET", url, nil, 200)
}

func (self *KPIDriver) Put(params map[string]string) (commons.Result, commons.RuntimeError) {
	return nil, commons.NotImplemented
}

func (self *KPIDriver) Create(params map[string]string) (commons.Result, commons.RuntimeError) {
	return nil, commons.NotImplemented
}

func (self *KPIDriver) Delete(params map[string]string) (commons.Result, commons.RuntimeError) {
	return nil, commons.NotImplemented
}
