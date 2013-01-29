package discovery

import (
	"commons"
	"commons/errutils"
	"commons/netutils"
	"errors"
	"fmt"
	"log"
	"net"
	"snmp"
	"strings"
	"time"
)

const (
	DEBUG         = "debug"
	WARN          = "warn"
	ERROR         = "error"
	FATAL         = "fatal"
	INFO          = "info"
	END_TOKEN     = "end"
	TIMEOUT_TOKEN = "timeout"
)

type replyResult struct {
	ok  string
	err error
}
type Discoverer struct {
	ch          chan string
	command_ch  chan map[string]interface{}
	result_ch   chan commons.RuntimeError
	isCompleted bool

	params *DiscoveryParams

	icmp_pinger *netutils.Pingers
	snmp_pinger *snmp.Pinger
	snmp_drv    commons.Driver
	metrics_drv commons.Driver

	devices    map[string]*Device
	ip2managed map[string]string
}

func NewDiscoverer(params *DiscoveryParams, drvMgr *commons.DriverManager) (*Discoverer, commons.RuntimeError) {

	if nil == params {
		return nil, errutils.InternalError("params is nil.")
	}

	if nil == params.Communities {
		params.Communities = []string{"public"}
	} else {
		isFound := false
		for _, s := range params.Communities {
			if s == "public" {
				isFound = true
			}
		}

		if !isFound {
			params.Communities = append(params.Communities, "public")
		}
	}

	icmp_pinger, err := netutils.NewPingers(nil, 10000)
	if nil != err {
		return nil, errutils.InternalError("icmp failed, " + err.Error())
	}

	snmp_pinger, err := snmp.NewPinger("udp4", "0.0.0.0:0", 10000)
	if nil != err {
		return nil, errutils.InternalError("snmp failed, " + err.Error())
	}

	snmp_drv, ok := drvMgr.Connect("snmp")
	if !ok {
		return nil, errutils.InternalError("snmp failed, driver is not found.")
	}

	metrics_drv, ok := drvMgr.Connect("metrics")
	if !ok {
		return nil, errutils.InternalError("metrics failed, driver is not found.")
	}

	discoverer := &Discoverer{ch: make(chan string, 1000),
		command_ch:  make(chan map[string]interface{}),
		result_ch:   make(chan commons.RuntimeError),
		icmp_pinger: icmp_pinger,
		snmp_pinger: snmp_pinger,
		snmp_drv:    snmp_drv,
		metrics_drv: metrics_drv,

		params: params,

		devices:    make(map[string]*Device),
		ip2managed: make(map[string]string)}

	go discoverer.serve()

	return discoverer, nil
}

func (self *Discoverer) readMetric(drv *Device, name string) (interface{}, error) {

	if nil == drv.Communities || 0 == len(drv.Communities) {
		return nil, errors.New("community ip of local host is empty.")
	}

	res, e := self.metrics_drv.Get(map[string]string{"id": drv.ManagedIP, "metric": name,
		"snmp.community": drv.Communities[0]})
	if nil != e {
		return nil, e
	}
	return commons.GetReturn(res), nil
}

func (self *Discoverer) guessCommunities(ip string) []string {

	valid := make([]string, 0, len(self.params.Communities))
	ch := make(chan string)
	defer func() { close(ch) }()

	for _, c := range self.params.Communities {
		go func(h chan string, cm string, params map[string]string) {

			defer func() {
				if err := recover(); nil != err {
					log.Printf("guess community - %s, %v", cm, err)
				}
			}()
			_, e := self.snmp_drv.Get(params)
			if nil != e {
				h <- cm
			}
		}(ch, c, map[string]string{"id": ip + "/1.3.6.1.2.1.1.2.0", "action": "get", "community": c, "timeout": "30"})
	}

	self.log(DEBUG, "guess password for "+ip)
	for len(valid) != len(self.params.Communities) {
		select {
		case c := <-ch:
			valid = append(valid, c)
		case <-time.After(1 * time.Minute):
			break
		}
	}

	self.log(DEBUG, "guess password for "+ip+", result is "+strings.Join(valid, ","))
	return valid
}

func (self *Discoverer) readLocal() (*Device, error) {
	managed_ip := ""
	interfaces := make([]Interface, 0)
	ifs, e := net.Interfaces()
	if nil != e {
		return nil, e
	}

	for _, f := range ifs {
		addrs, e := f.Addrs()
		if nil != e {
			return nil, e
		}

		for _, addr := range addrs {
			if "" == managed_ip {
				if !netutils.IsInvalidAddress(addr.String()) {
					managed_ip = addr.String()
				}
			}
			interfaces = append(interfaces, Interface{Index: f.Index,
				MTU:          f.MTU,
				Description:  f.Name,
				Address:      addr.String(),
				HardwareAddr: f.HardwareAddr.String()})
		}
	}

	if "" == managed_ip {
		return nil, errors.New("managed ip of local host is empty.")
	}

	communities := self.guessCommunities(managed_ip)
	if nil == communities || 0 == len(communities) {
		return nil, errors.New("community ip of local host is empty.")
	}

	drv := &Device{ManagedIP: managed_ip, Communities: communities, Interfaces: interfaces,
		Attributes: map[string]interface{}{}}
	self.initDevice(drv)
	return drv, nil
}

