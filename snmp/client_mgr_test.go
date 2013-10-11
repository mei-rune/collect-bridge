package snmp

import (
	"testing"
	"time"
)

func TestClientMgrTimeout(t *testing.T) {
	is_unit_test_for_client_mgr = true
	var mgr *ClientManager = nil
	defer func() {
		is_unit_test_for_client_mgr = false
		if nil != mgr {
			mgr.Close()
		}
	}()

	mgr = &ClientManager{}

	err := mgr.init()
	if nil != err {
		t.Error("start failed, ", err)
		return
	}

	i1, _ := mgr.GetClient("1")
	i1_copy, _ := mgr.GetClient("1")
	i2, _ := mgr.GetClient("2")

	c1 := i1.(*TestClient)
	c1_copy := i1_copy.(*TestClient)
	c2 := i2.(*TestClient)

	if c1 != c1_copy {
		t.Error("c1 != c1_copy")
		return
	}

	for i := 0; i < 19 && 0 != len(mgr.clients); i++ {
		time.Sleep(time.Second)
	}

	if 0 != len(mgr.clients) {
		t.Error("client is not timeout")
		return
	}

	if true != c1.stop || true != c1.test {
		t.Errorf("func invoke failed,  stop = %v, test=%v", c1.stop, c1.test)
		return
	}

	if true != c2.stop || true != c2.test {
		t.Errorf("func invoke failed, stop = %v, test=%v", c2.stop, c2.test)
		return
	}
}
