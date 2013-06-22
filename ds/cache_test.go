package ds

import (
	"commons/types"
	"fmt"
	"github.com/runner-mei/go-restful"
	"reflect"
	"testing"
	"time"
)

func TestCacheBasic(t *testing.T) {
	srvTest2(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "device", map[string]string{})

		id1 := createMockDevice(t, client, "1")
		id2 := createMockDevice(t, client, "2")
		id3 := createMockDevice(t, client, "3")
		id4 := createMockDevice(t, client, "4")
		if "" == id1 {
			return
		}

		cache := NewCache(100*time.Minute, client, "device")

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

func TestCacheAlreadyDelete(t *testing.T) {
	srvTest2(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "device", map[string]string{})

		id1 := createMockDevice(t, client, "1")
		id2 := createMockDevice(t, client, "2")
		id3 := createMockDevice(t, client, "3")
		id4 := createMockDevice(t, client, "4")
		if "" == id1 {
			return
		}

		cache := NewCache(100*time.Minute, client, "device")
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

		deleteById(t, client, "device", id4)

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
	srvTest2(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "device", map[string]string{})

		id1 := createMockDevice(t, client, "1")
		if "" == id1 {
			return
		}

		cache := NewCache(100*time.Minute, client, "device")
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
	srvTest2(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		deleteBy(t, client, "device", map[string]string{})

		id1 := createMockDevice(t, client, "1")
		id2 := createMockDevice(t, client, "2")
		id3 := createMockDevice(t, client, "3")
		if "" == id1 {
			t.Error("device1 create failed")
			return
		}

		cache := NewCache(100*time.Minute, client, "device")
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
		deleteById(t, client, "device", id3)
		id4 := createMockDevice(t, client, "4")
		cache.Refresh()

		messages := make([]string, 0, 3)
		excepted := []string{"GET,/device/2", "GET,/device/3", "GET,/device/4"}

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
