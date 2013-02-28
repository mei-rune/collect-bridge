package mdb

import (
	"fmt"
	"testing"
)

func createMockSnmpParams(t *testing.T, factor string) string {
	return createJson(t, "snmp_params", fmt.Sprintf(`{"address":"20.0.9.%s", "port":%s, "version":"snmp_v2c", "community":"aa"}`, factor, factor))
}
func createMockSnmpParams2(t *testing.T, factor string) string {
	return createJson(t, "access_params", fmt.Sprintf(`{"type":"snmp_params", "address":"20.0.9.%s", "port":%s, "version":"snmp_v2c", "community":"aa"}`, factor, factor))
}

func validMockSNMPWithFactor(t *testing.T, factor string, drvs []map[string]interface{}) {
	drv := searchBy(drvs, func(r map[string]interface{}) bool { return r["address"] == "20.0.9."+factor })
	if nil == drv {
		t.Errorf("find snmp_params with factor=" + factor + " failed")
	} else {
		validMockSNMP(t, factor, drv)
	}
}

func validMockSNMP(t *testing.T, factor string, drv map[string]interface{}) {
	defer func() {
		if e := recover(); nil != e {
			t.Errorf("snmp panic '%v', %v", e, drv)
			return
		}
	}()

	if "20.0.9."+factor != drv["address"].(string) {
		t.Errorf("excepted address is '20.0.9.%s', actual address is '%v'", factor, drv["address"])
		return
	}
	if atoi(factor) != fetchInt(drv, "port") {
		t.Errorf("excepted port is '%s', actual port is '%v'", factor, drv["port"])
		return
	}
	if "snmp_v2c" != drv["version"].(string) {
		t.Errorf("excepted version is 'snmp_v2c', actual version is '%v'", drv["version"])
		return
	}
	if "aa" != drv["community"].(string) {
		t.Errorf("excepted community is 'aa', actual community is '%v'", drv["community"])
		return
	}
}

func createMockSshParams(t *testing.T, factor string) string {
	return createJson(t, "ssh_params", fmt.Sprintf(`{"address":"20.0.8.%s", "port":23%s, "user":"a%s", "password":"cc%s"}`, factor, factor, factor, factor))
}
func createMockSshParams2(t *testing.T, factor string) string {
	return createJson(t, "access_params", fmt.Sprintf(`{"type":"ssh_params", "address":"20.0.8.%s", "port":23%s, "user":"a%s", "password":"cc%s"}`, factor, factor, factor, factor))
}

func validMockSshWithFactor(t *testing.T, factor string, drvs []map[string]interface{}) {
	drv := searchBy(drvs, func(r map[string]interface{}) bool { return r["address"] == "20.0.8."+factor })
	if nil == drv {
		t.Errorf("find ssh_params with factor=" + factor + " failed")
	} else {
		validMockSsh(t, factor, drv)
	}
}

func validMockSsh(t *testing.T, factor string, drv map[string]interface{}) {

	defer func() {
		if e := recover(); nil != e {
			t.Errorf("ssh panic '%v', %v", e, drv)
			return
		}
	}()

	if "20.0.8."+factor != drv["address"].(string) {
		t.Errorf("excepted address is '20.0.8.%s', actual address is '%v'", factor, drv["address"])
		return
	}
	if atoi("23"+factor) != fetchInt(drv, "port") {
		t.Errorf("excepted port is '23%s', actual port is '%v'", factor, drv["port"])
		return
	}
	if "a"+factor != drv["user"].(string) {
		t.Errorf("excepted user is 'a%s', actual user is '%v'", factor, drv["user"])
		return
	}
	if "cc"+factor != drv["password"].(string) {
		t.Errorf("excepted password is 'cc%s', actual password is '%v'", factor, drv["password"])
		return
	}
}

func createMockWbemParams(t *testing.T, factor string) string {
	return createJson(t, "wbem_params", fmt.Sprintf(`{"url":"tcp://20.0.8.%s", "user":"aa%s", "password":"cccc%s"}`, factor, factor, factor))
}
func createMockWbemParams2(t *testing.T, factor string) string {
	return createJson(t, "access_params", fmt.Sprintf(`{"type":"wbem_params", "url":"tcp://20.0.8.%s", "user":"aa%s", "password":"cccc%s"}`, factor, factor, factor))
}

func validMockWbemWithFactor(t *testing.T, factor string, drvs []map[string]interface{}) {
	drv := searchBy(drvs, func(r map[string]interface{}) bool { return r["url"] == "tcp://20.0.8."+factor })
	if nil == drv {
		t.Errorf("find wbem_params with factor=" + factor + " failed")
	} else {
		validMockWbem(t, factor, drv)
	}
}

