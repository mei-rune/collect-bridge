package sampling

import (
	"commons"
	"commons/types"
	ds "data_store"
	"strings"
	"testing"
	"time"
)

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
