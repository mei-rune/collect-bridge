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
	DEBUG         = "[DEBUG]"
	WARN          = "[WARN]"
	ERROR         = "[ERROR]"
	FATAL         = "[FATAL]"
	INFO          = "[INFO]"
	END_TOKEN     = "end"
	TIMEOUT_TOKEN = "timeout"
)

type replyResult struct {
	ok  string
	err error
}
type Discoverer struct {
	ch          chan string
	drv_ch      chan Device
	command_ch  chan map[string]interface{}
	result_ch   chan commons.RuntimeError
	isCompleted bool

	params *DiscoveryParams

	//icmp_pinger *netutils.Pingers
	snmp_pinger *snmp.Pinger
	snmp_drv    commons.Driver
	metrics_drv commons.Driver

	devices    map[string]Device
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

	// icmp_pinger, err := netutils.NewPingers(nil, 10000)
	// if nil != err {
	//	return nil, errutils.InternalError("icmp failed, " + err.Error())
	// }

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
		drv_ch:     make(chan Device),
		command_ch: make(chan map[string]interface{}),
		result_ch:  make(chan commons.RuntimeError),
		//icmp_pinger: icmp_pinger,
		snmp_pinger: snmp_pinger,
		snmp_drv:    snmp_drv,
		metrics_drv: metrics_drv,

		params: params,

		devices:    make(map[string]Device),
		ip2managed: make(map[string]string)}

	go discoverer.serve()

	return discoverer, nil
}

func (self *Discoverer) readMetric(drv Device, name string) (interface{}, error) {
	communities := drv["communities"].([]string)
	if nil == communities || 0 == len(communities) {
		return nil, errors.New("community ip of local host is empty.")
	}

	res, e := self.metrics_drv.Get(map[string]string{"id": drv["address"].(string), "metric": name,
		"snmp.community": communities[0], "charset": "GB18030"})
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
			if nil == e {
				h <- cm
			} else {
				h <- ""
			}
		}(ch, c, map[string]string{"id": ip, "snmp.oid": "1.3.6.1.2.1.1.2.0", "snmp.action": "get", "snmp.community": c, "timeout": "30"})
	}

	tries := len(self.params.Communities)
	for 0 != tries {
		select {
		case c := <-ch:
			if "" != c {
				valid = append(valid, c)
			}
			tries--
		case <-time.After(1 * time.Minute):
			goto End
		}
	}
End:
	if 0 == len(valid) {
		self.log(DEBUG, "guess password for "+ip+", result is empty")
	} else {
		self.log(DEBUG, "guess password for "+ip+", result is "+strings.Join(valid, ","))
	}
	return valid
}

func (self *Discoverer) readLocal() (Device, error) {
	managed_ip := ""
	interfaces := make([]interface{}, 0)
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
			interfaces = append(interfaces, map[string]interface{}{"ifIndex": f.Index,
				"ifMtu":         f.MTU,
				"ifDescr":       f.Name,
				"address":       addr.String(),
				"ifPhysAddress": f.HardwareAddr.String()})
		}
	}

	if "" == managed_ip {
		managed_ip = "127.0.0.1"
		//return nil, errors.New("managed ip of local host is empty.")
	}

	communities := self.guessCommunities(managed_ip)
	if nil == communities || 0 == len(communities) {
		return nil, errors.New("community of localhost is empty.")
	}

	drv := Device{"address": managed_ip, "communities": communities, "interfaceTables": interfaces}
	self.initDevice(drv)
	return drv, nil
}

func (self *Discoverer) discoverDevice(ip, port string) (Device, error) {
	if "" == ip {
		return nil, errors.New("managed ip of local host is empty.")
	}

	communities := self.guessCommunities(ip)
	if nil == communities || 0 == len(communities) {
		return nil, errors.New("community ip of local host is empty.")
	}

	drv := Device{"address": ip, "communities": communities}
	return drv, self.initDevice(drv)
}

