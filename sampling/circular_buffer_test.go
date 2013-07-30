package sampling

import (
	"testing"
)

func TestCircularBuffer(t *testing.T) {
	cb := newCircularBuffer(make([]interface{}, 11))

	check := func(cb *circularBuffer, c int) {
		if c < 10 {
			if cb.size() != (1 + c) {
				t.Error("size is error, excepted is", 1+c, ", actual is", cb.size())
			}

			all := cb.all()
			if len(all) != (1 + c) {
				t.Error("len(all) is error, excepted is 10, actual is", cb.size())
			}

			for i := 0; i <= c; i++ {
				if all[i] != i {
					t.Error("all[", i, "] is error, excepted is ", all[i], ", actual is", i)
				}
			}

		} else {
			if cb.size() != 10 {
				t.Error("size is error, excepted is 10, actual is", cb.size())
			}

			all := cb.all()
			if len(all) != 10 {
				t.Error("len(all) is error, excepted is 10, actual is", cb.size())
			}

			for i := 0; i < 10; i++ {
				if all[i] != c-9+i {
					t.Error("all[", i, "] is error, excepted is", all[i], ", actual is", c-9+i)
				}
			}
		}
	}

	for i := 0; i < 100; i++ {
		cb.push(i)
		check(cb, i)
	}
}
