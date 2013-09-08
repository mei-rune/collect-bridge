package snmp

import (
	"commons"
	"commons/netutils"
	"encoding/json"
	"github.com/runner-mei/snmpclient"
	"net"
	"strings"
	"time"
)

type PingerDriver struct {
	drvMgr  *commons.DriverManager
	pingers map[string]*snmpclient.Pinger
}

func NewPingerDriver(drvMgr *commons.DriverManager) *PingerDriver {
	return &PingerDriver{drvMgr: drvMgr, pingers: make(map[string]*snmpclient.Pinger)}
}

// func (self *PingerDriver) Start(*Pinger) (err error) {
// 	if nil == self.pinger {
// 		self.pinger, err = NewPinger(self.network, self.laddr)
// 	}
// 	return
// }

// func (self *PingerDriver) Stop(*Pinger) {
// 	if nil != self.pinger {
// 		self.pinger.Close()
// 		self.pinger = nil
// 	}
// }

// func (self *PingerDriver) Reset() error {
// 	self.Stop()
// 	return self.Start()
// }

func (self *PingerDriver) Get(params map[string]string) commons.Result {
	id, ok := params["id"]
	if !ok {
		return commons.ReturnWithIsRequired("id")
	}
	pinger, ok := self.pingers[id]
	if !ok {
		return commons.ReturnWithRecordNotFound("snmp_pinger", id)
	}

	values := make([][2]string, 0, 10)
	for {
		addr, version, e := pinger.Recv(time.Duration(1))
		if nil != e {
			if commons.IsTimeout(e) {
				break
			}
			return commons.ReturnError(commons.InternalErrorCode, e.Error())
		}
		values = append(values, [2]string{addr.String(), version.String()})
	}
	return commons.Return(values)
}

func (self *PingerDriver) Put(params map[string]string) commons.Result {
	id, ok := params["id"]
	if !ok {
		return commons.ReturnWithIsRequired("id")
	}

	pinger, ok := self.pingers[id]
	if !ok {
		return commons.ReturnWithRecordNotFound("snmp_pinger", id)
	}

	port, ok := params["snmp.port"]
	if !ok {
		port = "161"
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
		return commons.ReturnWithBadRequest(
			"read body failed, it is not []string of json - " + e.Error() + body)
	}

	communities := params["snmp.communities"]
	if "" == communities {
		communities = params["snmp.community"]
	}
	if "" == communities {
		communities = "public"
	}

	versions := []snmpclient.SnmpVersion{snmpclient.SNMP_V2C, snmpclient.SNMP_V3}
	version, e := getVersion(params)
	if snmpclient.SNMP_Verr != version {
		versions = []snmpclient.SnmpVersion{version}
	}

	for _, ip_raw := range ipList {
		ip_range, e := netutils.ParseIPRange(ip_raw)
		if nil != e {
			return commons.ReturnWithInternalError(e.Error())
		}

		for i, v := range versions {
			if i != 0 {
				time.Sleep(500 * time.Millisecond)
				ip_range.Reset()
			}

			for j, community := range strings.Split(communities, ";") {
				if j != 0 {
					time.Sleep(500 * time.Millisecond)
				}

				for ip_range.HasNext() {
					e = pinger.Send(0, net.JoinHostPort(ip_range.Current().String(), port), v, community)
					if nil != e {
						return commons.ReturnWithInternalError(e.Error())
					}
				}
			}
		}
	}
	return commons.Return(true)
}

func (self *PingerDriver) Create(params map[string]string) commons.Result {
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

	pinger, err := snmpclient.NewPinger(network, address, 256)
	if nil != err {
		return commons.ReturnWithInternalError(err.Error())
	}
	self.pingers[id] = pinger
	return commons.Return(id)
}

func (self *PingerDriver) Delete(params map[string]string) commons.Result {
	id, ok := params["id"]
	if !ok {
		return commons.ReturnWithIsRequired("id")
	}
	pinger, ok := self.pingers[id]
	if !ok {
		return commons.ReturnWithRecordNotFound("snmp_pinger", id)

	}
	delete(self.pingers, id)
	pinger.Close()

	return commons.Return(true)
}
