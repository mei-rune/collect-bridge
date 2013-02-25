package mdb

// import (
// 	"strings"
// 	"testing"
// )

// func createMockAction(t *testing.T, ptype, id, factor string) string {
// 	return createJson(t, "history_rule", fmt.Sprintf(`{"name":"%s", "type":"redis_map_action", "parent_type":"%s", "parent_type":"devices", "parent_id":"%s"}`, factor, factor, factor, id))
// }

// func createMockHistoryRuleAndAction(t *testing.T, id, factor string) {
// 	hid := createMockHistoryRule(t, id, factor)
// 	createMockAction(t, "trigger", hid, stirngs.Itoa(atoi(factor)+10))
// 	createMockAction(t, "trigger", hid, stirngs.Itoa(atoi(factor)+20))
// 	createMockAction(t, "trigger", hid, stirngs.Itoa(atoi(factor)+30))
// 	createMockAction(t, "trigger", hid, stirngs.Itoa(atoi(factor)+40))
// }

// func createMockHistoryRuleAndAction2(t *testing.T, id, factor string) {
// 	hid := createMockHistoryRule2(t, id, factor)
// 	createMockAction(t, "trigger", hid, stirngs.Itoa(atoi(factor)+10))
// 	createMockAction(t, "trigger", hid, stirngs.Itoa(atoi(factor)+20))
// 	createMockAction(t, "trigger", hid, stirngs.Itoa(atoi(factor)+30))
// 	createMockAction(t, "trigger", hid, stirngs.Itoa(atoi(factor)+40))
// }

// func initPolymorphicData(t *testing.T) []string {

// 	id1 := createMockDevice(t, "1")
// 	id2 := createMockDevice(t, "2")
// 	id3 := createMockDevice(t, "3")
// 	id4 := createMockDevice(t, "4")

// 	createMockHistoryRuleAndAction2(t, "s")
// 	createMockHistoryRuleAndAction(t, id1, "10001")
// 	createMockHistoryRuleAndAction(t, id1, "10002")
// 	createMockHistoryRuleAndAction(t, id1, "10003")
// 	createMockHistoryRuleAndAction(t, id1, "10004")

// 	createMockHistoryRuleAndAction(t, id2, "20001")
// 	createMockHistoryRuleAndAction(t, id2, "20002")
// 	createMockHistoryRuleAndAction(t, id2, "20003")
// 	createMockHistoryRuleAndAction(t, id2, "20004")

// 	createMockHistoryRuleAndAction(t, id3, "30001")
// 	createMockHistoryRuleAndAction(t, id3, "30002")
// 	createMockHistoryRuleAndAction(t, id3, "30003")
// 	createMockHistoryRuleAndAction(t, id3, "30004")

// 	createMockHistoryRuleAndAction(t, id4, "40001")
// 	createMockHistoryRuleAndAction(t, id4, "40002")
// 	createMockHistoryRuleAndAction(t, id4, "40003")
// 	createMockHistoryRuleAndAction(t, id4, "40004")

// 	return []string{id1, id2, id3, id4}
// }

// func TestDeviceDeleteCascadeByAll(t *testing.T) {
// 	deleteById(t, "device", "all")
// 	deleteById(t, "trigger", "all")
// 	deleteById(t, "action", "all")

// 	idlist := initPolymorphicData(t)

// 	checkInterfaceCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 16, 4, 4, 4, 4)
// 	checkHistoryRuleCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 17, 4, 4, 4, 4)
// 	deleteById(t, "device", "all")
// 	checkInterfaceCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 0, 0, 0, 0, 0)
// 	checkHistoryRuleCount(t, idlist[0], idlist[1], idlist[2], idlist[3], 1, 0, 0, 0, 0)
// }

// func TestDeviceDeleteCascadeByQuery(t *testing.T) {

// 	deleteById(t, "device", "all")
// 	deleteById(t, "trigger", "all")
// 	deleteById(t, "action", "all")

// 	id1 := createMockDevice(t, "1")
// 	id2 := createMockDevice(t, "2")
// 	id3 := createMockDevice(t, "3")
// 	id4 := createMockDevice(t, "4")
// 	if "" == id1 {
// 		return
// 	}

// 	createMockHistoryRuleAndAction2(t, "s")
// 	createMockInterface(t, id1, "10001")
// 	createMockInterface(t, id1, "10002")
// 	createMockInterface(t, id1, "10003")
// 	createMockInterface(t, id1, "10004")
// 	createMockHistoryRuleAndAction(t, id1, "10001")
// 	createMockHistoryRuleAndAction(t, id1, "10002")
// 	createMockHistoryRuleAndAction(t, id1, "10003")
// 	createMockHistoryRuleAndAction(t, id1, "10004")

// 	createMockInterface(t, id2, "20001")
// 	createMockInterface(t, id2, "20002")
// 	createMockInterface(t, id2, "20003")
// 	createMockInterface(t, id2, "20004")
// 	createMockHistoryRuleAndAction(t, id2, "20001")
// 	createMockHistoryRuleAndAction(t, id2, "20002")
// 	createMockHistoryRuleAndAction(t, id2, "20003")
// 	createMockHistoryRuleAndAction(t, id2, "20004")

