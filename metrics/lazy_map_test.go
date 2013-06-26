package metrics

import (
	"commons/types"
	"ds"
	"testing"
	"time"
)

var (
	emptyParams = map[string]string{}
)

func TestLazyMapBasic(t *testing.T) {
	ds.SrvTest(t, "../ds/etc/mj_models.xml", func(client *ds.Client, definitions *types.TableDefinitions) {
		_, e := client.DeleteBy("access_param", emptyParams)
		if nil != e {
			t.Error(e)
			return
		}

		ds.CreateMockDeviceForTest(t, client, "1")
		lazy_map := &context{managed_type: "device", managed_id: "1", caches: ds.NewCaches(100*time.Minute, client, nil)}
		s := lazy_map.GetStringWithDefault("device#name", "")
		if s != "dd1" {
			t.Errorf("name is error, excepted is %v, actual is %v", "dd1", s)
		}

		s, _ = lazy_map.GetString("device#name")
		if s != "dd1" {
			t.Errorf("name is error, excepted is %v, actual is %v", "dd1", s)
		}
	})
}