func validMockWbem(t *testing.T, factor string, drv map[string]interface{}) {
	defer func() {
		if e := recover(); nil != e {
			t.Errorf("wbem panic '%v', %v", e, drv)
			return
		}
	}()

	if "tcp://20.0.8."+factor != drv["url"].(string) {
		t.Errorf("excepted url is 'tcp://20.0.8.%s', actual url is '%v'", factor, drv["url"])
		return
	}
	if "aa"+factor != drv["user"].(string) {
		t.Errorf("excepted user is 'aa%s', actual user is '%v'", factor, drv["user"])
		return
	}
	if "cccc"+factor != drv["password"].(string) {
		t.Errorf("excepted password is 'cc%s', actual password is '%v'", factor, drv["password"])
		return
	}
}

func checkHierarchyCount(t *testing.T, tName string, all int) {
	if c := count(t, tName, map[string]string{}); all != c {
		t.Errorf("%d != len(all.%s), actual is %d", all, tName, c)
	}
}

func checkHierarchyForFindById(t *testing.T, tName, tName2, id, factor string) {
	c := findById(t, tName, id)
	switch tName2 {
	case "snmp_params":
		validMockSNMP(t, factor, c)
	case "ssh_params":
		validMockSsh(t, factor, c)
	case "wbem_params":
		validMockWbem(t, factor, c)
	}
}
func checkHierarchyForFindBy(t *testing.T, tName string, all int) {
	c := findBy(t, tName, map[string]string{})
	if all != len(c) {
		t.Errorf("%d != len(all.%s), actual is %d", all, tName, c)
	}
	switch tName {
	case "snmp_params":
		validMockSNMPWithFactor(t, "1", c)
		validMockSNMPWithFactor(t, "2", c)
		validMockSNMPWithFactor(t, "3", c)
		validMockSNMPWithFactor(t, "4", c)
		validMockSNMPWithFactor(t, "5", c)
	case "ssh_params":
		validMockSshWithFactor(t, "1", c)
		validMockSshWithFactor(t, "2", c)
		validMockSshWithFactor(t, "3", c)
		validMockSshWithFactor(t, "4", c)
		validMockSshWithFactor(t, "5", c)

	case "wbem_params":
		validMockWbemWithFactor(t, "1", c)
		validMockWbemWithFactor(t, "2", c)
		validMockWbemWithFactor(t, "3", c)
		validMockWbemWithFactor(t, "4", c)
		validMockWbemWithFactor(t, "5", c)
	case "access_params":
		validMockSNMPWithFactor(t, "1", c)
		validMockSNMPWithFactor(t, "2", c)
		validMockSNMPWithFactor(t, "3", c)
		validMockSNMPWithFactor(t, "4", c)
		validMockSNMPWithFactor(t, "5", c)
		validMockSshWithFactor(t, "1", c)
		validMockSshWithFactor(t, "2", c)
		validMockSshWithFactor(t, "3", c)
		validMockSshWithFactor(t, "4", c)
		validMockSshWithFactor(t, "5", c)
		validMockWbemWithFactor(t, "1", c)
		validMockWbemWithFactor(t, "2", c)
		validMockWbemWithFactor(t, "3", c)
		validMockWbemWithFactor(t, "4", c)
		validMockWbemWithFactor(t, "5", c)
	case "endpoint_params":
		validMockSNMPWithFactor(t, "1", c)
		validMockSNMPWithFactor(t, "2", c)
		validMockSNMPWithFactor(t, "3", c)
		validMockSNMPWithFactor(t, "4", c)
		validMockSNMPWithFactor(t, "5", c)
		validMockSshWithFactor(t, "1", c)
		validMockSshWithFactor(t, "2", c)
		validMockSshWithFactor(t, "3", c)
		validMockSshWithFactor(t, "4", c)
		validMockSshWithFactor(t, "5", c)
	}
}

func initAccessParamsData(t *testing.T) []string {
	snmpid1 := createMockSnmpParams(t, "1")
	snmpid2 := createMockSnmpParams(t, "2")
	snmpid3 := createMockSnmpParams(t, "3")
	snmpid4 := createMockSnmpParams(t, "4")
	snmpid5 := createMockSnmpParams(t, "5")

	sshid1 := createMockSshParams(t, "1")
	sshid2 := createMockSshParams(t, "2")
	sshid3 := createMockSshParams(t, "3")
	sshid4 := createMockSshParams(t, "4")
	sshid5 := createMockSshParams(t, "5")

	wbemid1 := createMockWbemParams(t, "1")
	wbemid2 := createMockWbemParams(t, "2")
	wbemid3 := createMockWbemParams(t, "3")
	wbemid4 := createMockWbemParams(t, "4")
	wbemid5 := createMockWbemParams(t, "5")

	return []string{snmpid1, snmpid2, snmpid3, snmpid4, snmpid5,
		sshid1, sshid2, sshid3, sshid4, sshid5, wbemid1, wbemid2, wbemid3, wbemid4, wbemid5}

}

