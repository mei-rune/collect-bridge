package ds

import (
	"commons/types"
	"fmt"
	"testing"
)

func createMockSnmpParams(t *testing.T, client *Client, factor string) string {
	return createJson(t, client, "snmp_param", fmt.Sprintf(`{ "port":%s, "version":"snmp_v2c", "read_community":"aa"}`, factor))
}
func createMockSnmpParams2(t *testing.T, client *Client, factor string) string {
	return createJson(t, client, "access_param", fmt.Sprintf(`{"type":"snmp_param", "port":%s, "version":"snmp_v2c", "read_community":"aa"}`, factor))
}

func validMockSNMPWithFactor(t *testing.T, client *Client, factor string, drvs []map[string]interface{}) {
	t.Logf("find snmp with " + factor)
	drv := searchBy(drvs, func(r map[string]interface{}) bool { return fmt.Sprint(r["port"]) == factor })
	if nil == drv {
		t.Errorf("find snmp_params with factor=" + factor + " failed")
	} else {
		validMockSNMP(t, client, factor, drv)
	}
}

func validMockSNMP(t *testing.T, client *Client, factor string, drv map[string]interface{}) {
	defer func() {
		if e := recover(); nil != e {
			t.Errorf("snmp panic '%v', %v", e, drv)
			return
		}
	}()

	if atoi(factor) != fetchInt(drv, "port") {
		t.Errorf("excepted port is '%s', actual port is '%v'", factor, drv["port"])
		return
	}
	if "snmp_v2c" != drv["version"].(string) {
		t.Errorf("excepted version is 'snmp_v2c', actual version is '%v'", drv["version"])
		return
	}
	if "aa" != drv["read_community"].(string) {
		t.Errorf("excepted read_community is 'aa', actual read_community is '%v'", drv["read_community"])
		return
	}
}

func createMockSshParams(t *testing.T, client *Client, factor string) string {
	return createJson(t, client, "ssh_param", fmt.Sprintf(`{ "port":23%s, "user_name":"a%s", "user_password":"cc%s"}`, factor, factor, factor))
}
func createMockSshParams2(t *testing.T, client *Client, factor string) string {
	return createJson(t, client, "access_param", fmt.Sprintf(`{"type":"ssh_param", "port":23%s, "user_name":"a%s", "user_password":"cc%s"}`, factor, factor, factor))
}

func validMockSshWithFactor(t *testing.T, client *Client, factor string, drvs []map[string]interface{}) {
	t.Logf("find ssh with " + factor)
	drv := searchBy(drvs, func(r map[string]interface{}) bool { return fmt.Sprint(r["port"]) == "23"+factor })
	if nil == drv {
		t.Errorf("find ssh_params with factor=" + factor + " failed")
	} else {
		validMockSsh(t, client, factor, drv)
	}
}

func validMockSsh(t *testing.T, client *Client, factor string, drv map[string]interface{}) {
	defer func() {
		if e := recover(); nil != e {
			t.Errorf("ssh panic '%v', %v", e, drv)
			return
		}
	}()

	if atoi("23"+factor) != fetchInt(drv, "port") {
		t.Errorf("excepted port is '23%s', actual port is '%v'", factor, drv["port"])
		return
	}
	if "a"+factor != drv["user_name"].(string) {
		t.Errorf("excepted user is 'a%s', actual user is '%v'", factor, drv["user_name"])
		return
	}
	if "cc"+factor != drv["user_password"].(string) {
		t.Errorf("excepted password is 'cc%s', actual password is '%v'", factor, drv["user_password"])
		return
	}
}

func createMockWbemParams(t *testing.T, client *Client, factor string) string {
	return createJson(t, client, "wbem_param", fmt.Sprintf(`{"url":"tcp://20.0.8.%s", "user_name":"aa%s", "user_password":"cccc%s"}`, factor, factor, factor))
}
func createMockWbemParams2(t *testing.T, client *Client, factor string) string {
	return createJson(t, client, "access_param", fmt.Sprintf(`{"type":"wbem_param", "url":"tcp://20.0.8.%s", "user_name":"aa%s", "user_password":"cccc%s"}`, factor, factor, factor))
}

func validMockWbemWithFactor(t *testing.T, client *Client, factor string, drvs []map[string]interface{}) {
	drv := searchBy(drvs, func(r map[string]interface{}) bool { return r["url"] == "tcp://20.0.8."+factor })
	if nil == drv {
		t.Errorf("find wbem_params with factor=" + factor + " failed")
	} else {
		validMockWbem(t, client, factor, drv)
	}
}

func validMockWbem(t *testing.T, client *Client, factor string, drv map[string]interface{}) {
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
	if "aa"+factor != drv["user_name"].(string) {
		t.Errorf("excepted user is 'aa%s', actual user is '%v'", factor, drv["user_name"])
		return
	}
	if "cccc"+factor != drv["user_password"].(string) {
		t.Errorf("excepted password is 'cc%s', actual password is '%v'", factor, drv["user_password"])
		return
	}
}

