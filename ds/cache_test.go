package ds

import (
	"commons"
	"commons/types"
	"fmt"
	"github.com/runner-mei/go-restful"
	"reflect"
	"testing"
	"time"
)

func TestCacheBasic(t *testing.T) {
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "network_device", map[string]string{})

		id1 := createMockDevice(t, client, "1")
		id2 := createMockDevice(t, client, "2")
		id3 := createMockDevice(t, client, "3")
		id4 := createMockDevice(t, client, "4")
		if "" == id1 {
			return
		}

		cache := NewCache(100*time.Minute, client, "network_device")
		defer cache.Close()

		t.Log("test Get")
		//fmt.Println("test Get")
		d1, _ := cache.Get(fmt.Sprint(id1))
		d2, _ := cache.Get(fmt.Sprint(id2))
		d3, _ := cache.Get(fmt.Sprint(id3))
		d4, _ := cache.Get(fmt.Sprint(id4))

		if nil == d1 {
			return
		}

		validMockDevice(t, client, "1", d1)
		validMockDevice(t, client, "2", d2)
		validMockDevice(t, client, "3", d3)
		validMockDevice(t, client, "4", d4)

		// test Refresh
		t.Log("test Refresh")
		//fmt.Println("test Refresh")
		updateMockDevice(t, client, id1, "11")
		updateMockDevice(t, client, id2, "21")
		updateMockDevice(t, client, id3, "31")
		updateMockDevice(t, client, id4, "41")

		d1, _ = cache.Get(fmt.Sprint(id1))
		d2, _ = cache.Get(fmt.Sprint(id2))
		d3, _ = cache.Get(fmt.Sprint(id3))
		d4, _ = cache.Get(fmt.Sprint(id4))
		validMockDevice(t, client, "1", d1)
		validMockDevice(t, client, "2", d2)
		validMockDevice(t, client, "3", d3)
		validMockDevice(t, client, "4", d4)

		cache.Refresh()

		d1, _ = cache.Get(fmt.Sprint(id1))
		d2, _ = cache.Get(fmt.Sprint(id2))
		d3, _ = cache.Get(fmt.Sprint(id3))
		d4, _ = cache.Get(fmt.Sprint(id4))
		validMockDevice(t, client, "11", d1)
		validMockDevice(t, client, "21", d2)
		validMockDevice(t, client, "31", d3)
		validMockDevice(t, client, "41", d4)

		t.Log("test Delete")
		//fmt.Println("test Delete")
		updateMockDevice(t, client, id1, "111")
		updateMockDevice(t, client, id2, "211")
		updateMockDevice(t, client, id3, "311")
		updateMockDevice(t, client, id4, "411")

		d1, _ = cache.Get(fmt.Sprint(id1))
		d2, _ = cache.Get(fmt.Sprint(id2))
		d3, _ = cache.Get(fmt.Sprint(id3))
		d4, _ = cache.Get(fmt.Sprint(id4))
		validMockDevice(t, client, "11", d1)
		validMockDevice(t, client, "21", d2)
		validMockDevice(t, client, "31", d3)
		validMockDevice(t, client, "41", d4)

		cache.Delete(fmt.Sprint(id1))
		cache.Delete(fmt.Sprint(id2))
		cache.Delete(fmt.Sprint(id3))
		cache.Delete(fmt.Sprint(id4))

		d1, _ = cache.Get(fmt.Sprint(id1))
		d2, _ = cache.Get(fmt.Sprint(id2))
		d3, _ = cache.Get(fmt.Sprint(id3))
		//d4, _ = cache.Get(fmt.Sprint(id4))
		validMockDevice(t, client, "111", d1)
		validMockDevice(t, client, "211", d2)
		validMockDevice(t, client, "311", d3)
		validMockDevice(t, client, "41", d4)
	})
}

func TestCacheBySuper(t *testing.T) {
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "network_device", map[string]string{})

		id1 := createMockDevice(t, client, "1")
		id2 := createMockDevice(t, client, "2")
		id3 := createMockDevice(t, client, "3")
		id4 := createMockDevice(t, client, "4")
		if "" == id1 {
			return
		}

		cache := NewCache(100*time.Minute, client, "managed_object")
		defer cache.Close()

		d1, _ := cache.Get(fmt.Sprint(id1))
		d2, _ := cache.Get(fmt.Sprint(id2))
		d3, _ := cache.Get(fmt.Sprint(id3))
		d4, _ := cache.Get(fmt.Sprint(id4))

		if nil == d1 {
			return
		}

		validMockDevice(t, client, "1", d1)
		validMockDevice(t, client, "2", d2)
		validMockDevice(t, client, "3", d3)
		validMockDevice(t, client, "4", d4)
	})
}

