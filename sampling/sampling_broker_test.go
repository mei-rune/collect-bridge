package sampling

import (
	"bytes"
	"commons"
	"commons/types"
	ds "data_store"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
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

func TestBrokerWithSubscribe(t *testing.T) {
	for test_idx, test := range []struct {
		metric_name string
		e           error
	}{{metric_name: "sys"},
		{metric_name: "aaa", e: errors.New("'aaa' is not acceptable.")},
		{metric_name: "icmp", e: errors.New("sampled is pending.")}} {
		SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
			_, err := client.DeleteBy("network_device", emptyParams)
			if nil != err {
				t.Error("test[", test_idx, "] "+err.Error())
				return
			}

			id := ds.CreateItForTest(t, client, "network_device", mo)
			ds.CreateItByParentForTest(t, client, "network_device", id, "wbem_param", wbem_params)
			ds.CreateItByParentForTest(t, client, "network_device", id, "snmp_param", snmp_params)

			broker, e := NewBroker("test", sampling_url+"/batch")
			if nil != e {
				t.Error("test[", test_idx, "] "+e.Error())
				return
			}
			defer broker.Close()

			c_list := []chan interface{}{make(chan interface{}, 10), make(chan interface{}, 10)}
			cl_list := []ChannelClient{nil, nil}
			for i := 0; i < 2; i++ {
				cl, e := broker.SubscribeClient("aa", c_list[i], "GET", test.metric_name, "managed_object", id, "", nil, nil, 0)
				if nil != e {
					t.Error("test[", test_idx, "] "+e.Error())
					return
				}
				cl_list[i] = cl
			}
			defer func() {
				for _, c := range cl_list {
					c.Close()
				}
			}()

			cl_list[0].Send()

			for _, c := range c_list {
				select {
				case res := <-c:
					resp, ok := res.(*ExchangeResponse)
					if !ok {
						t.Error("test[", test_idx, "]values is nil")
						break
					}
					if nil != test.e {
						if test.e.Error() != resp.Error().Error() {
							t.Error("test[", test_idx, "]error message is not excepted")
							t.Error("test[", test_idx, "]excepted is", test.e)
							t.Error("test[", test_idx, "]actual is", resp.Error())
						}
					} else {
						if nil == resp.InterfaceValue() {
							t.Error("test[", test_idx, "]values is nil")
						}
					}

				case <-time.After(10 * time.Second):
					t.Error("test[", test_idx, "]timeout")
				}
			}

			for _, cl := range cl_list[1:] {
				cl.Close()
			}

			cl_list[0].Send()

			for idx, c := range c_list {
				if 0 == idx {
					select {
					case res := <-c:
						resp, ok := res.(*ExchangeResponse)
						if !ok {
							t.Error("test[", test_idx, "]values is nil")
							break
						}
						if nil != test.e {
							if nil == resp.Error() {
								if "icmp" != test.metric_name {
									t.Error("test[", test_idx, "] error message is not excepted, it is nil.")
									t.Error(resp)
								}
							} else if test.e.Error() != resp.Error().Error() {
								t.Error("test[", test_idx, "]error message is not excepted")
								t.Error("test[", test_idx, "]excepted is", test.e)
								t.Error("test[", test_idx, "]actual is", resp.Error())
							}
						} else {
							if nil == resp.InterfaceValue() {
								t.Error("test[", test_idx, "]values is nil,", resp.Error())
							}
						}
					case <-time.After(10 * time.Second):
						t.Error("test[", test_idx, "]timeout")
					}
				} else {
					select {
					case <-c:
						t.Error("test[", test_idx, "]error recv")
					default:
					}
				}
			}

		})
	}
}