func checkInheritsForCount(t *testing.T, client *Client, tName string, all int64) {
	if c := count(t, client, tName, map[string]string{}); all != c {
		t.Errorf("%d != len(all.%s), actual is %d", all, tName, c)
	}
}

func checkInheritsForFindById(t *testing.T, client *Client, tName, tName2, id, factor string) {
	c := findById(t, client, tName, id)
	switch tName2 {
	case "snmp_param":
		validMockSNMP(t, client, factor, c)
	case "ssh_param":
		validMockSsh(t, client, factor, c)
	case "wbem_param":
		validMockWbem(t, client, factor, c)
	}
}
func checkInheritsForFindBy(t *testing.T, client *Client, tName string, all int) {
	c := findBy(t, client, tName, map[string]string{})
	if all != len(c) {
		t.Errorf("%d != len(all.%s), actual is %d", all, tName, c)
	}

	t.Logf("%#v", c)
	//t.Logf("check query result by '%v'", tName)
	switch tName {
	case "snmp_param":
		validMockSNMPWithFactor(t, client, "1", c)
		validMockSNMPWithFactor(t, client, "2", c)
		validMockSNMPWithFactor(t, client, "3", c)
		validMockSNMPWithFactor(t, client, "4", c)
		validMockSNMPWithFactor(t, client, "5", c)
	case "ssh_param":
		validMockSshWithFactor(t, client, "1", c)
		validMockSshWithFactor(t, client, "2", c)
		validMockSshWithFactor(t, client, "3", c)
		validMockSshWithFactor(t, client, "4", c)
		validMockSshWithFactor(t, client, "5", c)

	case "wbem_param":
		validMockWbemWithFactor(t, client, "1", c)
		validMockWbemWithFactor(t, client, "2", c)
		validMockWbemWithFactor(t, client, "3", c)
		validMockWbemWithFactor(t, client, "4", c)
		validMockWbemWithFactor(t, client, "5", c)
	case "access_param":
		validMockSNMPWithFactor(t, client, "1", c)
		validMockSNMPWithFactor(t, client, "2", c)
		validMockSNMPWithFactor(t, client, "3", c)
		validMockSNMPWithFactor(t, client, "4", c)
		validMockSNMPWithFactor(t, client, "5", c)
		validMockSshWithFactor(t, client, "1", c)
		validMockSshWithFactor(t, client, "2", c)
		validMockSshWithFactor(t, client, "3", c)
		validMockSshWithFactor(t, client, "4", c)
		validMockSshWithFactor(t, client, "5", c)
		validMockWbemWithFactor(t, client, "1", c)
		validMockWbemWithFactor(t, client, "2", c)
		validMockWbemWithFactor(t, client, "3", c)
		validMockWbemWithFactor(t, client, "4", c)
		validMockWbemWithFactor(t, client, "5", c)
	case "endpoint_param":
		validMockSNMPWithFactor(t, client, "1", c)
		validMockSNMPWithFactor(t, client, "2", c)
		validMockSNMPWithFactor(t, client, "3", c)
		validMockSNMPWithFactor(t, client, "4", c)
		validMockSNMPWithFactor(t, client, "5", c)
		validMockSshWithFactor(t, client, "1", c)
		validMockSshWithFactor(t, client, "2", c)
		validMockSshWithFactor(t, client, "3", c)
		validMockSshWithFactor(t, client, "4", c)
		validMockSshWithFactor(t, client, "5", c)
	}
}

func initAccessParamsData(t *testing.T, client *Client) []string {
	snmpid1 := createMockSnmpParams(t, client, "1")
	snmpid2 := createMockSnmpParams(t, client, "2")
	snmpid3 := createMockSnmpParams(t, client, "3")
	snmpid4 := createMockSnmpParams(t, client, "4")
	snmpid5 := createMockSnmpParams(t, client, "5")

	sshid1 := createMockSshParams(t, client, "1")
	sshid2 := createMockSshParams(t, client, "2")
	sshid3 := createMockSshParams(t, client, "3")
	sshid4 := createMockSshParams(t, client, "4")
	sshid5 := createMockSshParams(t, client, "5")

	wbemid1 := createMockWbemParams(t, client, "1")
	wbemid2 := createMockWbemParams(t, client, "2")
	wbemid3 := createMockWbemParams(t, client, "3")
	wbemid4 := createMockWbemParams(t, client, "4")
	wbemid5 := createMockWbemParams(t, client, "5")

	return []string{snmpid1, snmpid2, snmpid3, snmpid4, snmpid5,
		sshid1, sshid2, sshid3, sshid4, sshid5, wbemid1, wbemid2, wbemid3, wbemid4, wbemid5}

}