func TestCacheAlreadyDelete(t *testing.T) {
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "network_device", map[string]string{})

		id1 := createMockDevice(t, client, "1")
		id2 := createMockDevice(t, client, "2")
		id3 := createMockDevice(t, client, "3")
		id4 := createMockDevice(t, client, "4")
		if "" == id1 {
			return
		}

		cache := NewCache(100*time.Minute, client, "network_device")
		defer cache.Close()
		d1, _ := cache.Get(fmt.Sprint(id1))
		d2, _ := cache.Get(fmt.Sprint(id2))
		d3, _ := cache.Get(fmt.Sprint(id3))
		d4, _ := cache.Get(fmt.Sprint(id4))

		if nil == d1 {
			return
		}

		validMockDevice(t, client, "1", d1)
		validMockDevice(t, client, "2", d2)
		validMockDevice(t, client, "3", d3)
		validMockDevice(t, client, "4", d4)

		deleteById(t, client, "network_device", id4)

		d1, _ = cache.Get(fmt.Sprint(id1))
		d2, _ = cache.Get(fmt.Sprint(id2))
		d3, _ = cache.Get(fmt.Sprint(id3))
		d4, _ = cache.Get(fmt.Sprint(id4))

		validMockDevice(t, client, "1", d1)
		validMockDevice(t, client, "2", d2)
		validMockDevice(t, client, "3", d3)
		validMockDevice(t, client, "4", d4)

		cache.Refresh()

		d1, _ = cache.Get(fmt.Sprint(id1))
		d2, _ = cache.Get(fmt.Sprint(id2))
		d3, _ = cache.Get(fmt.Sprint(id3))
		d4, _ = cache.Get(fmt.Sprint(id4))
		validMockDevice(t, client, "1", d1)
		validMockDevice(t, client, "2", d2)
		validMockDevice(t, client, "3", d3)
		if nil != d4 {
			t.Error("d4 is not delete from cache")
		}
	})
}

func TestCacheAdd(t *testing.T) {
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "network_device", map[string]string{})

		id1 := createMockDevice(t, client, "1")
		if "" == id1 {
			return
		}

		cache := NewCache(100*time.Minute, client, "network_device")
		defer cache.Close()
		d1, _ := cache.Get(fmt.Sprint(id1))
		if nil == d1 {
			return
		}

		validMockDevice(t, client, "1", d1)
		cache.Refresh()

		id4 := createMockDevice(t, client, "4")

		d1, _ = cache.Get(fmt.Sprint(id1))
		d4, _ := cache.Get(fmt.Sprint(id4))

		validMockDevice(t, client, "1", d1)
		validMockDevice(t, client, "4", d4)
	})
}

func TestCacheRefresh(t *testing.T) {
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "network_device", map[string]string{})

		id1 := createMockDevice(t, client, "1")
		id2 := createMockDevice(t, client, "2")
		id3 := createMockDevice(t, client, "3")
		if "" == id1 {
			t.Error("device1 create failed")
			return
		}

		cache := NewCache(100*time.Minute, client, "network_device")
		defer cache.Close()
		d1, _ := cache.Get(fmt.Sprint(id1))
		d2, _ := cache.Get(fmt.Sprint(id2))
		d3, _ := cache.Get(fmt.Sprint(id3))

		if nil == d1 {
			t.Error("device1 find failed")
			return
		}
		if nil == d2 {
			t.Error("device2 find failed")
			return
		}
		if nil == d3 {
			t.Error("device3 find failed")
			return
		}

		updateMockDevice(t, client, id2, "211")
		deleteById(t, client, "network_device", id3)
		id4 := createMockDevice(t, client, "4")
		cache.Refresh()

		messages := make([]string, 0, 3)
		excepted := []string{"GET,/network_device/2", "GET,/network_device/3", "GET,/network_device/4"}

		ws_instance.Filter(func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
			messages = append(messages, fmt.Sprintf("%s,%s", req.Request.Method, req.Request.URL))
			chain.ProcessFilter(req, resp)
		})

		d1, _ = cache.Get(fmt.Sprint(id1))
		d2, _ = cache.Get(fmt.Sprint(id2))
		d3, _ = cache.Get(fmt.Sprint(id3))
		d4, _ := cache.Get(fmt.Sprint(id4))

		validMockDevice(t, client, "1", d1)
		validMockDevice(t, client, "211", d2)
		validMockDevice(t, client, "4", d4)
		if nil != d3 {
			t.Error("d3 is not delete from cache")
		}

		if !reflect.DeepEqual(excepted, messages) {
			t.Errorf("3 != len(messages), excepted is %v, actual is %v", excepted, messages)
		}

	})
}

