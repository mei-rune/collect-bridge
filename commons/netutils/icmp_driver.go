package netutils

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

type ICMPDriver struct {
	drvMgr  *commons.DriverManager
	pingers map[string]*Pinger
}

func NewICMPDriver(drvMgr *commons.DriverManager) *ICMPDriver {
	return &ICMPDriver{drvMgr: drvMgr, pingers: make(map[string]*Pinger)}
}

func (self *ICMPDriver) Get(params map[string]string) commons.Result {
	id, ok := params["id"]
	if !ok {
		return commons.ReturnWithIsRequired("id")
	}
	pinger, ok := self.pingers[id]
	if !ok {
		return commons.ReturnWithNotFound("icmp_pinger", id)
	}

	values := make([]string, 0, 10)
	for {
		addr, _, e := pinger.Recv(time.Duration(1))
		if nil != e {
			if commons.IsTimeout(e) {
				break
			}
			return commons.Return(values).SetError(commons.InterruptErrorCode, e.Error())
		}
		values = append(values, addr.String())
	}
	return commons.Return(values)
}

func (self *ICMPDriver) Put(params map[string]string) commons.Result {
	id, ok := params["id"]
	if !ok {
		return commons.ReturnWithIsRequired("id")
	}
	pinger, ok := self.pingers[id]
	if !ok {
		return commons.ReturnWithNotFound("icmp_pinger", id)
	}

	body, ok := params["body"]
	if !ok {
		return commons.ReturnWithIsRequired("body")
	}
	if "" == body {
		return commons.ReturnWithIsRequired("body")
	}
	ipList := make([]string, 0, 100)
	e := json.Unmarshal([]byte(body), &ipList)
	if nil != e {
		return commons.ReturnWithBadRequest("read body failed, it is not []string of json - " + e.Error() + body)
	}

	for _, ip_raw := range ipList {

		ip_range, e := ParseIPRange(ip_raw)
		if nil != e {
			return commons.ReturnWithInternalError(e.Error())
		}

		for ip_range.HasNext() {
			e = pinger.Send(ip_range.Current().String(), nil)
			if nil != e {
				return commons.ReturnWithInternalError(e.Error())
			}
		}
	}
	return commons.Return(true)
}

func (self *ICMPDriver) Create(params map[string]string) commons.Result {
	body, _ := params["body"]
	if "" == body {
		body = "{}"
	}

	params2 := make(map[string]string)
	e := json.Unmarshal([]byte(body), &params2)
	if nil != e {
		return commons.ReturnWithBadRequest("read body failed, it is not map[string]string of json - " + e.Error())
	}
	network, _ := params2["network"]
	if "" == network {
		network, _ = params["network"]
		if "" == network {
			return commons.ReturnWithIsRequired("network")
		}
	}

	address, _ := params2["address"]
	if "" == address {
		address, _ = params["address"]
	}

	id := network + "," + address
	_, ok := self.pingers[id]
	if ok {
		return commons.ReturnWithRecordAlreadyExists(id)
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
		return commons.ReturnWithInternalError(err.Error())
	}
	self.pingers[id] = icmp
	return commons.Return(id)
}

func (self *ICMPDriver) Delete(params map[string]string) commons.Result {
	id, ok := params["id"]
	if !ok {
		return commons.ReturnWithIsRequired("id")
	}
	pinger, ok := self.pingers[id]
	if !ok {
		return commons.ReturnWithNotFound("icmp_pinger", id)
	}
	delete(self.pingers, id)
	pinger.Close()

	return commons.Return(true)
}