func initAccessParamsData2(t *testing.T) []string {
	snmpid1 := createMockSnmpParams2(t, "1")
	snmpid2 := createMockSnmpParams2(t, "2")
	snmpid3 := createMockSnmpParams2(t, "3")
	snmpid4 := createMockSnmpParams2(t, "4")
	snmpid5 := createMockSnmpParams2(t, "5")

	sshid1 := createMockSshParams2(t, "1")
	sshid2 := createMockSshParams2(t, "2")
	sshid3 := createMockSshParams2(t, "3")
	sshid4 := createMockSshParams2(t, "4")
	sshid5 := createMockSshParams2(t, "5")

	wbemid1 := createMockWbemParams2(t, "1")
	wbemid2 := createMockWbemParams2(t, "2")
	wbemid3 := createMockWbemParams2(t, "3")
	wbemid4 := createMockWbemParams2(t, "4")
	wbemid5 := createMockWbemParams2(t, "5")

	return []string{snmpid1, snmpid2, snmpid3, snmpid4, snmpid5,
		sshid1, sshid2, sshid3, sshid4, sshid5, wbemid1, wbemid2, wbemid3, wbemid4, wbemid5}

}
func TestHierarchyQuery(t *testing.T) {
	deleteById(t, "access_params", "all")
	idlist := initAccessParamsData(t)

	checkHierarchyCount(t, "access_params", 15)
	checkHierarchyCount(t, "endpoint_params", 10)
	checkHierarchyCount(t, "snmp_params", 5)
	checkHierarchyCount(t, "ssh_params", 5)
	checkHierarchyCount(t, "wbem_params", 5)

	checkHierarchyForFindBy(t, "snmp_params", 5)
	checkHierarchyForFindBy(t, "ssh_params", 5)
	checkHierarchyForFindBy(t, "wbem_params", 5)

	checkHierarchyForFindBy(t, "access_params", 15)
	checkHierarchyForFindBy(t, "endpoint_params", 10)

	checkHierarchyForFindById(t, "endpoint_params", "snmp_params", idlist[0], "1")
	checkHierarchyForFindById(t, "endpoint_params", "snmp_params", idlist[1], "2")
	checkHierarchyForFindById(t, "endpoint_params", "snmp_params", idlist[2], "3")
	checkHierarchyForFindById(t, "endpoint_params", "snmp_params", idlist[3], "4")
	checkHierarchyForFindById(t, "endpoint_params", "snmp_params", idlist[4], "5")

	checkHierarchyForFindById(t, "endpoint_params", "ssh_params", idlist[5], "1")
	checkHierarchyForFindById(t, "endpoint_params", "ssh_params", idlist[6], "2")
	checkHierarchyForFindById(t, "endpoint_params", "ssh_params", idlist[7], "3")
	checkHierarchyForFindById(t, "endpoint_params", "ssh_params", idlist[8], "4")
	checkHierarchyForFindById(t, "endpoint_params", "ssh_params", idlist[9], "5")
	checkHierarchyForFindById(t, "access_params", "wbem_params", idlist[10], "1")
	checkHierarchyForFindById(t, "access_params", "wbem_params", idlist[11], "2")
	checkHierarchyForFindById(t, "access_params", "wbem_params", idlist[12], "3")
	checkHierarchyForFindById(t, "access_params", "wbem_params", idlist[13], "4")
	checkHierarchyForFindById(t, "access_params", "wbem_params", idlist[14], "5")
}

func TestHierarchyCreate(t *testing.T) {
	deleteById(t, "access_params", "all")
	idlist := initAccessParamsData2(t)

	checkHierarchyForFindById(t, "endpoint_params", "snmp_params", idlist[0], "1")
	checkHierarchyForFindById(t, "endpoint_params", "snmp_params", idlist[1], "2")
	checkHierarchyForFindById(t, "endpoint_params", "snmp_params", idlist[2], "3")
	checkHierarchyForFindById(t, "endpoint_params", "snmp_params", idlist[3], "4")
	checkHierarchyForFindById(t, "endpoint_params", "snmp_params", idlist[4], "5")

	checkHierarchyForFindById(t, "endpoint_params", "ssh_params", idlist[5], "1")
	checkHierarchyForFindById(t, "endpoint_params", "ssh_params", idlist[6], "2")
	checkHierarchyForFindById(t, "endpoint_params", "ssh_params", idlist[7], "3")
	checkHierarchyForFindById(t, "endpoint_params", "ssh_params", idlist[8], "4")
	checkHierarchyForFindById(t, "endpoint_params", "ssh_params", idlist[9], "5")
	checkHierarchyForFindById(t, "access_params", "wbem_params", idlist[10], "1")
	checkHierarchyForFindById(t, "access_params", "wbem_params", idlist[11], "2")
	checkHierarchyForFindById(t, "access_params", "wbem_params", idlist[12], "3")
	checkHierarchyForFindById(t, "access_params", "wbem_params", idlist[13], "4")
	checkHierarchyForFindById(t, "access_params", "wbem_params", idlist[14], "5")
}
