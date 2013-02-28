package discovery

import (
	"commons"
	"commons/errutils"
	"encoding/json"
	"time"
)

func getTimeout(params map[string]string, timeout time.Duration) time.Duration {
	v, ok := params["timeout"]
	if !ok {
		return timeout
	}

	ret, err := commons.ParseTime(v)
	if nil != err {
		panic(err)
	}
	return ret
}

type DiscoveryDriver struct {
	drvMgr      *commons.DriverManager
	discoverers map[string]*Discoverer
}

func NewDiscoveryDriver(timeout time.Duration, drvMgr *commons.DriverManager) *DiscoveryDriver {
	return &DiscoveryDriver{drvMgr: drvMgr, discoverers: make(map[string]*Discoverer)}
}

func (self *DiscoveryDriver) Get(params map[string]string) (commons.Result, commons.RuntimeError) {
	id, ok := params["id"]
	if !ok {
		return nil, commons.IdNotExists
	}
	discoverer, ok := self.discoverers[id]
	if !ok {
		return nil, errutils.RecordNotFound(id)
	}

	dst, ok := params["dst"]
	if "message" == dst {
		messages := make([]string, 0, 10)
		for !discoverer.IsCompleted() {
			message := discoverer.Read(1 * time.Second)
			if TIMEOUT_TOKEN == message {
				break
			}
			messages = append(messages, message)
		}
		return commons.Return(messages), nil
	}
	if discoverer.IsCompleted() {
		return commons.Return(discoverer.Result()), nil
	}
	return nil, commons.NewRuntimeError(503, "discovering, try again later.")
}

func (self *DiscoveryDriver) Put(params map[string]string) (commons.Result, commons.RuntimeError) {
	id, ok := params["id"]
	if !ok {
		return nil, commons.IdNotExists
	}
	discoverer, ok := self.discoverers[id]
	if !ok {
		return nil, errutils.RecordNotFound(id)
	}

	body, ok := params["body"]
	if !ok {
		return nil, commons.BodyNotExists
	}
	if "" == body {
		return nil, commons.BodyIsEmpty
	}
	params2 := make(map[string]interface{})
	e := json.Unmarshal([]byte(body), &params2)
	if nil != e {
		return nil, errutils.BadRequest("read body failed, it is not a map[string]interface{} - " + e.Error() + body)
	}
	for k, v := range params {
		params2[k] = v
	}

	err := discoverer.Control(params2)
	if nil != err {
		return nil, err
	}
	return commons.Return(true), nil
}

func (self *DiscoveryDriver) Create(params map[string]string) (commons.Result, commons.RuntimeError) {
	body, _ := params["body"]
	if "" == body {
		return nil, commons.BodyIsEmpty
	}

	discovery_params := &DiscoveryParams{}
	e := json.Unmarshal([]byte(body), discovery_params)
	if nil != e {
		return nil, errutils.BadRequest("read body failed, it is not DiscoveryParams - " + e.Error())
	}

	discoverer, err := NewDiscoverer(discovery_params, self.drvMgr)
	if nil != err {
		return nil, commons.NewRuntimeError(500, err.Error())
	}
	id := time.Now().UTC().String()
	if _, ok := self.discoverers[id]; ok {
		return nil, commons.ServiceUnavailable
	}
	self.discoverers[id] = discoverer
	return commons.Return(id), nil
}

func (self *DiscoveryDriver) Delete(params map[string]string) (commons.Result, commons.RuntimeError) {
	id, ok := params["id"]
	if !ok {
		return nil, commons.IdNotExists
	}
	discoverer, ok := self.discoverers[id]
	if !ok {
		return nil, errutils.RecordNotFound(id)
	}
	delete(self.discoverers, id)
	discoverer.Close()

	return commons.Return(true), nil
}