// 	createMockInterface(t, id3, "30001")
// 	createMockInterface(t, id3, "30002")
// 	createMockInterface(t, id3, "30003")
// 	createMockInterface(t, id3, "30004")
// 	createMockHistoryRuleAndAction(t, id3, "30001")
// 	createMockHistoryRuleAndAction(t, id3, "30002")
// 	createMockHistoryRuleAndAction(t, id3, "30003")
// 	createMockHistoryRuleAndAction(t, id3, "30004")

// 	createMockInterface(t, id4, "40001")
// 	createMockInterface(t, id4, "40002")
// 	createMockInterface(t, id4, "40003")
// 	createMockInterface(t, id4, "40004")
// 	createMockHistoryRuleAndAction(t, id4, "40001")
// 	createMockHistoryRuleAndAction(t, id4, "40002")
// 	createMockHistoryRuleAndAction(t, id4, "40003")
// 	createMockHistoryRuleAndAction(t, id4, "40004")

// 	checkInterfaceCount(t, id1, id2, id3, id4, 16, 4, 4, 4, 4)
// 	checkHistoryRuleCount(t, id1, id2, id3, id4, 17, 4, 4, 4, 4)
// 	deleteBy(t, "device", map[string]string{"catalog": "[gte]3"})
// 	checkInterfaceCount(t, id1, id2, id3, id4, 8, 4, 4, 0, 0)
// 	checkHistoryRuleCount(t, id1, id2, id3, id4, 9, 4, 4, 0, 0)
// }

// func TestDeviceDeleteCascadeById(t *testing.T) {

// 	deleteById(t, "device", "all")
// 	deleteById(t, "trigger", "all")
// 	deleteById(t, "action", "all")

// 	id1 := createMockDevice(t, "1")
// 	id2 := createMockDevice(t, "2")
// 	id3 := createMockDevice(t, "3")
// 	id4 := createMockDevice(t, "4")
// 	if "" == id1 {
// 		return
// 	}

// 	createMockHistoryRuleAndAction2(t, "s")
// 	createMockInterface(t, id1, "10001")
// 	createMockInterface(t, id1, "10002")
// 	createMockInterface(t, id1, "10003")
// 	createMockInterface(t, id1, "10004")
// 	createMockHistoryRuleAndAction(t, id1, "10001")
// 	createMockHistoryRuleAndAction(t, id1, "10002")
// 	createMockHistoryRuleAndAction(t, id1, "10003")
// 	createMockHistoryRuleAndAction(t, id1, "10004")

// 	createMockInterface(t, id2, "20001")
// 	createMockInterface(t, id2, "20002")
// 	createMockInterface(t, id2, "20003")
// 	createMockInterface(t, id2, "20004")
// 	createMockHistoryRuleAndAction(t, id2, "20001")
// 	createMockHistoryRuleAndAction(t, id2, "20002")
// 	createMockHistoryRuleAndAction(t, id2, "20003")
// 	createMockHistoryRuleAndAction(t, id2, "20004")

// 	createMockInterface(t, id3, "30001")
// 	createMockInterface(t, id3, "30002")
// 	createMockInterface(t, id3, "30003")
// 	createMockInterface(t, id3, "30004")
// 	createMockHistoryRuleAndAction(t, id3, "30001")
// 	createMockHistoryRuleAndAction(t, id3, "30002")
// 	createMockHistoryRuleAndAction(t, id3, "30003")
// 	createMockHistoryRuleAndAction(t, id3, "30004")

// 	createMockInterface(t, id4, "40001")
// 	createMockInterface(t, id4, "40002")
// 	createMockInterface(t, id4, "40003")
// 	createMockInterface(t, id4, "40004")
// 	createMockHistoryRuleAndAction(t, id4, "40001")
// 	createMockHistoryRuleAndAction(t, id4, "40002")
// 	createMockHistoryRuleAndAction(t, id4, "40003")
// 	createMockHistoryRuleAndAction(t, id4, "40004")

// 	checkInterfaceCount(t, id1, id2, id3, id4, 16, 4, 4, 4, 4)
// 	checkHistoryRuleCount(t, id1, id2, id3, id4, 17, 4, 4, 4, 4)

// 	deleteById(t, "device", id1)

// 	checkInterfaceCount(t, id1, id2, id3, id4, 12, 0, 4, 4, 4)
// 	checkHistoryRuleCount(t, id1, id2, id3, id4, 13, 0, 4, 4, 4)
// 	deleteById(t, "device", id2)

// 	checkInterfaceCount(t, id1, id2, id3, id4, 8, 0, 0, 4, 4)
// 	checkHistoryRuleCount(t, id1, id2, id3, id4, 9, 0, 0, 4, 4)
// 	deleteById(t, "device", id3)

// 	checkInterfaceCount(t, id1, id2, id3, id4, 4, 0, 0, 0, 4)
// 	checkHistoryRuleCount(t, id1, id2, id3, id4, 5, 0, 0, 0, 4)
// 	deleteById(t, "device", id4)

// 	checkInterfaceCount(t, id1, id2, id3, id4, 0, 0, 0, 0, 0)
// 	checkHistoryRuleCount(t, id1, id2, id3, id4, 1, 0, 0, 0, 0)
// }
