package sampling

// import (
// 	"commons"
// 	"commons/types"
// 	ds "data_store"
// 	"testing"
// 	"time"
// )

var (
	emptyParams = map[string]string{}
)

// func TestContextBasic(t *testing.T) {
// 	ds.SrvTest(t, "../data_store/etc/tpt_models.xml", func(client *ds.Client, definitions *types.TableDefinitions) {
// 		_, e := client.DeleteBy("access_param", emptyParams)
// 		if nil != e {
// 			t.Error(e)
// 			return
// 		}

// 		id := ds.CreateMockDeviceForTest(t, client, "1")
// 		ds.CreateMockSshParamsForTest(t, client, "sss")
// 		ds.CreateMockSnmpParamsForTest(t, client, "ccc")

// 		lazy_map := &context{managed_type: "device",
// 			managed_id: id,
// 			mo:         commons.InterfaceMap(map[string]interface{}{}),
// 			caches:     ds.NewCaches(100*time.Minute, client, nil),
// 			local:      map[string]commons.Map{}}
// 		s, err := lazy_map.GetString("device.name")
// 		if nil != err {
// 			t.Error(err)
// 			return
// 		}
// 		if s != "dd1" {
// 			t.Errorf("name is error, excepted is %v, actual is %v", "dd1", s)
// 		}

// 		s, err = lazy_map.GetString("device.name")
// 		if nil != err {
// 			t.Error(err)
// 			return
// 		}
// 		if s != "dd1" {
// 			t.Errorf("name is error, excepted is %v, actual is %v", "dd1", s)
// 		}

// 		s, err = lazy_map.GetString("snmp_param.read_community")
// 		if nil != err {
// 			t.Error(err)
// 			return
// 		}
// 		if s != "ccc" {
// 			t.Errorf("name is error, excepted is %v, actual is %v", "dd1", s)
// 		}
// 	})
// }