func TestCacheClose(t *testing.T) {
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		e := func() (e error) {
			defer func() {
				if o := recover(); nil != o {
					e = fmt.Errorf("%v", o)
				}
			}()

			cache := NewCache(100*time.Minute, client, "network_device")
			cache.Close()
			cache.Get("df")
			return
		}()

		if nil == e {
			t.Error("except panic, but no panic")
		}
	})
}

func TestCachesBasic(t *testing.T) {
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "network_device", map[string]string{})

		id1 := createMockDevice(t, client, "1")
		if "" == id1 {
			return
		}

		caches := NewCaches(100*time.Minute, client, "", nil)
		defer caches.Close()

		messages := make([]string, 0, 3)
		excepted := []string{"GET,/network_device/@count", "GET,/network_device/1"}

		ws_instance.Filter(func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
			messages = append(messages, fmt.Sprintf("%s,%s", req.Request.Method, req.Request.URL))
			chain.ProcessFilter(req, resp)
		})

		cache1, _ := caches.GetCache("network_device")
		d1, _ := cache1.Get(fmt.Sprint(id1))

		if nil == d1 {
			return
		}

		validMockDevice(t, client, "1", d1)

		cache2, _ := caches.GetCache("network_device")

		if !reflect.DeepEqual(excepted, messages) {
			t.Errorf("excepted_messages != actual_messages, excepted is %v, actual is %v", excepted, messages)
		}

		if cache1 != cache2 {
			t.Error("cache1 != cache2")
		}

	})
}

func TestCachesBasicAlias(t *testing.T) {
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "network_device", map[string]string{})

		id1 := createMockDevice(t, client, "1")
		if "" == id1 {
			return
		}

		caches := NewCaches(100*time.Minute, client, "", map[string]string{"d": "network_device"})
		defer caches.Close()

		messages := make([]string, 0, 3)
		excepted := []string{"GET,/network_device/@count", "GET,/network_device/1"}

		ws_instance.Filter(func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
			messages = append(messages, fmt.Sprintf("%s,%s", req.Request.Method, req.Request.URL))
			chain.ProcessFilter(req, resp)
		})

		cache1, _ := caches.GetCache("d")
		d1, _ := cache1.Get(fmt.Sprint(id1))

		if nil == d1 {
			return
		}

		validMockDevice(t, client, "1", d1)

		cache2, _ := caches.GetCache("d")

		if !reflect.DeepEqual(excepted, messages) {
			t.Errorf("excepted_messages != actual_messages, excepted is %v, actual is %v", excepted, messages)
		}

		if cache1 != cache2 {
			t.Error("cache1 != cache2")
		}

	})
}

func createSnmpParamsForCache(t *testing.T, client *Client, id, factor string) string {
	return createJson(t, client, "snmp_param", fmt.Sprintf(`{ "port":%v, "managed_object_id":%v, "version":"snmp_v2c", "read_community":"aa"}`, factor, id))
}