func (self *Discoverer) initDevice(drv Device) error {
	// read basic attributes
	for name, alias := range map[string]string{"sys.oid": "oid", "sys.descr": "description",
		"sys.type":       "catalog",
		"sys.services":   "services",
		"sys.name":       "name",
		"sys.location":   "location",
		"interfaceDescr": "interfaces",
		"ipAddress":      "addresses"} {
		self.logf(DEBUG, "read '%s' for '%v", name, drv["address"])
		metric, e := self.readMetric(drv, name)
		if nil != e {
			self.logf(ERROR, "read %s of '%v' failed, %v.", name, drv["address"], e)
		} else if nil == metric {
			self.logf(ERROR, "read %s of '%v' failed, result is nil.", name, drv["address"])
		} else {
			self.logf(DEBUG, "read '%s' for '%v' successed", name, drv["address"])
			drv[alias] = metric
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

func (self *Discoverer) addDevice(drv Device) {
	self.devices[drv["address"].(string)] = drv

	if ift := drv["interfaceTables"]; nil != ift {
		interfaceTables, ok := ift.([]interface{})
		if !ok {
			self.logf(FATAL, "interfaceTables is not []interface{}, actual is %T.", interfaceTables)
		} else if nil != interfaceTables {
			for _, ifs := range interfaceTables {
				row, ok := ifs.(map[string]interface{})
				if !ok {
					self.logf(FATAL, "interfaceTables is not map[string]interface{}, actual is %T.", ifs)
					break
				}
				self.ip2managed[row["address"].(string)] = drv["address"].(string)
			}
		}
	}

	ipAddresses, ok := drv["addresses"].(map[string]interface{})
	if !ok || nil == ipAddresses || 0 == len(ipAddresses) {
		return
	}

	for _, r := range ipAddresses {
		row, ok := r.(map[string]interface{})
		if !ok || nil == row {
			continue
		}
		self.ip2managed[row["address"].(string)] = drv["address"].(string)
	}
}

func (self *Discoverer) detectNewAddress(drv Device) {
	ip_list := map[string]int{}

	if ift := drv["interfaceTables"]; nil != ift {
		interfaceTables, ok := ift.([]interface{})
		if !ok {
			self.logf(FATAL, "interfaceTables is not []interface{}, actual is %T.", interfaceTables)
		} else if nil != interfaceTables {
			for _, ifs := range interfaceTables {
				row, ok := ifs.(map[string]interface{})
				if !ok {
					self.logf(FATAL, "interfaceTables is not map[string]interface{}, actual is %T.", ifs)
					continue
				}
				ip_list[row["address"].(string)] = 0
			}
		}
	}
	ipAddresses, ok := drv["addresses"].(map[string]interface{})
	if ok && nil != ipAddresses {
		for _, r := range ipAddresses {
			row, ok := r.(map[string]interface{})
			if !ok || nil == row {
				self.logf(FATAL, "ipAddress is not map[string]interface{}, actual is %T.", r)
				continue
			}
			ip_list[row["address"].(string)] = 0
		}
	}

	for ip, _ := range ip_list {
		if netutils.IsInvalidAddress(ip) {
			self.log(DEBUG, "skip invalid address - "+ip)
			continue
		}

		ip_range, err := netutils.ParseIPRange(ip + "/24")
		if nil != err {
			self.log(DEBUG, "scan ip range '"+ip+"/24' failed, "+err.Error())
			continue
		}

		self.log(DEBUG, "scan ip range '"+ip+"/24' success")

		communities := []string{"public"}
		if nil != self.params.Communities && 0 != len(self.params.Communities) {
			communities = self.params.Communities
		}

		for ip_range.HasNext() {
			addr := ip_range.Current().String()
			// err = self.icmp_pinger.Send(addr, nil)
			// if nil != err {
			//	self.log(DEBUG, "send icmp to '"+addr+"'' failed, "+err.Error())
			// }
			self.log(DEBUG, "send probe packet to '"+addr+"'")
			for i, community := range communities {
				if i != 0 {
					//select {
					//case <-time.After(500 * time.Millisecond): // it is required, otherwise target host should busy and discard it
					//}
					//time.Sleep(500 * time.Millisecond) // it is required, otherwise target host should busy and discard it
				}

				err = self.snmp_pinger.Send(addr+":161", snmp.SNMP_V2C, community)
				if nil != err {
					self.log(DEBUG, "send snmp to '"+addr+"' failed, "+err.Error())
				}
			}

			err = self.snmp_pinger.Send(addr+":161", snmp.SNMP_V3, "")
			if nil != err {
				self.log(DEBUG, "send snmp to '"+addr+"' failed, "+err.Error())
			}
		}
	}
}

func (self *Discoverer) isExists(ip string) bool {
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

	self.logf(INFO, "discover device '%v'", local["address"])
	self.addDevice(local)
	self.detectNewAddress(local)

	for d := 0; ; d++ {
		pending_drvs := make([]Device, 0, 10)
		goroutes := 0
		running := true
		for running {

			select {
			case reply := <-self.snmp_pinger.GetChannel():
				if nil != reply.Error {
					self.log(ERROR, "recv icmp failed - "+reply.Error.Error())
					running = false
					break
				}

				addr := reply.Addr.String()
				port := "161"
				idx := strings.IndexRune(addr, ':')
				if -1 != idx {
					port = addr[idx+1:]
					addr = addr[0:idx]
				}

				if netutils.IsInvalidAddress(addr) {
					self.log(DEBUG, "skip invalid address - "+addr)
					break
				}

				if !self.isExists(addr) {
					self.log(DEBUG, "skip old address - "+addr)
					break
				}

				go func() {
					drv, e := self.discoverDevice(addr, port)
					if nil != e {
						self.log(ERROR, "discover device '"+addr+"' failed, "+e.Error())
					} else {
						self.log(INFO, "discover device '"+addr+"'")
					}
					self.drv_ch <- drv
				}()
			case drv := <-self.drv_ch:
				if nil != drv {
					self.addDevice(drv)
					pending_drvs = append(pending_drvs, drv)
				}
			case <-time.After(1 * time.Minute):
				if 0 == goroutes {
					running = false
				}
			}
		}

		if 0 == len(pending_drvs) {
			self.log(INFO, "pending device is empty and exit.")
			break
		}
		if d > self.params.Depth {
			self.logf(INFO, "Reach the specified depth '%d' and exit.", self.params.Depth)
			break
		}
		go func() {
			for _, drv := range pending_drvs {
				self.detectNewAddress(drv)
			}
		}()
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
	//self.icmp_pinger.Close()
	self.snmp_pinger.Close()
}
