package sampling

// import (
// 	"bytes"
// 	"commons"
// 	"commons/types"
// 	ds "data_store"
// 	"errors"
// 	"fmt"
// 	"testing"
// )

// func TestBatchGetTableBasic(t *testing.T) {
// 	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
// 		_, err := client.DeleteBy("network_device", emptyParams)
// 		if nil != err {
// 			t.Error(err)
// 			return
// 		}

// 		res, e := batchGet(t, sampling_url+"/batch", []*ExchangeRequest{&ExchangeRequest{Address: "127.0.0.1", Action: "GET", Name: "sys", Params: map[string]string{"snmp.version": "v2c", "snmp.read_community": "public"}}})
// 		if nil != e {
// 			t.Error(e)
// 			return
// 		}
// 		for i := 0; i < 1000; i++ {
// 			res, e = batchGet(t, sampling_url+"/batch", []*ExchangeRequest{})
// 			if nil != e {
// 				t.Error(e)
// 				return
// 			}
// 			if nil != res && 0 != len(res) {
// 				break
// 			}
// 			time.Sleep(10 * time.Millisecond)
// 		}

// 		if nil == res || 0 == len(res) {
// 			t.Error("not result")
// 			return
// 		}

// 		if nil == res[0].Evalue {
// 			t.Error("values is nil")
// 		}

// 		t.Log(res[0])

// 		res, e = batchGet(t, sampling_url+"/batch", []*ExchangeRequest{})
// 		if nil != e {
// 			t.Error(e)
// 			return
// 		}
// 		if nil != res && 0 != len(res) {
// 			t.Error("repeated result")
// 		}
// 	})
// }
