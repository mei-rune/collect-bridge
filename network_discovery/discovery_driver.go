package discovery

import (
	"commons"
	"encoding/json"
	"time"
)

func getTimeout(params map[string]string, timeout time.Duration) time.Duration {
	v, ok := params["timeout"]
	if !ok {
		return timeout
	}

	ret, err := time.ParseDuration(v)
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

func (self *DiscoveryDriver) Get(params map[string]string) commons.Result {
	id, ok := params["id"]
	if !ok {
		return commons.ReturnWithBadRequest(commons.IdNotExists.Error())
	}
	discoverer, ok := self.discoverers[id]
	if !ok {
		return commons.ReturnWithNotFound("discovery", id)
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
		return commons.Return(messages)
	}
	if discoverer.IsCompleted() {
		return commons.Return(discoverer.Result())
	}
	return commons.ReturnWithServiceUnavailable("discovering, try again later.")
}

func (self *DiscoveryDriver) Put(params map[string]string) commons.Result {
	id, ok := params["id"]
	if !ok {
		return commons.ReturnWithBadRequest(commons.IdNotExists.Error())
	}
	discoverer, ok := self.discoverers[id]
	if !ok {
		return commons.ReturnWithNotFound("discovery", id)
	}

	body, ok := params["body"]
	if !ok {
		return commons.ReturnWithBadRequest(commons.BodyNotExists.Error())
	}
	if "" == body {
		return commons.ReturnWithBadRequest(commons.BodyIsEmpty.Error())
	}
	params2 := make(map[string]interface{})
	e := json.Unmarshal([]byte(body), &params2)
	if nil != e {
		return commons.ReturnWithBadRequest("read body failed, it is not a map[string]interface{} - " + e.Error() + body)
	}
	for k, v := range params {
		params2[k] = v
	}

	err := discoverer.Control(params2)
	if nil != err {
		return commons.ReturnWithInternalError(err.Error())
	}
	return commons.Return(true)
}

func (self *DiscoveryDriver) Create(params map[string]string) commons.Result {
	body, _ := params["body"]
	if "" == body {
		return commons.ReturnWithBadRequest(commons.BodyIsEmpty.Error())
	}

	discovery_params := &DiscoveryParams{}
	e := json.Unmarshal([]byte(body), discovery_params)
	if nil != e {
		return commons.ReturnWithBadRequest("read body failed, it is not DiscoveryParams - " + e.Error())
	}

	discoverer, err := NewDiscoverer(discovery_params, self.drvMgr)
	if nil != err {
		return commons.ReturnWithInternalError(err.Error())
	}
	id := time.Now().UTC().String()
	if _, ok := self.discoverers[id]; ok {
		return commons.ReturnWithServiceUnavailable("discovering, try again later.")
	}
	self.discoverers[id] = discoverer
	return commons.Return(id)
}

func (self *DiscoveryDriver) Delete(params map[string]string) commons.Result {
	id, ok := params["id"]
	if !ok {
		return commons.ReturnWithBadRequest(commons.IdNotExists.Error())
	}
	discoverer, ok := self.discoverers[id]
	if !ok {
		return commons.ReturnWithNotFound("discovery", id)
	}

	delete(self.discoverers, id)
	discoverer.Close()

	return commons.Return(true)
}
