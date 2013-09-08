package commons

import (
	"testing"
)

func TestInt64Buffer(t *testing.T) {
	cb := NewInt64Buffer(make([]int64, 10))

	check := func(cb *Int64Buffer, c int) {
		if c < 10 {
			if cb.Size() != (1 + c) {
				t.Error("size is error, excepted is", 1+c, ", actual is", cb.Size())
			}

			all := cb.All()
			if len(all) != (1 + c) {
				t.Error("len(all) is error, excepted is 10, actual is", cb.Size())
			}

			for i := 0; i <= c; i++ {
				if all[i] != int64(i) {
					t.Error("all[", i, "] is error, excepted is ", all[i], ", actual is", i)
				}
			}

			for i := 0; i <= c; i++ {
				if all[i] != cb.Get(i) {
					t.Error("all[", i, "] != cb.Get(", i, "), excepted is ", all[i], ", actual is", cb.Get(i))
				}
			}

			if int64(c) != cb.Last() {
				t.Error("excepted last is", c, ", actual is", cb.Last())
			}

			if all[0] != cb.First() {
				t.Error("excepted first is", all[0], ", actual is", cb.First())
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
				if all[i] != int64(c-9+i) {
					t.Error("all[", i, "] is error, excepted is", all[i], ", actual is", c-9+i)
				}
			}

			for i := 0; i < 10; i++ {
				if all[i] != cb.Get(i) {
					t.Error("all[", i, "] != cb.Get(", i, "), excepted is ", all[i], ", actual is", cb.Get(i))
				}
			}

			if int64(c) != cb.Last() {
				t.Error("excepted last is", c, ", actual is", cb.Last())
			}

			if int64(c-9) != cb.First() {
				t.Error("excepted first is", c-9, ", actual is", cb.First())
			}

			if all[0] != cb.First() {
				t.Error("excepted first is", all[0], ", actual is", cb.First())
			}

			if all[len(all)-1] != cb.Last() {
				t.Error("excepted first is", all[len(all)-1], ", actual is", cb.Last())
			}
		}
	}

	for i := 0; i < 100; i++ {
		cb.Push(int64(i))
		check(cb, i)
	}
}