func TestCacheGetChildren(t *testing.T) {
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "network_device", map[string]string{})

		id1 := createMockDevice(t, client, "1")
		createSnmpParamsForCache(t, client, id1, "11")
		createSnmpParamsForCache(t, client, id1, "12")
		id2 := createMockDevice(t, client, "2")
		createSnmpParamsForCache(t, client, id2, "11")
		createSnmpParamsForCache(t, client, id2, "12")

		if "" == id1 {
			return
		}

		cache := NewCacheWithIncludes(100*time.Minute, client, "network_device", "*")
		defer cache.Close()

		d1_11, _ := cache.GetChildren(fmt.Sprint(id1), "attributes", map[string]commons.Matcher{"type": commons.EqualString("snmp_param"),
			"port": commons.EqualInt(11)})
		d1_12, _ := cache.GetChildren(fmt.Sprint(id1), "attributes", map[string]commons.Matcher{"type": commons.EqualString("snmp_param"),
			"port": commons.EqualInt(12)})

		d2_11, _ := cache.GetChildren(fmt.Sprint(id2), "attributes", map[string]commons.Matcher{"type": commons.EqualString("snmp_param"),
			"port": commons.EqualInt(11)})
		d2_12, _ := cache.GetChildren(fmt.Sprint(id2), "attributes", map[string]commons.Matcher{"type": commons.EqualString("snmp_param"),
			"port": commons.EqualInt(12)})

		if nil == d1_11 || 0 == len(d1_11) || nil == d1_11[0] {
			t.Error("d1_11 is nil")
		} else {
			validMockSNMP(t, client, "11", d1_11[0])
		}

		if nil == d1_12 || 0 == len(d1_12) || nil == d1_12[0] {
			t.Error("d1_12 is nil")
		} else {
			validMockSNMP(t, client, "12", d1_12[0])
		}

		if nil == d2_11 || 0 == len(d2_11) || nil == d2_11[0] {
			t.Error("d2_11 is nil")
		} else {
			validMockSNMP(t, client, "11", d2_11[0])
		}

		if nil == d2_12 || 0 == len(d2_12) || nil == d2_12[0] {
			t.Error("d2_12 is nil")
		} else {
			validMockSNMP(t, client, "12", d2_12[0])
		}

	})
}

func TestCachesFailed(t *testing.T) {
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {

		messages := make([]string, 0, 3)
		excepted := []string{"GET,/devdddice/@count"}

		ws_instance.Filter(func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
			messages = append(messages, fmt.Sprintf("%s,%s", req.Request.Method, req.Request.URL))
			chain.ProcessFilter(req, resp)
		})

		caches := NewCaches(100*time.Minute, client, "", nil)
		defer caches.Close()
		cache1, e1 := caches.GetCache("devdddice")
		cache2, e2 := caches.GetCache("devdddice")

		if nil != cache1 {
			t.Error("cache1 is not nil")
		}

		if nil != cache2 {
			t.Error("cache2 is not nil")
		}

		if nil != e1 {
			t.Error("e1 is not nil")
		}

		if nil != e2 {
			t.Error("e2 is not nil")
		}

		if !reflect.DeepEqual(excepted, messages) {
			t.Errorf("excepted_messages != actual_messages, excepted is %v, actual is %v", excepted, messages)
		}

	})
}

func TestCachesNetworkFailed(t *testing.T) {
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {

		caches := NewCaches(100*time.Minute, NewClient("http://127.0.0.1:803"), "", nil)
		defer caches.Close()
		cache1, e1 := caches.GetCache("devdddice")
		cache2, e2 := caches.GetCache("devdddice")

		if nil != cache1 {
			t.Error("cache1 is not nil")
		}

		if nil != cache2 {
			t.Error("cache2 is not nil")
		}

		if nil == e1 {
			t.Error("e1 is nil")
		}

		if nil == e2 {
			t.Error("e2 is nil")
		}

	})
}

func TestCachesClose(t *testing.T) {
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		e := func() (e error) {
			defer func() {
				if o := recover(); nil != o {
					e = fmt.Errorf("%v", o)
				}
			}()

			caches := NewCaches(100*time.Minute, client, "", nil)
			caches.Close()
			caches.GetCache("df")

			return
		}()

		if nil == e {
			t.Error("except panic, but no panic")
		}
	})
}

func TestCacheCloseInCaches(t *testing.T) {
	SrvTest(t, "etc/tpt_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		e := func() (e error) {
			defer func() {
				if o := recover(); nil != o {
					e = fmt.Errorf("%v", o)
				}
			}()

			caches := NewCaches(100*time.Minute, client, "", nil)
			cache, _ := caches.GetCache("network_device")
			caches.Close()
			v, e := cache.Get("1")
			t.Log(v, e)
			return
		}()

		if nil == e {
			t.Error("except panic, but no panic")
		}
	})
}
