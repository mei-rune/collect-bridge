package sampling

import (
	"testing"
)

func TestSnmpTestResultBuffer(t *testing.T) {
	cb := NewSnmpTestResultBuffer(make([]SnmpTestResult, 10))
	check := func(cb *snmpTestResultBuffer, c int64, isNoCommit bool) {
		if c < 10 {
			all := cb.All()
			if isNoCommit {
				if cb.Size() != int(c) {
					t.Error("size is error, excepted is", c, ", actual is", cb.Size())
				}

				if len(all) != int(c) {
					t.Error("len(all) is error, excepted is ", c, ", actual is", cb.Size())
				}
			} else {
				if cb.Size() != int(1+c) {
					t.Error("size is error, excepted is", 1+c, ", actual is", cb.Size())
				}

				if len(all) != int(1+c) {
					t.Error("len(all) is error, excepted is ", 1+c, ", actual is", cb.Size())
				}
			}

			for i := 0; i < len(all); i++ {
				if all[i].SendAt != int64(i) {
					t.Error("all[", i, "] is error, excepted is ", all[i].SendAt, ", actual is", i)
				}
			}

			for i := 0; i < len(all); i++ {
				if all[i].SendAt != cb.Get(i).SendAt {
					t.Error("all[", i, "] != cb.Get(", i, "), excepted is ", all[i].SendAt, ", actual is", cb.Get(i).SendAt)
				}
			}

			if 0 != len(all) {

				if isNoCommit {
					if c-1 != cb.Last().SendAt {
						t.Error("excepted last is", c-1, ", actual is", cb.Last().SendAt)
					}
				} else {
					if c != cb.Last().SendAt {
						t.Error("excepted last is", c, ", actual is", cb.Last().SendAt)
					}
				}

				if all[0].SendAt != cb.First().SendAt {
					t.Error("excepted first is", all[0].SendAt, ", actual is", cb.First().SendAt)
				}

				if all[len(all)-1].SendAt != cb.Last().SendAt {
					t.Error("excepted first is", all[len(all)-1].SendAt, ", actual is", cb.Last().SendAt)
				}
			}
		} else {
			all := cb.All()
			if isNoCommit {
				if cb.Size() != 9 {
					t.Error("size is error, excepted is 9, actual is", cb.Size())
				}

				if len(all) != 9 {
					t.Error("len(all) is error, excepted is 9, actual is", cb.Size())
				}
			} else {
				if cb.Size() != 10 {
					t.Error("size is error, excepted is 10, actual is", cb.Size())
				}
				if len(all) != 10 {
					t.Error("len(all) is error, excepted is 10, actual is", cb.Size())
				}
			}

			for i := 0; i < len(all); i++ {
				if all[i].SendAt != c-9+int64(i) {
					t.Error("all[", i, "] is error, excepted is", all[i].SendAt, ", actual is", c-9+int64(i))
				}
			}

			for i := 0; i < len(all); i++ {
				if all[i].SendAt != cb.Get(i).SendAt {
					t.Error("all[", i, "] != cb.Get(", i, "), excepted is ", all[i].SendAt, ", actual is", cb.Get(i).SendAt)
				}
			}

			if isNoCommit {
				if c-1 != cb.Last().SendAt {
					t.Error("excepted last is", c-1, ", actual is", cb.Last().SendAt)
				}
			} else {
				if c != cb.Last().SendAt {
					t.Error("excepted last is", c, ", actual is", cb.Last().SendAt)
				}
			}

			if c-9 != cb.First().SendAt {
				t.Error("excepted first is", c-9, ", actual is", cb.First().SendAt)
			}

			if all[0].SendAt != cb.First().SendAt {
				t.Error("excepted first is", all[0].SendAt, ", actual is", cb.First().SendAt)
			}

			if all[len(all)-1].SendAt != cb.Last().SendAt {
				t.Error("excepted first is", all[len(all)-1].SendAt, ", actual is", cb.Last().SendAt)
			}
		}
	}

	for i := 0; i < 100; i++ {
		cb.Push(SnmpTestResult{SendAt: int64(i)})

		check(cb, int64(i), false)
	}
}
