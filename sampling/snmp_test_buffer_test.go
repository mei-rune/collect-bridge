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
				if all[i].SampledAt != int64(i) {
					t.Error("all[", i, "] is error, excepted is ", all[i].SampledAt, ", actual is", i)
				}
			}

			for i := 0; i < len(all); i++ {
				if all[i].SampledAt != cb.Get(i).SampledAt {
					t.Error("all[", i, "] != cb.Get(", i, "), excepted is ", all[i].SampledAt, ", actual is", cb.Get(i).SampledAt)
				}
			}

			if 0 != len(all) {

				if isNoCommit {
					if c-1 != cb.Last().SampledAt {
						t.Error("excepted last is", c-1, ", actual is", cb.Last().SampledAt)
					}
				} else {
					if c != cb.Last().SampledAt {
						t.Error("excepted last is", c, ", actual is", cb.Last().SampledAt)
					}
				}

				if all[0].SampledAt != cb.First().SampledAt {
					t.Error("excepted first is", all[0].SampledAt, ", actual is", cb.First().SampledAt)
				}

				if all[len(all)-1].SampledAt != cb.Last().SampledAt {
					t.Error("excepted first is", all[len(all)-1].SampledAt, ", actual is", cb.Last().SampledAt)
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
				if all[i].SampledAt != c-9+int64(i) {
					t.Error("all[", i, "] is error, excepted is", all[i].SampledAt, ", actual is", c-9+int64(i))
				}
			}

			for i := 0; i < len(all); i++ {
				if all[i].SampledAt != cb.Get(i).SampledAt {
					t.Error("all[", i, "] != cb.Get(", i, "), excepted is ", all[i].SampledAt, ", actual is", cb.Get(i).SampledAt)
				}
			}

			if isNoCommit {
				if c-1 != cb.Last().SampledAt {
					t.Error("excepted last is", c-1, ", actual is", cb.Last().SampledAt)
				}
			} else {
				if c != cb.Last().SampledAt {
					t.Error("excepted last is", c, ", actual is", cb.Last().SampledAt)
				}
			}

			if c-9 != cb.First().SampledAt {
				t.Error("excepted first is", c-9, ", actual is", cb.First().SampledAt)
			}

			if all[0].SampledAt != cb.First().SampledAt {
				t.Error("excepted first is", all[0].SampledAt, ", actual is", cb.First().SampledAt)
			}

			if all[len(all)-1].SampledAt != cb.Last().SampledAt {
				t.Error("excepted first is", all[len(all)-1].SampledAt, ", actual is", cb.Last().SampledAt)
			}
		}
	}

	for i := 0; i < 100; i++ {
		cb.BeginPush()

		check(cb, int64(i), true)
		flux := cb.BeginPush()
		flux.SampledAt = int64(i)
		cb.CommitPush()
		check(cb, int64(i), false)
	}
}
