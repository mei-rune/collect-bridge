package sampling

import (
	"testing"
)

func TestPendingBuffer(t *testing.T) {
	cb := newPendingBuffer(make([]testingRequst, 10))

	check := func(cb *snmpPendingBuffer, c int) {
		if c < 10 {
			if cb.Size() != (1 + c) {
				t.Error("size is error, excepted is", 1+c, ", actual is", cb.Size())
			}

			all := cb.All()
			if len(all) != (1 + c) {
				t.Error("len(all) is error, excepted is 10, actual is", cb.Size())
			}

			for i := 0; i <= c; i++ {
				if all[i].id != (i) {
					t.Error("all[", i, "] is error, excepted is ", all[i].id, ", actual is", i)
				}
			}

			for i := 0; i <= c; i++ {
				if all[i].id != cb.Get(i).id {
					t.Error("all[", i, "] != cb.Get(", i, "), excepted is ", all[i].id, ", actual is", cb.Get(i).id)
				}
			}

			if c != cb.Last().id {
				t.Error("excepted last is", c, ", actual is", cb.Last().id)
			}

			if all[0].id != cb.First().id {
				t.Error("excepted first is", all[0].id, ", actual is", cb.First().id)
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
				if all[i].id != (c - 9 + i) {
					t.Error("all[", i, "] is error, excepted is", all[i].id, ", actual is", c-9+i)
				}
			}

			for i := 0; i < 10; i++ {
				if all[i].id != cb.Get(i).id {
					t.Error("all[", i, "] != cb.Get(", i, "), excepted is ", all[i].id, ", actual is", cb.Get(i).id)
				}
			}

			if c != cb.Last().id {
				t.Error("excepted last is", c, ", actual is", cb.Last().id)
			}

			if (c - 9) != cb.First().id {
				t.Error("excepted first is", c-9, ", actual is", cb.First().id)
			}

			if all[0].id != cb.First().id {
				t.Error("excepted first is", all[0].id, ", actual is", cb.First().id)
			}

			if all[len(all)-1].id != cb.Last().id {
				t.Error("excepted first is", all[len(all)-1].id, ", actual is", cb.Last().id)
			}
		}
	}

	for i := 0; i < 100; i++ {
		cb.Push(testingRequst{id: i})
		check(cb, i)
	}
}
