package metrics

import (
	"commons"
)

type ipAddress struct {
	dispatcherBase
}

func (self *ipAddress) Init(params map[string]interface{}, drvName string) commons.RuntimeError {
	e := self.dispatcherBase.Init(params, drvName)
	if nil != e {
		return e
	}
	self.get = func(params map[string]string) commons.Result {
		return self.GetDefault(params)
	}
	return nil
}

func (self *ipAddress) GetDefault(params map[string]string) commons.Result {
	return self.GetTable(params, "1.3.6.1.2.1.4.20.1", "1,2,3,4,5",
		func(table map[string]interface{}, key string, old_row map[string]interface{}) error {
			new_row := map[string]interface{}{}
			new_row["address"] = GetIPAddress(params, old_row, "1")
			new_row["ifIndex"] = GetInt32(params, old_row, "2", -1)
			new_row["netmask"] = GetIPAddress(params, old_row, "3")
			new_row["bcastAddress"] = GetInt32(params, old_row, "4", -1)
			new_row["reasmMaxSize"] = GetInt32(params, old_row, "5", -1)
			table[key] = new_row
			return nil
		})
}

type route struct {
	dispatcherBase
}

func (self *route) Init(params map[string]interface{}, drvName string) commons.RuntimeError {
	e := self.dispatcherBase.Init(params, drvName)
	if nil != e {
		return e
	}
	self.get = func(params map[string]string) commons.Result {
		return self.GetDefault(params)
	}
	return nil
}

func (self *route) GetDefault(params map[string]string) commons.Result {
	return self.GetTable(params, "1.3.6.1.2.1.4.21.1", "1,2,7,8,9,10,11",
		func(table map[string]interface{}, key string, old_row map[string]interface{}) error {
			new_row := map[string]interface{}{}
			new_row["dest"] = GetIPAddress(params, old_row, "1")
			new_row["ifIndex"] = GetInt32(params, old_row, "2", -1)
			new_row["next_hop"] = GetIPAddress(params, old_row, "7")
			new_row["type"] = GetInt32(params, old_row, "8", -1)
			new_row["proto"] = GetInt32(params, old_row, "9", -1)
			new_row["age"] = GetInt32(params, old_row, "10", -1)
			new_row["mask"] = GetIPAddress(params, old_row, "11")
			table[key] = new_row
			return nil
		})
}

type arp struct {
	dispatcherBase
}

func (self *arp) Init(params map[string]interface{}, drvName string) commons.RuntimeError {
	e := self.dispatcherBase.Init(params, drvName)
	if nil != e {
		return e
	}
	self.get = func(params map[string]string) commons.Result {
		return self.GetDefault(params)
	}
	return nil
}

func (self *arp) GetDefault(params map[string]string) commons.Result {
	return self.GetTable(params, "1.3.6.1.2.1.4.22.1", "1,2,3,4",
		func(table map[string]interface{}, key string, old_row map[string]interface{}) error {
			new_row := map[string]interface{}{}
			new_row["ifIndex"] = GetInt32(params, old_row, "1", -1)
			new_row["next_hop"] = GetHardwareAddress(params, old_row, "2")
			new_row["address"] = GetIPAddress(params, old_row, "3")
			new_row["type"] = GetInt32(params, old_row, "4", -1)
			table[key] = new_row
			return nil
		})
}

func init() {
	commons.METRIC_DRVS["ipAddress"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &ipAddress{}
		return drv, drv.Init(params, "snmp")
	}
	commons.METRIC_DRVS["route"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &route{}
		return drv, drv.Init(params, "snmp")
	}
	commons.METRIC_DRVS["arp"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &arp{}
		return drv, drv.Init(params, "snmp")
	}
}
