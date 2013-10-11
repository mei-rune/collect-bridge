package snmp

import (
	"fmt"
	"testing"
)

func TestRequestBuffer(t *testing.T) {
	cb := newRequestBuffer(make([]*mgrRequest, 10))

	check := func(cb *requestBuffer, c int) {
		if c < 10 {
			if cb.Size() != (1 + c) {
				t.Error("size is error, excepted is", 1+c, ", actual is", cb.Size())
			}

			all := cb.All()
			if len(all) != (1 + c) {
				t.Error("len(all) is error, excepted is 10, actual is", cb.Size())
			}

			for i := 0; i <= c; i++ {
				if all[i].host != fmt.Sprint(i) {
					t.Error("all[", i, "] is error, excepted is ", all[i].host, ", actual is", i)
				}
			}

			for i := 0; i <= c; i++ {
				if all[i].host != cb.Get(i).host {
					t.Error("all[", i, "] != cb.Get(", i, "), excepted is ", all[i].host, ", actual is", cb.Get(i).host)
				}
			}

			if fmt.Sprint(c) != cb.Last().host {
				t.Error("excepted last is", c, ", actual is", cb.Last().host)
			}

			if all[0].host != cb.First().host {
				t.Error("excepted first is", all[0].host, ", actual is", cb.First().host)
			}

		} else {
			if cb.Size() != 10 {
				t.Error("size is error, excepted is 10, actual is", cb.Size())
			}

			all := cb.All()
			if len(all) != 10 {
				t.Error("len(all) is error, excepted is 10, actual is", cb.Size())
			}

			for i := 0; i < 10; i++ {
				if all[i].host != fmt.Sprint(c-9+i) {
					t.Error("all[", i, "] is error, excepted is", all[i].host, ", actual is", c-9+i)
				}
			}

			for i := 0; i < 10; i++ {
				if all[i].host != cb.Get(i).host {
					t.Error("all[", i, "] != cb.Get(", i, "), excepted is ", all[i].host, ", actual is", cb.Get(i).host)
				}
			}

			if fmt.Sprint(c) != cb.Last().host {
				t.Error("excepted last is", c, ", actual is", cb.Last().host)
			}

			if fmt.Sprint(c-9) != cb.First().host {
				t.Error("excepted first is", c-9, ", actual is", cb.First().host)
			}

			if all[0].host != cb.First().host {
				t.Error("excepted first is", all[0].host, ", actual is", cb.First().host)
			}

			if all[len(all)-1].host != cb.Last().host {
				t.Error("excepted first is", all[len(all)-1].host, ", actual is", cb.Last().host)
			}
		}
	}

	for i := 0; i < 100; i++ {
		cb.Push(&mgrRequest{host: fmt.Sprint(i)})
		check(cb, i)
	}
}
