package commons

import (
	"testing"
)

func TestCircularBuffer(t *testing.T) {
	cb := NewCircularBuffer(make([]interface{}, 10))

	check := func(cb *CircularBuffer, c int) {
		if c < 10 {
			if cb.Size() != (1 + c) {
				t.Error("size is error, excepted is", 1+c, ", actual is", cb.Size())
			}

			all := cb.All()
			if len(all) != (1 + c) {
				t.Error("len(all) is error, excepted is 10, actual is", cb.Size())
			}

			for i := 0; i <= c; i++ {
				if all[i] != i {
					t.Error("all[", i, "] is error, excepted is ", all[i], ", actual is", i)
				}
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
				if all[i] != c-9+i {
					t.Error("all[", i, "] is error, excepted is", all[i], ", actual is", c-9+i)
				}
			}
		}
	}

	for i := 0; i < 100; i++ {
		cb.Push(i)
		check(cb, i)
	}
}
