package sampling

import (
	"commons"
)

type ipAddress struct {
	snmpBase
}

func (self *ipAddress) Call(params MContext) commons.Result {
	return self.GetAllResult(params, "1.3.6.1.2.1.4.20.1", "1,2,3,4,5",
		func(key string, old_row map[string]interface{}) (map[string]interface{}, error) {
			new_row := map[string]interface{}{}
			new_row["address"] = GetIPAddress(params, old_row, "1")
			new_row["ifIndex"] = GetInt32(params, old_row, "2", -1)
			new_row["netmask"] = GetIPAddress(params, old_row, "3")
			new_row["bcastAddress"] = GetInt32(params, old_row, "4", -1)
			new_row["reasmMaxSize"] = GetInt32(params, old_row, "5", -1)
			return new_row, nil
		})
}

type route struct {
	snmpBase
}

// ipRouteTable    	1.3.6.1.2.1.4.21
//   ipRouteEntry  	  1.3.6.1.2.1.4.21.1
//     ipRouteDest    	1.3.6.1.2.1.4.21.1.1
//     ipRouteIfIndex  	1.3.6.1.2.1.4.21.1.2
//     ipRouteMetric1  	1.3.6.1.2.1.4.21.1.3
//     ipRouteMetric2  	1.3.6.1.2.1.4.21.1.4
//     ipRouteMetric3  	1.3.6.1.2.1.4.21.1.5
//     ipRouteMetric4  	1.3.6.1.2.1.4.21.1.6
//     ipRouteNextHop  	1.3.6.1.2.1.4.21.1.7
//     ipRouteType     	1.3.6.1.2.1.4.21.1.8
//     ipRouteProto    	1.3.6.1.2.1.4.21.1.9
//     ipRouteAge  	    1.3.6.1.2.1.4.21.1.10
//     ipRouteMask  	  1.3.6.1.2.1.4.21.1.11
//     ipRouteMetric5  	1.3.6.1.2.1.4.21.1.12
//     ipRouteInfo  	  1.3.6.1.2.1.4.21.1.13

func (self *route) Call(params MContext) commons.Result {
	return self.GetAllResult(params, "1.3.6.1.2.1.4.21.1", "1,2,3,4,5,6,7,8,9,10,11,12,13",
		func(key string, old_row map[string]interface{}) (map[string]interface{}, error) {
			new_row := map[string]interface{}{}
			new_row["dest"] = GetIPAddress(params, old_row, "1")
			new_row["ifIndex"] = GetInt32(params, old_row, "2", -1)
			new_row["metric1"] = GetInt32(params, old_row, "3", -1)
			new_row["metric2"] = GetInt32(params, old_row, "4", -1)
			new_row["metric3"] = GetInt32(params, old_row, "5", -1)
			new_row["metric4"] = GetInt32(params, old_row, "6", -1)
			new_row["next_hop"] = GetIPAddress(params, old_row, "7")
			new_row["type"] = GetInt32(params, old_row, "8", -1)
			new_row["proto"] = GetInt32(params, old_row, "9", -1)
			new_row["age"] = GetInt32(params, old_row, "10", -1)
			new_row["mask"] = GetIPAddress(params, old_row, "11")
			new_row["metric5"] = GetInt32(params, old_row, "12", -1)
			new_row["info"] = GetOid(params, old_row, "13")
			return new_row, nil
		})
}

type arp struct {
	snmpBase
}

func (self *arp) Call(params MContext) commons.Result {
	return self.GetAllResult(params, "1.3.6.1.2.1.4.22.1", "1,2,3,4",
		func(key string, old_row map[string]interface{}) (map[string]interface{}, error) {
			new_row := map[string]interface{}{}
			new_row["ifIndex"] = GetInt32(params, old_row, "1", -1)
			new_row["next_hop"] = GetHardwareAddress(params, old_row, "2")
			new_row["address"] = GetIPAddress(params, old_row, "3")
			new_row["type"] = GetInt32(params, old_row, "4", -1)
			return new_row, nil
		})
}

func init() {

	Methods["ip_address"] = newRouteSpec("get", "ipAddress", "the address table", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &ipAddress{}
			return drv, drv.Init(params)
		})

	Methods["route"] = newRouteSpec("get", "route", "the route table", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &route{}
			return drv, drv.Init(params)
		})

	Methods["arp"] = newRouteSpec("get", "arp", "the arp table", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &arp{}
			return drv, drv.Init(params)
		})
}