func TestBrokerWithSubscribeAndError(t *testing.T) {
	excepted_error := "sdfsdfsdfsdfsdf"
	hsrv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(excepted_error))
	}))
	defer hsrv.Close()

	test_idx := 0

	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, err := client.DeleteBy("network_device", emptyParams)
		if nil != err {
			t.Error("test[", test_idx, "] "+err.Error())
			return
		}

		id := ds.CreateItForTest(t, client, "network_device", mo)
		ds.CreateItByParentForTest(t, client, "network_device", id, "wbem_param", wbem_params)
		ds.CreateItByParentForTest(t, client, "network_device", id, "snmp_param", snmp_params)

		broker, e := NewBroker("test", hsrv.URL+"/batch")
		if nil != e {
			t.Error("test[", test_idx, "] "+e.Error())
			return
		}
		defer broker.Close()

		c_list := []chan interface{}{make(chan interface{}, 10), make(chan interface{}, 10)}
		cl_list := []ChannelClient{nil, nil}
		for i := 0; i < 2; i++ {
			cl, e := broker.SubscribeClient("aa", c_list[i], "GET", "cpu", "managed_object", id, "", nil, nil, 0)
			if nil != e {
				t.Error("test[", test_idx, "] "+e.Error())
				return
			}
			cl_list[i] = cl
		}
		defer func() {
			for _, c := range cl_list {
				c.Close()
			}
		}()

		check := func() {
			for idx, c := range c_list {
				if 0 == idx {
					select {
					case res := <-c:
						resp, ok := res.(*ExchangeResponse)
						if !ok {
							t.Error("test[", test_idx, "]values is nil")
							break
						}

						if !strings.Contains(resp.Error().Error(), excepted_error) {
							t.Error("test[", test_idx, "]error message is not excepted")
							t.Error("test[", test_idx, "]excepted is", excepted_error)
							t.Error("test[", test_idx, "]actual is", resp.Error().Error())
						}

					case <-time.After(10 * time.Second):
						t.Error("test[", test_idx, "]timeout")
					}
				} else {
					select {
					case <-c:
						t.Error("test[", test_idx, "]error recv")
					default:
					}
				}
			}
		}

		cl_list[0].Send()
		check()
		for _, cl := range cl_list[1:] {
			cl.Close()
		}

		cl_list[0].Send()
		check()
	})
}

func TestBrokerWithInvoke(t *testing.T) {
	for _, test := range []struct {
		metric_name string
		e           error
	}{{metric_name: "sys"},
		{metric_name: "aaa", e: errors.New("'aaa' is not acceptable.")},
		{metric_name: "icmp", e: errors.New("sampled is pending.")}} {
		SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
			_, err := client.DeleteBy("network_device", emptyParams)
			if nil != err {
				t.Error(err)
				return
			}

			id := ds.CreateItForTest(t, client, "network_device", mo)
			ds.CreateItByParentForTest(t, client, "network_device", id, "wbem_param", wbem_params)
			ds.CreateItByParentForTest(t, client, "network_device", id, "snmp_param", snmp_params)

			broker, e := NewBroker("test", sampling_url+"/batch")
			if nil != e {
				t.Error(e)
				return
			}
			defer broker.Close()

			cl, e := broker.CreateClient("GET", test.metric_name, "managed_object", id, "", nil, nil)
			if nil != e {
				t.Error(e)
				return
			}

			res, e := cl.Invoke(10 * time.Second)
			if nil != e {
				t.Error(e)
				return
			}

			resp, ok := res.(*ExchangeResponse)
			if !ok {
				t.Error("values is not a ExchangeResponse")
				return
			}
			if nil != test.e {
				if test.e.Error() != resp.Error().Error() {
					t.Error("error message is not excepted")
					t.Error("excepted is", test.e)
					t.Error("actual is", resp.Error())
				}
			} else {
				if nil == resp.InterfaceValue() {
					t.Error("values is nil")
				}
			}
		})
	}
}

