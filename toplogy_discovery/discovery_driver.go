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

	ret, err := commons.ParseDuration(v)
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
		return commons.ReturnWithError(commons.IdNotExists)
	}
	discoverer, ok := self.discoverers[id]
	if !ok {
		return commons.ReturnWithError(commons.RecordNotFound(id))
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
	return commons.ReturnError(503, "discovering, try again later.")
}

func (self *DiscoveryDriver) Put(params map[string]string) commons.Result {
	id, ok := params["id"]
	if !ok {
		return commons.ReturnWithError(commons.IdNotExists)
	}
	discoverer, ok := self.discoverers[id]
	if !ok {
		return commons.ReturnWithError(commons.RecordNotFound(id))
	}

	body, ok := params["body"]
	if !ok {
		return commons.ReturnWithError(commons.BodyNotExists)
	}
	if "" == body {
		return commons.ReturnWithError(commons.BodyIsEmpty)
	}
	params2 := make(map[string]interface{})
	e := json.Unmarshal([]byte(body), &params2)
	if nil != e {
		return commons.ReturnWithError(commons.BadRequest("read body failed, it is not a map[string]interface{} - " + e.Error() + body))
	}
	for k, v := range params {
		params2[k] = v
	}

	err := discoverer.Control(params2)
	if nil != err {
		return commons.ReturnWithError(err)
	}
	return commons.Return(true)
}

func (self *DiscoveryDriver) Create(params map[string]string) commons.Result {
	body, _ := params["body"]
	if "" == body {
		return commons.ReturnWithError(commons.BodyIsEmpty)
	}

	discovery_params := &DiscoveryParams{}
	e := json.Unmarshal([]byte(body), discovery_params)
	if nil != e {
		return commons.ReturnWithError(commons.BadRequest("read body failed, it is not DiscoveryParams - " + e.Error()))
	}

	discoverer, err := NewDiscoverer(discovery_params, self.drvMgr)
	if nil != err {
		return commons.ReturnError(500, err.Error())
	}
	id := time.Now().UTC().String()
	if _, ok := self.discoverers[id]; ok {
		return commons.ReturnWithError(commons.ServiceUnavailable)
	}
	self.discoverers[id] = discoverer
	return commons.Return(id)
}

func (self *DiscoveryDriver) Delete(params map[string]string) commons.Result {
	id, ok := params["id"]
	if !ok {
		return commons.ReturnWithError(commons.IdNotExists)
	}
	discoverer, ok := self.discoverers[id]
	if !ok {
		return commons.ReturnWithError(commons.RecordNotFound(id))
	}

	delete(self.discoverers, id)
	discoverer.Close()

	return commons.Return(true)
}
