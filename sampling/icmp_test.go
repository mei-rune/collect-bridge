package sampling

import (
	"commons"
	"commons/types"
	ds "data_store"
	"strings"
	"testing"
	"time"
)

func TestICMPExpired1(t *testing.T) {
	drv := &icmpWorker{icmpBuffers: make(map[string]*icmpBucket), c: make(chan string, 10)}
	defer drv.Close()

	now := time.Now().Unix()
	bucket, _ := drv.GetOrCreate("127.0.0.1")
	bucket.updated_at = now - *snmp_test_expired - 1

	if 1 != len(drv.icmpBuffers) {
		t.Error("1 != len(drv.icmpBuffers), actual is", len(drv.icmpBuffers))
		return
	}

	drv.clearTimeout()

	if 0 != len(drv.icmpBuffers) {
		t.Error("0 != len(drv.icmpBuffers), actual is", len(drv.icmpBuffers))
		return
	}
}

func TestICMPExpired2(t *testing.T) {
	drv := &icmpWorker{icmpBuffers: make(map[string]*icmpBucket), c: make(chan string, 10)}
	defer drv.Close()

	now := time.Now().Unix()
	bucket, _ := drv.GetOrCreate("127.0.0.1")
	bucket.updated_at = now - *snmp_test_expired - 1

	if 1 != len(drv.icmpBuffers) {
		t.Error("1 != len(drv.icmpBuffers), actual is", len(drv.icmpBuffers))
		return
	}

	drv.scan(0)

	if 0 != len(drv.icmpBuffers) {
		t.Error("0 != len(drv.icmpBuffers), actual is", len(drv.icmpBuffers))
		return
	}
}

func TestICMPNative(t *testing.T) {
	is_icmp_test = false
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, e := client.DeleteBy("network_device", emptyParams)
		if nil != e {
			t.Error(e)
			return
		}
		res := nativeGet(t, sampling_url, "127.0.0.1", "icmp", nil)
		if res.HasError() && !strings.Contains(res.ErrorMessage(), "sampled is pending.") {
			t.Error(res.Error())
			return
		}
		time.Sleep(1 * time.Second)
		res = nativeGet(t, sampling_url, "127.0.0.1", "icmp", nil)
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

func TestICMP(t *testing.T) {
	is_icmp_test = false
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, e := client.DeleteBy("network_device", emptyParams)
		if nil != e {
			t.Error(e)
			return
		}

		id := ds.CreateItForTest(t, client, "network_device", mo)
		res := urlGet(t, sampling_url, "network_device", id, "icmp")
		if res.HasError() && !strings.Contains(res.ErrorMessage(), "sampled is pending.") {
			t.Error(res.Error())
			return
		}
		time.Sleep(1 * time.Second)

		res = urlGet(t, sampling_url, "network_device", id, "icmp")
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