func TestBrokerWithRemoveChannel(t *testing.T) {
	result := []*ExchangeResponse{&ExchangeResponse{Evalue: map[string]interface{}{"name": "this is a name", "a": "b"}}}
	called := int32(0)

	hsrv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&called, 1)
		resp.WriteHeader(http.StatusAccepted)
		result[0].ChannelName = "aa"
		if e := json.NewEncoder(resp).Encode(result); nil != e {
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte(e.Error()))
		}
	}))

	defer hsrv.Close()

	broker, e := NewBroker("test", hsrv.URL)
	if nil != e {
		t.Error("create broker failed,", e.Error())
		return
	}
	defer broker.Close()

	for i := 0; i < 10; i++ {
		c := make(chan interface{}, 10)
		cl, e := broker.SubscribeClient("aa", c, "GET", "cpu", "managed_object", "12", "", nil, nil, 0)
		if nil != e {
			t.Error("create client failed", e.Error())
			return
		}
		close(c)
		if i == 9 {
			cl.Send()
		}
	}

	time.Sleep(1000 * time.Millisecond)

	if 0 != len(broker.channelGroups) {
		t.Error("not remove client", len(broker.channelGroups))
	}
}

func TestBrokerWithMergeRequestInClient(t *testing.T) {
	called := int32(0)

	hsrv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if nil == req.Body {
			resp.WriteHeader(http.StatusNoContent)
			return
		}
		var er []*ExchangeRequest
		if e := json.NewDecoder(req.Body).Decode(&er); nil != e {
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte(e.Error()))
			return
		}
		if nil == er {
			resp.WriteHeader(http.StatusNoContent)
			return
		}

		atomic.AddInt32(&called, int32(len(er)))

		resp.WriteHeader(http.StatusNoContent)
	}))

	defer hsrv.Close()

	broker, e := NewBroker("test", hsrv.URL)
	if nil != e {
		t.Error("create broker failed,", e.Error())
		return
	}
	defer broker.Close()
	c := make(chan interface{}, 10)
	cl, e := broker.SubscribeClient("aa", c, "GET", "cpu", "managed_object", "12", "", nil, nil, 5*time.Second)
	if nil != e {
		t.Error("create client failed", e.Error())
		return
	}
	defer cl.Close()

	for i := 0; i < 10; i++ {
		cl.Send()
	}
	time.Sleep(1000 * time.Millisecond)
	if 1 != atomic.LoadInt32(&called) {
		t.Error("not merge - ", atomic.LoadInt32(&called))
	}
}

func TestBrokerWithMergeRequestInServer(t *testing.T) {

	called := int32(0)
	var handler MockHandler = func() commons.Result {
		time.Sleep(100 * time.Millisecond)
		return commons.Return(atomic.AddInt32(&called, 1))
	}
	Methods["test_handler"] = newRouteSpec("get", "TestBrokerWithMergeRequestInRequest", "the mem of cisco", Match().Build(),
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			return handler, nil
		})

	SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, sampling_url string, definitions *types.TableDefinitions) {
		_, err := client.DeleteBy("network_device", emptyParams)
		if nil != err {
			t.Error(err)
			return
		}

		id := ds.CreateItForTest(t, client, "network_device", mo)
		ds.CreateItByParentForTest(t, client, "network_device", id, "wbem_param", wbem_params)
		ds.CreateItByParentForTest(t, client, "network_device", id, "snmp_param", snmp_params)

		request := &ExchangeRequest{ChannelName: "aa",
			Address: "127.0.0.1",
			Action:  "GET",
			Name:    "TestBrokerWithMergeRequestInRequest",
			Params:  map[string]string{"snmp.version": "v2c", "snmp.read_community": "public"}}

		_, e := batchGet(t, sampling_url+"/batch", []*ExchangeRequest{request, request, request, request, request, request, request, request})
		if nil != e {
			t.Error(e)
			return
		}

		time.Sleep(1 * time.Second)

		if 1 != atomic.LoadInt32(&called) {
			t.Error("called count is error, ", atomic.LoadInt32(&called))
		}
	})
}
