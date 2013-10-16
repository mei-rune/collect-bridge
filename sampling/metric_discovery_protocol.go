package sampling

import (
	"commons"
)

type cisco_discovery_protocol struct {
	snmpBase
}

func (self *cisco_discovery_protocol) Call(params MContext) commons.Result {
	return self.GetAllResult(params, "1.3.6.1.4.1.9.9.23.1.2.1.1", "4,6,7,12",
		func(key string, old_row map[string]interface{}) (map[string]interface{}, error) {
			new_row := map[string]interface{}{}
			new_row["peer_address"] = GetIPAddressWithDefault(params, old_row, "4")
			new_row["peer_ifIndex"] = GetStringWithDefault(params, old_row, "6")
			new_row["link_mode"] = GetInt32WithDefault(params, old_row, "7", -1)
			new_row["local_ifIndex"] = GetInt32WithDefault(params, old_row, "12", -1)
			return new_row, nil
		})
}

type huawei_discovery_protocol struct {
	snmpBase
}

func (self *huawei_discovery_protocol) Call(params MContext) commons.Result {
	return self.GetAllResult(params, "1.3.6.1.4.1.2011.6.7.5.6.1", "1,2,3",
		func(key string, old_row map[string]interface{}) (map[string]interface{}, error) {
			new_row := map[string]interface{}{}
			new_row["peer_address"] = GetIPAddressWithDefault(params, old_row, "1")
			new_row["peer_ifIndex"] = GetInt32WithDefault(params, old_row, "2", -1)
			new_row["local_ifIndex"] = GetInt32WithDefault(params, old_row, "3", -1)
			return new_row, nil
		})
}

func init() {

	Methods["cisco_discovery_protocol"] = newRouteSpec("get", "cisco_discovery_protocol", "the discovery protocol of cisco", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &cisco_discovery_protocol{}
			return drv, drv.Init(params)
		})

	Methods["huawei_discovery_protocol"] = newRouteSpec("get", "huawei_discovery_protocol", "the discovery protocol of huawei", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &huawei_discovery_protocol{}
			return drv, drv.Init(params)
		})
}
