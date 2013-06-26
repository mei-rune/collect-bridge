package metrics

// import (
// 	"commons"
// )

// type cisco_discovry_protocol struct {
// 	SnmpBase
// }

// func (self *cisco_discovry_protocol) Get(params commons.Map) commons.Result {
// 	return self.GetTable(params, "1.3.6.1.4.1.9.9.23.1.2.1.1", "4,6,7,12",
// 		func(table map[string]interface{}, key string, old_row map[string]interface{}) error {
// 			new_row := map[string]interface{}{}
// 			new_row["peer_address"] = GetIPAddress(params, old_row, "4")
// 			new_row["peer_ifIndex"] = GetString(params, old_row, "6")
// 			new_row["link_mode"] = GetInt32(params, old_row, "7", -1)
// 			new_row["local_ifIndex"] = GetInt32(params, old_row, "12", -1)
// 			table[key] = new_row
// 			return nil
// 		})
// }

// type huawei_discovry_protocol struct {
// 	SnmpBase
// }

// func (self *huawei_discovry_protocol) Get(params commons.Map) commons.Result {
// 	return self.GetTable(params, "1.3.6.1.4.1.2011.6.7.5.6.1", "1,2,3",
// 		func(table map[string]interface{}, key string, old_row map[string]interface{}) error {
// 			new_row := map[string]interface{}{}
// 			new_row["peer_address"] = GetIPAddress(params, old_row, "1")
// 			new_row["peer_ifIndex"] = GetInt32(params, old_row, "2", -1)
// 			new_row["local_ifIndex"] = GetInt32(params, old_row, "3", -1)
// 			table[key] = new_row
// 			return nil
// 		})
// }

// // func init() {
// // 	commons.METRIC_DRVS["cisco_discovry_protocol"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
// // 		drv := &cisco_discovry_protocol{}
// // 		return drv, drv.Init(params, "snmp")
// // 	}
// // 	commons.METRIC_DRVS["huawei_discovry_protocol"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
// // 		drv := &huawei_discovry_protocol{}
// // 		return drv, drv.Init(params, "snmp")
// // 	}
// // }