func (self *Discoverer) discoverDevice(ip string) (*Device, error) {
	if "" == ip {
		return nil, errors.New("managed ip of local host is empty.")
	}

	communities := self.guessCommunities(ip)
	if nil == communities || 0 == len(communities) {
		return nil, errors.New("community ip of local host is empty.")
	}

	drv := &Device{ManagedIP: ip, Communities: communities}
	return drv, self.initDevice(drv)
}

func (self *Discoverer) initDevice(drv *Device) error {
	// read basic attributes
	drv.Attributes = map[string]interface{}{}
	for _, name := range []string{"system.oid", "system.descr"} {
		self.logf(DEBUG, "read '%s' for ip", name, drv.ManagedIP)
		metric, e := self.readMetric(drv, name)
		if nil != e {
			self.logf(ERROR, "read %s of '%s' failed, %v.", name, drv.ManagedIP, e)
		} else if nil == metric {
			self.logf(ERROR, "read %s of '%s' failed, result is nil.", name, drv.ManagedIP)
		} else {
			self.logf(DEBUG, "read '%s' for ip, result is %v", name, drv.ManagedIP, metric)
			drv.Attributes[name] = metric
		}
	}

	return nil
}

func (self *Discoverer) log(level string, message string) {
	self.ch <- level + " " + message
}

func (self *Discoverer) logf(level string, format string, params ...interface{}) {
	self.ch <- fmt.Sprintf(level+" "+format, params...)
}

func (self *Discoverer) addDevice(drv *Device) {
	self.devices[drv.ManagedIP] = drv
	if nil == drv.Interfaces {
		return
	}
	for _, ifs := range drv.Interfaces {
		self.ip2managed[ifs.Address] = drv.ManagedIP
	}
}

func (self *Discoverer) scanIPRange(drv *Device) {

	for _, ifs := range drv.Interfaces {

		ip_range, err := netutils.ParseIPRange(ifs.Address + "/24")
		if nil != err {
			self.log(DEBUG, "scan ip range failed, "+err.Error())
			continue
		}

		for ip_range.HasNext() {
			addr := ip_range.Current().String()
			err = self.icmp_pinger.Send(addr, nil)
			if nil != err {
				self.log(DEBUG, "send icmp to '"+addr+"'' failed, "+err.Error())
			}

			err = self.snmp_pinger.Send(addr, snmp.SNMP_V2C)
			if nil != err {
				self.log(DEBUG, "send snmp to '"+addr+"'' failed, "+err.Error())
			}
		}
	}
}

func (self *Discoverer) isNewDevice(ip string) bool {
	_, ok := self.devices[ip]
	if ok {
		return false
	}
	_, ok = self.ip2managed[ip]
	if ok {
		return false
	}
	return true
}

func (self *Discoverer) serve() {
	defer func() {
		self.ch <- END_TOKEN
	}()

	local, e := self.readLocal()
	if nil != e {
		self.log(FATAL, e.Error())
		return
	}

	self.addDevice(local)
	self.log(INFO, "discover device '"+local.ManagedIP+"'")

	for {
		ra, _, err := self.icmp_pinger.Recv(1 * time.Minute)
		if nil == ra {
			if nil != err && !commons.IsTimeout(err) {
				self.log(DEBUG, "recv icmp failed - "+err.Error())
			}
			ra, _, err = self.snmp_pinger.Recv(1 * time.Minute)
			if nil == ra {
				if nil != err && !commons.IsTimeout(err) {
					self.log(DEBUG, "recv icmp failed - "+err.Error())
				}
				break
			}
		}

		if netutils.IsInvalidAddress(ra.String()) {
			self.log(DEBUG, "skip invalid address - "+ra.String())
			continue
		}

		if !self.isNewDevice(ra.String()) {
			self.log(DEBUG, "skip old address - "+ra.String())
			continue
		}

		drv, e := self.discoverDevice(ra.String())
		if nil != e {
			self.log(ERROR, "discover device '"+ra.String()+"' failed, "+e.Error())
			continue
		}
		self.log(INFO, "discover device '"+ra.String()+"'")
		self.addDevice(drv)
		self.scanIPRange(drv)
	}
}

func (self *Discoverer) Result() map[string]interface{} {
	result := map[string]interface{}{}
	for k, v := range self.devices {
		result[k] = v
	}
	return result
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
