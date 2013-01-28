package discovery

import (
	"commons"
	"commons/errutils"
	"commons/netutils"
	"net"
	"snmp"
	"time"
)

const (
	END_TOKEN     = "end"
	TIMEOUT_TOKEN = "timeout"
)

type Discoverer struct {
	ch          chan string
	command_ch  chan map[string]interface{}
	result_ch   chan commons.RuntimeError
	isCompleted bool

	icmp_pinger *netutils.Pingers
	snmp_pinger *snmp.Pinger

	devices    map[string]*Device
	ip2managed map[string]string
}

func NewDiscoverer(params *DiscoveryParams) (*Discoverer, commons.RuntimeError) {

	icmp_pinger, err := netutils.NewPingers(nil)
	if nil != err {
		return nil, errutils.InternalError("icmp failed, " + err.Error())
	}

	snmp_pinger, err := snmp.NewPinger("udp4", "")
	if nil != err {
		return nil, errutils.InternalError("snmp failed, " + err.Error())
	}

	discoverer := &Discoverer{ch: make(chan string, 1000),
		command_ch:  make(chan map[string]interface{}),
		result_ch:   make(chan commons.RuntimeError),
		icmp_pinger: icmp_pinger,
		snmp_pinger: snmp_pinger,

		devices:    make(map[string]*Device),
		ip2managed: make(map[string]string)}

	go discoverer.serve()

	return discoverer, nil
}

func (self *Discoverer) serve() {
	for i, f := range net.Interfaces() {
		f.Addrs()
	}
}

func (self *Discoverer) Result() map[string]interface{} {
	return self.devices
}

func (self *Discoverer) IsCompleted() bool {
	return self.isCompleted
}

func (self *Discoverer) Control(params map[string]interface{}) commons.RuntimeError {
	self.command_ch <- params
	return <-self.result_ch
}

func (self *Discoverer) Read(timeout time.Duration) string {
	select {
	case res := <-self.ch:
		if END_TOKEN == res {
			self.isCompleted = true
		}
		return res
	case <-time.After(timeout):
		return TIMEOUT_TOKEN
	}
	return ""
}

func (self *Discoverer) Close() {
	self.icmp_pinger.Close()
	self.snmp_pinger.Close()
}
