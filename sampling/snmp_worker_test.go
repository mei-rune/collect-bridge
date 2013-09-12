package sampling

import (
	"commons"
	"commons/types"
	ds "data_store"
	"github.com/runner-mei/snmpclient"
	"strings"
	"testing"
	"time"
)

func TestSnmpTestExpired1(t *testing.T) {
	drv := &snmpWorker{buckets: make(map[string]*bucketByAddress), c: make(chan *snmpBucket, 10)}
	defer drv.Close()

	now := time.Now().Unix()
	bucket, _ := drv.GetOrCreate("127.0.0.1:161", snmpclient.SNMP_V3, "p")
	bucket.updated_at = now - *snmp_test_expired - 1

	if 1 != len(drv.buckets) {
		t.Error("1 != len(drv.buckets), actual is", len(drv.buckets))
		return
	}

	drv.clearTimeout()

	if 0 != len(drv.buckets) {
		t.Error("0 != len(drv.buckets), actual is", len(drv.buckets))
		return
	}
}

func TestSnmpTestExpired2(t *testing.T) {
	drv := &snmpWorker{buckets: make(map[string]*bucketByAddress), c: make(chan *snmpBucket, 10)}
	defer drv.Close()

	now := time.Now().Unix()
	bucket, _ := drv.GetOrCreate("127.0.0.1:161", snmpclient.SNMP_V3, "p")
	bucket.updated_at = now - *snmp_test_expired - 1

	if 1 != len(drv.buckets) {
		t.Error("1 != len(drv.buckets), actual is", len(drv.buckets))
		return
	}

	drv.scan(0)

	if 0 != len(drv.buckets) {
		t.Error("0 != len(drv.buckets), actual is", len(drv.buckets))
		return
	}
}

func TestSnmpTestNative(t *testing.T) {
	is_icmp_test = false
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, e := client.DeleteBy("network_device", emptyParams)
		if nil != e {
			t.Error(e)
			return
		}
		res := nativeGet(t, sampling_url, "127.0.0.1", "snmp_test", map[string]string{"snmp.version": "v2c", "snmp.read_community": "public"})
		if res.HasError() && !strings.Contains(res.ErrorMessage(), "sampled is pending.") {
			t.Error(res.Error())
			return
		}
		time.Sleep(1 * time.Second)
		res = nativeGet(t, sampling_url, "127.0.0.1", "snmp_test", map[string]string{"snmp.version": "v2c", "snmp.read_community": "public"})
		if res.HasError() {
			t.Error(res.Error())
			return
		}
		if nil == res.InterfaceValue() {
			t.Error("values is nil")
		}

		t.Log(res.InterfaceValue())
		m, err := res.Value().AsObject()
		if nil != err {
			t.Error(err)
			return
		}
		if true != commons.GetBoolWithDefault(m, "result", false) {
			t.Error("values is error")
		}
	})
}

func TestSnmpTest(t *testing.T) {
	is_icmp_test = false
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, e := client.DeleteBy("network_device", emptyParams)
		if nil != e {
			t.Error(e)
			return
		}

		id := ds.CreateItForTest(t, client, "network_device", mo)
		ds.CreateItByParentForTest(t, client, "network_device", id, "wbem_param", wbem_params)
		ds.CreateItByParentForTest(t, client, "network_device", id, "snmp_param", snmp_params)

		res := urlGet(t, sampling_url, "network_device", id, "snmp_test")
		if res.HasError() && !strings.Contains(res.ErrorMessage(), "sampled is pending.") {
			t.Error(res.Error())
			return
		}
		time.Sleep(1 * time.Second)

		res = urlGet(t, sampling_url, "network_device", id, "snmp_test")
		if res.HasError() {
			t.Error(res.Error())
			return
		}

		if nil == res.InterfaceValue() {
			t.Error("values is nil")
		}

		t.Log(res.InterfaceValue())
		m, err := res.Value().AsObject()
		if nil != err {
			t.Error(err)
			return
		}
		if true != commons.GetBoolWithDefault(m, "result", false) {
			t.Error("values is error")
		}
	})
}

// func TestICMPNotRecv(t *testing.T) {
// 	is_icmp_test = true

// 	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
// 		_, e := client.DeleteBy("network_device", emptyParams)
// 		if nil != e {
// 			t.Error(e)
// 			return
// 		}

// 		id := ds.CreateItForTest(t, client, "network_device", mo)
// 		res := urlGet(t, sampling_url, "network_device", id, "icmp")
// 		if res.HasError() && !strings.Contains(res.ErrorMessage(), "sample is pending.") {
// 			t.Error(res.Error())
// 			return
// 		}
// 		time.Sleep(1 * time.Second)

// 		res = urlGet(t, sampling_url, "network_device", id, "icmp")
// 		if res.HasError() {
// 			t.Error(res.Error())
// 			return
// 		}

// 		if nil == res.InterfaceValue() {
// 			t.Error("values is nil")
// 		}

// 		t.Log(res.InterfaceValue())
// 		_, err := res.Value().AsArray()
// 		if nil != err {
// 			t.Error(err)
// 			return
// 		}
// 	})
// }
