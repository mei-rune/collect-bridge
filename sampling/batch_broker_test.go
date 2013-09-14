package sampling

import (
	"bytes"
	"commons/types"
	ds "data_store"
	"encoding/json"
	"testing"
	"time"
)

func TestJson(t *testing.T) {
	for _, test := range []struct {
		id uint64
		o  interface{}
		s  string
		e  error
	}{{id: 0, o: &ExchangeRequest{}, s: `{"channel":"","request_id":0,"action":"","metric-name":""}`},
		{id: 13, o: &ExchangeRequest{ChannelName: "c",
			Id:          12,
			Action:      "GET",
			Name:        "CPU",
			ManagedType: "MO",
			ManagedId:   "SS",
			Address:     "127.0.0.1",
			Paths:       []P{{"q1", "q2"}},
			Params:      map[string]string{"a1": "34"},
			Body:        12,
		}, s: `{"channel":"c","request_id":12,"action":"GET","metric-name":"CPU","managed_type":"MO","managed_id":"SS","address":"127.0.0.1","paths":[["q1","q2"]],"params":{"a1":"34"},"body":12}`},
		{id: 13, o: &ExchangeRequest{ChannelName: "c",
			Id:          12,
			Action:      "GET",
			Name:        "CPU",
			ManagedType: "MO",
			ManagedId:   "SS",
			Address:     "127.0.0.1",
			Paths:       []P{{"q1", "q2"}, {"q3", "q4"}},
			Params:      map[string]string{"a1": "34"},
			Body:        "aer",
		}, s: `{"channel":"c","request_id":12,"action":"GET","metric-name":"CPU","managed_type":"MO","managed_id":"SS","address":"127.0.0.1","paths":[["q1","q2"],["q3","q4"]],"params":{"a1":"34"},"body":"aer"}`},
		{id: 13, o: &ExchangeRequest{ChannelName: "c",
			Id:          12,
			Action:      "GET",
			Name:        "CPU",
			ManagedType: "MO",
			ManagedId:   "SS",
			Address:     "127.0.0.1",
			Params:      map[string]string{"a1": "34"},
			Body:        "aer",
		}, s: `{"channel":"c","request_id":12,"action":"GET","metric-name":"CPU","managed_type":"MO","managed_id":"SS","address":"127.0.0.1","params":{"a1":"34"},"body":"aer"}`},
		{id: 13, o: &ExchangeRequest{ChannelName: "c",
			Id:          12,
			Action:      "GET",
			Name:        "CPU",
			ManagedType: "MO",
			ManagedId:   "SS",
			Address:     "127.0.0.1",
			Paths:       []P{{"q1", "q2"}, {"q3", "q4"}},
			Body:        "aer",
		}, s: `{"channel":"c","request_id":12,"action":"GET","metric-name":"CPU","managed_type":"MO","managed_id":"SS","address":"127.0.0.1","paths":[["q1","q2"],["q3","q4"]],"body":"aer"}`},
		{id: 13, o: &ExchangeRequest{ChannelName: "c",
			Id:          12,
			Action:      "GET",
			Name:        "CPU",
			ManagedType: "MO",
			ManagedId:   "SS",
			Address:     "127.0.0.1",
			Body:        map[string]string{"a1": "34"},
		}, s: `{"channel":"c","request_id":12,"action":"GET","metric-name":"CPU","managed_type":"MO","managed_id":"SS","address":"127.0.0.1","body":{"a1":"34"}}`},
		{id: 13, o: &ExchangeRequest{ChannelName: "c",
			Id:          12,
			Action:      "GET",
			Name:        "CPU",
			ManagedType: "MO",
			ManagedId:   "SS",
			Address:     "127.0.0.1",
		}, s: `{"channel":"c","request_id":12,"action":"GET","metric-name":"CPU","managed_type":"MO","managed_id":"SS","address":"127.0.0.1"}`}} {
		var buf bytes.Buffer
		if e := json.NewEncoder(&buf).Encode(test.o); nil != e {
			if nil == test.e || e.Error() != test.e.Error() {
				t.Error(e)
			}
			continue
		}

		var v interface{}
		if e := json.Unmarshal(buf.Bytes(), &v); nil != e {
			t.Error("umarshal failed,", e)
		}

		if test.s+"\n" != buf.String() {
			t.Error("excepted is not equals actual")
			t.Error("excepted is", test.s)
			t.Error("actual is `" + buf.String() + "`")
		}

	}
}

func TestBroker(t *testing.T) {
	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, err := client.DeleteBy("network_device", emptyParams)
		if nil != err {
			t.Error(err)
			return
		}

		res, e := batchGet(t, sampling_url+"/batch", []*ExchangeRequest{&ExchangeRequest{Address: "127.0.0.1", Action: "GET", Name: "sys", Params: map[string]string{"snmp.version": "v2c", "snmp.read_community": "public"}}})
		if nil != e {
			t.Error(e)
			return
		}
		for i := 0; i < 1000; i++ {
			res, e = batchGet(t, sampling_url+"/batch", []*ExchangeRequest{})
			if nil != e {
				t.Error(e)
				return
			}
			if nil != res && 0 != len(res) {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}

		if nil == res || 0 == len(res) {
			t.Error("not result")
			return
		}

		if nil == res[0].Evalue {
			t.Error("values is nil")
		}

		t.Log(res[0])

		res, e = batchGet(t, sampling_url+"/batch", []*ExchangeRequest{})
		if nil != e {
			t.Error(e)
			return
		}
		if nil != res && 0 != len(res) {
			t.Error("repeated result")
		}
	})
}
