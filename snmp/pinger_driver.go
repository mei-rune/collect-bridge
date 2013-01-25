package snmp

import (
	"commons"
	"commons/netutils"
	"net"
	"time"
)

type PingerDriver struct {
	drvMgr         *commons.DriverManager
	pinger         *Pinger
	network, laddr string
}

func NewPingerDriver(drvMgr *commons.DriverManager, network, laddr string) *PingerDriver {
	return &PingerDriver{drvMgr: drvMgr}
}

func (self *PingerDriver) Start() (err error) {
	if nil == self.pinger {
		self.pinger, err = NewPinger(self.network, self.laddr)
	}
	return
}

func (self *PingerDriver) Stop() {
	if nil != self.pinger {
		self.pinger.Close()
		self.pinger = nil
	}
}

func (self *PingerDriver) Reset() error {
	self.Stop()
	return self.Start()
}

func (self *PingerDriver) Get(params map[string]string) (map[string]interface{}, error) {
	values := make([][2]string, 0, 10)
	for {
		addr, version, e := self.pinger.Recv(time.Duration(1))
		if nil != e {
			if commons.IsTimeout(e) {
				break
			}
			e = self.Reset()
			if nil != e {
				return map[string]interface{}{"value": values}, e
			}
			break
		}
		values = append(values, [2]string{addr.String(), version.String()})
	}
	return map[string]interface{}{"value": values}, nil
}

func (self *PingerDriver) Put(params map[string]string) (map[string]interface{}, error) {
	id, ok := params["id"]
	if !ok {
		return nil, commons.IdNotFound
	}
	port, ok := params["port"]
	if !ok {
		port = "161"
	}

	ip_range, e := netutils.ParseIPRange(id)
	if nil != e {
		return nil, e
	}
	versions := []SnmpVersion{SNMP_V2C, SNMP_V3}
	version, e := getVersion(params)
	if SNMP_Verr != version {
		versions = []SnmpVersion{version}
	}

	for i, v := range versions {
		if i != 0 {
			time.Sleep(500 * time.Millisecond)
			ip_range.Reset()
		}

		for ip_range.HasNext() {
			e = self.pinger.Send(net.JoinHostPort(ip_range.Current().String(), port), v)
			if nil != e {
				e = self.Reset()
				if nil != e {
					return nil, e
				}
			}
		}
	}
	return map[string]interface{}{"value": "ok"}, nil
}

func (self *PingerDriver) Create(params map[string]string) (bool, error) {
	return false, commons.NotImplemented
}

func (self *PingerDriver) Delete(params map[string]string) (bool, error) {
	return false, commons.NotImplemented
}
