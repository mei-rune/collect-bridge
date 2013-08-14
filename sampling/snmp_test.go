package sampling

import (
	"commons"
	"commons/types"
	ds "data_store"
	"github.com/runner-mei/snmpclient"
	"strings"
	"testing"
)

func TestSnmpRead(t *testing.T) {
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		for idx, test := range []struct {
			action        string
			oid           string
			columns       string
			error_code    int
			error_message string
		}{{action: "snmp_get", oid: "", error_code: commons.BadRequestCode, error_message: "'snmp.oid' is empty."},
			{action: "snmp_get", oid: "1.3.6.1.2.1.1.5.0", error_code: 0, error_message: ""},
			{action: "snmp_next", oid: "1.3.6.1.2.1.1.4.0", error_code: 0, error_message: ""},
			//{action: "snmp_bulk", oid: "1.3.6.1.2.1.1.4.0,1.3.6.1.2.1.1.5.0", error_code: 0, error_message: ""},
			{action: "snmp_table", oid: "1.3.6.1.2.1.2.2.1", error_code: 0, error_message: ""},
			{action: "snmp_table", oid: "1.3.6.1.2.1.2.2.1", columns: "1,2,3,4,5,6", error_code: 0, error_message: ""}} {

			params := map[string]string{"snmp.oid": test.oid,
				"snmp.version":        "v2c",
				"snmp.read_community": "public"}

			if "snmp_table" == test.action && 0 != len(test.columns) {
				params["snmp.columns"] = test.columns
			}
			res := nativeGet(t, sampling_url, "127.0.0.1", test.action, params)

			if res.HasError() {
				if 0 != test.error_code {
					if test.error_code != res.ErrorCode() {
						t.Errorf("test[%v]: excepted error_code is '%v', actual error_code is '%v'", idx, test.error_code, res.ErrorCode())
					}

					if !strings.Contains(res.ErrorMessage(), test.error_message) {
						t.Errorf("test[%v]: excepted error_message contains '%v', actual error_message is '%v'", idx, test.error_message, res.ErrorMessage())
					}
					continue
				}
				t.Errorf("test[%v]: %v", idx, res.Error())
				continue
			}

			if nil == res.InterfaceValue() {
				t.Errorf("test[%v]: values is nil", idx)
				continue
			}

			var s string
			switch test.action {
			case "snmp_table":
				m, e := res.Value().AsObjects()
				if nil != e {
					t.Errorf("test[%v]: excepted is a []map, actual is '%v'", idx, res.Value())
					continue
				}
				if 0 == len(m) {
					t.Error("test[%v] result is empty", idx)
				}

				if 0 != len(test.columns) {
					if 6 != len(m[0]) {
						t.Errorf("test[%v] columns of the result is error, actual is %v", idx, len(m[0]))
					}
				} else if 22 != len(m[0]) {
					t.Errorf("test[%v] columns of the result is error, actual is %v", idx, len(m[0]))
				}
				continue
			case "snmp_bulk":
				m, e := res.Value().AsObject()
				if nil != e {
					t.Errorf("test[%v]: excepted is a map, actual is '%v'", idx, res.Value())
					continue
				}
				oid := "1.3.6.1.2.1.1.4.0"
				s, e = commons.GetString(m, oid)
				if nil != e {
					t.Errorf("test[%v]: excepted contains '%v', actual is '%v'", idx, oid, res.Value())
					continue
				}
				//t.Logf("test[%v]: result is '%v'", idx, res.Value())
				fallthrough
			case "snmp_next":

				m, e := res.Value().AsObject()
				if nil != e {
					t.Errorf("test[%v]: excepted is a map, actual is '%v'", idx, res.Value())
					continue
				}
				oid := "1.3.6.1.2.1.1.5.0"
				s, e = commons.GetString(m, oid)
				if nil != e {
					t.Errorf("test[%v]: excepted contains '%v', actual is '%v'", idx, oid, res.Value())
					continue
				}
			default:

				m, e := res.Value().AsObject()
				if nil != e {
					t.Errorf("test[%v]: excepted is a map, actual is '%v'", idx, res.Value())
					continue
				}
				s, e = commons.GetString(m, test.oid)
				if nil != e {
					t.Errorf("test[%v]: excepted is '%v', actual is '%v'", idx, test.oid, res.Value())
					continue
				}
			}

			if !strings.HasPrefix(s, "[octets]") {
				t.Errorf("test[%v]: excepted is '[octets]', actual is '%v'", idx, res.Value())
			}
			t.Logf("test[%v]: result is '%v'", idx, res.Value())
		}
	})
}

func TestSnmpWrite(t *testing.T) {
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		sss := snmpclient.SnmpOctetString("cccccc")
		for idx, test := range []struct {
			action        string
			oid           string
			body          interface{}
			error_code    int
			error_message string
		}{{action: "snmp_set", oid: "", body: "aaa", error_code: commons.BadRequestCode, error_message: "'snmp.oid' is empty."},
			{action: "snmp_set", oid: "1.3.6.1.2.1.1.6.0", body: sss.String(), error_code: 0, error_message: ""}} {

			read_params := map[string]string{"snmp.version": "v2c",
				"snmp.read_community": "public"}
			old_res1 := nativeGet(t, sampling_url, "127.0.0.1", "sys.location", read_params)
			if old_res1.HasError() {
				t.Errorf("test[%v]: %v", idx, old_res1.Error())
				continue
			}

			params := map[string]string{"snmp.oid": test.oid,
				"snmp.version":         "v2c",
				"snmp.write_community": "private",
				"timeout":              "1s"}
			res := nativePut(t, sampling_url, "127.0.0.1", test.action, params, test.body)
			if res.HasError() {
				if 0 != test.error_code {
					if test.error_code != res.ErrorCode() {
						t.Errorf("test[%v]: excepted error_code is '%v', actual error_code is '%v'", idx, test.error_code, res.ErrorCode())
					}

					if !strings.Contains(res.ErrorMessage(), test.error_message) {
						t.Errorf("test[%v]: excepted error_message contains '%v', actual error_message is '%v'", idx, test.error_message, res.ErrorMessage())
					}
					continue
				}
				t.Errorf("test[%v]: %v", idx, res.Error())
				continue
			}

			old_res2 := nativeGet(t, sampling_url, "127.0.0.1", "sys.location", read_params)
			if old_res2.HasError() {
				t.Errorf("test[%v]: %v", idx, old_res2.Error())
				continue
			}
			s1, e := old_res1.Value().AsString()
			if nil != e {
				t.Errorf("test[%v]: convert old_res1 failed, %v -- %T", idx, e.Error(), old_res1.InterfaceValue())
				continue
			}
			s2, e := old_res2.Value().AsString()
			if nil != e {
				t.Errorf("test[%v]: convert old_res2 failed, %v -- %T", idx, e.Error(), old_res1.InterfaceValue())
				continue
			}

			if s1 == s2 {
				t.Errorf("test[%v]: edit failed, s1 is %v, s2 is %v", idx, s1, s2)
				continue
			}
			t.Log(s1)

			old_value := snmpclient.SnmpOctetString(s1)
			res = nativePut(t, sampling_url, "127.0.0.1", test.action, params, old_value.String())
			if res.HasError() {
				t.Errorf("test[%v]: %v", idx, res.Error())
				continue
			}
		}
	})
}
