package poller

import (
	"testing"
)

func TestSamplingBuffer(t *testing.T) {
	cb := newSamplingBuffer(make([]samplingResult, 10))

	check := func(cb *samplingBuffer, c int) {
		if c < 10 {
			if cb.Size() != (1 + c) {
				t.Error("size is error, excepted is", 1+c, ", actual is", cb.Size())
			}

			all := cb.All()
			if len(all) != (1 + c) {
				t.Error("len(all) is error, excepted is 10, actual is", cb.Size())
			}

			for i := 0; i <= c; i++ {
				if all[i].sampled_at != int64(i) {
					t.Error("all[", i, "] is error, excepted is ", all[i].sampled_at, ", actual is", i)
				}
			}

			for i := 0; i <= c; i++ {
				if all[i].sampled_at != cb.Get(i).sampled_at {
					t.Error("all[", i, "] != cb.Get(", i, "), excepted is ", all[i].sampled_at, ", actual is", cb.Get(i).sampled_at)
				}
			}

			if int64(c) != cb.Last().sampled_at {
				t.Error("excepted last is", c, ", actual is", cb.Last().sampled_at)
			}

			if all[0].sampled_at != cb.First().sampled_at {
				t.Error("excepted first is", all[0].sampled_at, ", actual is", cb.First().sampled_at)
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
				if all[i].sampled_at != int64(c-9+i) {
					t.Error("all[", i, "] is error, excepted is", all[i].sampled_at, ", actual is", c-9+i)
				}
			}

			for i := 0; i < 10; i++ {
				if all[i].sampled_at != cb.Get(i).sampled_at {
					t.Error("all[", i, "] != cb.Get(", i, "), excepted is ", all[i].sampled_at, ", actual is", cb.Get(i).sampled_at)
				}
			}

			if int64(c) != cb.Last().sampled_at {
				t.Error("excepted last is", c, ", actual is", cb.Last().sampled_at)
			}

			if int64(c-9) != cb.First().sampled_at {
				t.Error("excepted first is", c-9, ", actual is", cb.First().sampled_at)
			}

			if all[0].sampled_at != cb.First().sampled_at {
				t.Error("excepted first is", all[0].sampled_at, ", actual is", cb.First().sampled_at)
			}

			if all[len(all)-1].sampled_at != cb.Last().sampled_at {
				t.Error("excepted first is", all[len(all)-1].sampled_at, ", actual is", cb.Last().sampled_at)
			}
		}
	}

	for i := 0; i < 100; i++ {
		cb.Push(samplingResult{sampled_at: int64(i)})
		check(cb, i)
	}
}