func initAccessParamsData2(t *testing.T, client *Client) []string {
	snmpid1 := createMockSnmpParams2(t, client, "1")
	snmpid2 := createMockSnmpParams2(t, client, "2")
	snmpid3 := createMockSnmpParams2(t, client, "3")
	snmpid4 := createMockSnmpParams2(t, client, "4")
	snmpid5 := createMockSnmpParams2(t, client, "5")

	sshid1 := createMockSshParams2(t, client, "1")
	sshid2 := createMockSshParams2(t, client, "2")
	sshid3 := createMockSshParams2(t, client, "3")
	sshid4 := createMockSshParams2(t, client, "4")
	sshid5 := createMockSshParams2(t, client, "5")

	wbemid1 := createMockWbemParams2(t, client, "1")
	wbemid2 := createMockWbemParams2(t, client, "2")
	wbemid3 := createMockWbemParams2(t, client, "3")
	wbemid4 := createMockWbemParams2(t, client, "4")
	wbemid5 := createMockWbemParams2(t, client, "5")

	return []string{snmpid1, snmpid2, snmpid3, snmpid4, snmpid5,
		sshid1, sshid2, sshid3, sshid4, sshid5, wbemid1, wbemid2, wbemid3, wbemid4, wbemid5}
}
func TestInheritsForFindBy(t *testing.T) {
	SrvTest(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "access_param", map[string]string{})
		idlist := initAccessParamsData(t, client)

		checkInheritsForCount(t, client, "access_param", 15)
		checkInheritsForCount(t, client, "endpoint_param", 10)
		checkInheritsForCount(t, client, "snmp_param", 5)
		checkInheritsForCount(t, client, "ssh_param", 5)
		checkInheritsForCount(t, client, "wbem_param", 5)

		t.Log("check snmp params")
		checkInheritsForFindBy(t, client, "snmp_param", 5)
		t.Log("check ssh params")
		checkInheritsForFindBy(t, client, "ssh_param", 5)
		t.Log("check wbem params")
		checkInheritsForFindBy(t, client, "wbem_param", 5)

		checkInheritsForFindBy(t, client, "access_param", 15)
		checkInheritsForFindBy(t, client, "endpoint_param", 10)

		t.Log("=======")

		checkInheritsForFindById(t, client, "endpoint_param", "snmp_param", idlist[0], "1")
		checkInheritsForFindById(t, client, "endpoint_param", "snmp_param", idlist[1], "2")
		checkInheritsForFindById(t, client, "endpoint_param", "snmp_param", idlist[2], "3")
		checkInheritsForFindById(t, client, "endpoint_param", "snmp_param", idlist[3], "4")
		checkInheritsForFindById(t, client, "endpoint_param", "snmp_param", idlist[4], "5")

		checkInheritsForFindById(t, client, "endpoint_param", "ssh_param", idlist[5], "1")
		checkInheritsForFindById(t, client, "endpoint_param", "ssh_param", idlist[6], "2")
		checkInheritsForFindById(t, client, "endpoint_param", "ssh_param", idlist[7], "3")
		checkInheritsForFindById(t, client, "endpoint_param", "ssh_param", idlist[8], "4")
		checkInheritsForFindById(t, client, "endpoint_param", "ssh_param", idlist[9], "5")

		checkInheritsForFindById(t, client, "access_param", "wbem_param", idlist[10], "1")
		checkInheritsForFindById(t, client, "access_param", "wbem_param", idlist[11], "2")
		checkInheritsForFindById(t, client, "access_param", "wbem_param", idlist[12], "3")
		checkInheritsForFindById(t, client, "access_param", "wbem_param", idlist[13], "4")
		checkInheritsForFindById(t, client, "access_param", "wbem_param", idlist[14], "5")
	})
}

func TestInheritsForFindById(t *testing.T) {
	SrvTest(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "access_param", map[string]string{})
		idlist := initAccessParamsData2(t, client)

		checkInheritsForFindById(t, client, "endpoint_param", "snmp_param", idlist[0], "1")
		checkInheritsForFindById(t, client, "endpoint_param", "snmp_param", idlist[1], "2")
		checkInheritsForFindById(t, client, "endpoint_param", "snmp_param", idlist[2], "3")
		checkInheritsForFindById(t, client, "endpoint_param", "snmp_param", idlist[3], "4")
		checkInheritsForFindById(t, client, "endpoint_param", "snmp_param", idlist[4], "5")

		checkInheritsForFindById(t, client, "endpoint_param", "ssh_param", idlist[5], "1")
		checkInheritsForFindById(t, client, "endpoint_param", "ssh_param", idlist[6], "2")
		checkInheritsForFindById(t, client, "endpoint_param", "ssh_param", idlist[7], "3")
		checkInheritsForFindById(t, client, "endpoint_param", "ssh_param", idlist[8], "4")
		checkInheritsForFindById(t, client, "endpoint_param", "ssh_param", idlist[9], "5")

		checkInheritsForFindById(t, client, "access_param", "wbem_param", idlist[10], "1")
		checkInheritsForFindById(t, client, "access_param", "wbem_param", idlist[11], "2")
		checkInheritsForFindById(t, client, "access_param", "wbem_param", idlist[12], "3")
		checkInheritsForFindById(t, client, "access_param", "wbem_param", idlist[13], "4")
		checkInheritsForFindById(t, client, "access_param", "wbem_param", idlist[14], "5")
	})
}
