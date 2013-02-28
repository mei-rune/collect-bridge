package netutils

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

type ICMPDriver struct {
	drvMgr  *commons.DriverManager
	pingers map[string]*Pinger
}

func NewICMPDriver(drvMgr *commons.DriverManager) *ICMPDriver {
	return &ICMPDriver{drvMgr: drvMgr, pingers: make(map[string]*Pinger)}
}

func (self *ICMPDriver) Get(params map[string]string) (commons.Result, commons.RuntimeError) {
	id, ok := params["id"]
	if !ok {
		return nil, commons.IdNotExists
	}
	pinger, ok := self.pingers[id]
	if !ok {
		return nil, errutils.RecordNotFound(id)
	}

	values := make([]string, 0, 10)
	for {
		addr, _, e := pinger.Recv(time.Duration(1))
		if nil != e {
			if commons.IsTimeout(e) {
				break
			}
			return commons.Return(values), commons.NewRuntimeError(500, e.Error())
		}
		values = append(values, addr.String())
	}
	return commons.Return(values), nil
}

func (self *ICMPDriver) Put(params map[string]string) (commons.Result, commons.RuntimeError) {
	id, ok := params["id"]
	if !ok {
		return nil, commons.IdNotExists
	}
	pinger, ok := self.pingers[id]
	if !ok {
		return nil, errutils.RecordNotFound(id)
	}

	body, ok := params["body"]
	if !ok {
		return nil, commons.BodyNotExists
	}
	if "" == body {
		return nil, errutils.IsRequired("body")
	}
	ipList := make([]string, 0, 100)
	e := json.Unmarshal([]byte(body), &ipList)
	if nil != e {
		return nil, errutils.BadRequest("read body failed, it is not []string of json - " + e.Error() + body)
	}

	for _, ip_raw := range ipList {

		ip_range, e := ParseIPRange(ip_raw)
		if nil != e {
			return nil, commons.NewRuntimeError(500, e.Error())
		}

		for ip_range.HasNext() {
			e = pinger.Send(ip_range.Current().String(), nil)
			if nil != e {
				return nil, commons.NewRuntimeError(500, e.Error())
			}
		}
	}
	return commons.Return(true), nil
}

func (self *ICMPDriver) Create(params map[string]string) (commons.Result, commons.RuntimeError) {
	body, _ := params["body"]
	if "" == body {
		body = "{}"
	}

	params2 := make(map[string]string)
	e := json.Unmarshal([]byte(body), &params2)
	if nil != e {
		return nil, errutils.BadRequest("read body failed, it is not map[string]string of json - " + e.Error())
	}
	network, _ := params2["network"]
	if "" == network {
		network, _ = params["network"]
		if "" == network {
			return nil, errutils.IsRequired("network")
		}
	}

	address, _ := params2["address"]
	if "" == address {
		address, _ = params["address"]
	}

	id := network + "," + address
	_, ok := self.pingers[id]
	if ok {
		return nil, errutils.RecordAlreadyExists(id)
	}

	echo, _ := params2["echo"]
	if "" == echo {
		echo, _ = params["echo"]
		if "" == echo {
			echo = "mfk"
		}
	}

	icmp, err := NewPinger(network, address, []byte(echo), 256)
	if nil != err {
		return nil, commons.NewRuntimeError(500, err.Error())
	}
	self.pingers[id] = icmp
	return commons.Return(id), nil
}

func (self *ICMPDriver) Delete(params map[string]string) (commons.Result, commons.RuntimeError) {
	id, ok := params["id"]
	if !ok {
		return nil, commons.IdNotExists
	}
	pinger, ok := self.pingers[id]
	if !ok {
		return nil, errutils.RecordNotFound(id)
	}
	delete(self.pingers, id)
	pinger.Close()

	return commons.Return(true), nil
}
