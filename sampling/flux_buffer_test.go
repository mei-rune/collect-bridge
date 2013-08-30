package sampling

import (
	"testing"
)

func TestFluxBuffer(t *testing.T) {
	cb := NewFluxBuffer(make([]Flux, 10))
	check := func(cb *fluxBuffer, c int, isNoCommit bool) {
		if c < 10 {
			all := cb.All()
			if isNoCommit {
				if cb.Size() != (c) {
					t.Error("size is error, excepted is", c, ", actual is", cb.Size())
				}

				if len(all) != (c) {
					t.Error("len(all) is error, excepted is ", c, ", actual is", cb.Size())
				}
			} else {
				if cb.Size() != (1 + c) {
					t.Error("size is error, excepted is", 1+c, ", actual is", cb.Size())
				}

				if len(all) != (1 + c) {
					t.Error("len(all) is error, excepted is ", 1+c, ", actual is", cb.Size())
				}
			}

			for i := 0; i < len(all); i++ {
				if all[i].IfIndex != i {
					t.Error("all[", i, "] is error, excepted is ", all[i].IfIndex, ", actual is", i)
				}
			}

			for i := 0; i < len(all); i++ {
				if all[i].IfIndex != cb.Get(i).IfIndex {
					t.Error("all[", i, "] != cb.Get(", i, "), excepted is ", all[i].IfIndex, ", actual is", cb.Get(i).IfIndex)
				}
			}

			if 0 != len(all) {

				if isNoCommit {
					if c-1 != cb.Last().IfIndex {
						t.Error("excepted last is", c-1, ", actual is", cb.Last().IfIndex)
					}
				} else {
					if c != cb.Last().IfIndex {
						t.Error("excepted last is", c, ", actual is", cb.Last().IfIndex)
					}
				}

				if all[0].IfIndex != cb.First().IfIndex {
					t.Error("excepted first is", all[0].IfIndex, ", actual is", cb.First().IfIndex)
				}

				if all[len(all)-1].IfIndex != cb.Last().IfIndex {
					t.Error("excepted first is", all[len(all)-1].IfIndex, ", actual is", cb.Last().IfIndex)
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
				if all[i].IfIndex != c-9+i {
					t.Error("all[", i, "] is error, excepted is", all[i].IfIndex, ", actual is", c-9+i)
				}
			}

			for i := 0; i < len(all); i++ {
				if all[i].IfIndex != cb.Get(i).IfIndex {
					t.Error("all[", i, "] != cb.Get(", i, "), excepted is ", all[i].IfIndex, ", actual is", cb.Get(i).IfIndex)
				}
			}

			if isNoCommit {
				if c-1 != cb.Last().IfIndex {
					t.Error("excepted last is", c-1, ", actual is", cb.Last().IfIndex)
				}
			} else {
				if c != cb.Last().IfIndex {
					t.Error("excepted last is", c, ", actual is", cb.Last().IfIndex)
				}
			}

			if c-9 != cb.First().IfIndex {
				t.Error("excepted first is", c-9, ", actual is", cb.First().IfIndex)
			}

			if all[0].IfIndex != cb.First().IfIndex {
				t.Error("excepted first is", all[0].IfIndex, ", actual is", cb.First().IfIndex)
			}

			if all[len(all)-1].IfIndex != cb.Last().IfIndex {
				t.Error("excepted first is", all[len(all)-1].IfIndex, ", actual is", cb.Last().IfIndex)
			}
		}
	}

	for i := 0; i < 100; i++ {
		cb.BeginPush()

		check(cb, i, true)
		flux := cb.BeginPush()
		flux.IfIndex = i
		cb.CommitPush()
		check(cb, i, false)
	}
}
