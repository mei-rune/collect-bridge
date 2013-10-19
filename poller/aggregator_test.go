package poller

import (
	"testing"
	"time"
)

func TestLastMerger(t *testing.T) {
	var merger aggregator = &last_merger{interval: 4 * time.Minute}
	for idx, test := range []struct {
		t        string
		value    interface{}
		ok       bool
		excepted interface{}
	}{{t: "2013-10-17T15:04:05+00:00", value: 12, ok: true, excepted: 12},
		{t: "2013-10-17T15:04:35+00:00", value: 12, ok: false, excepted: 12},
		{t: "2013-10-17T15:05:06+00:00", value: 12, ok: false, excepted: 12},
		{t: "2013-10-17T15:06:07+00:00", value: 12, ok: false, excepted: 12},
		{t: "2013-10-17T15:07:08+00:00", value: 12, ok: false, excepted: 12},
		{t: "2013-10-17T15:08:09+00:00", value: 12, ok: true, excepted: 12},
		{t: "2013-10-17T15:09:09+00:00", value: 13, ok: false, excepted: 13},
		{t: "2013-10-17T15:10:10+00:00", value: 13, ok: false, excepted: 13},
		{t: "2013-10-17T15:11:11+00:00", value: 13, ok: false, excepted: 13},
		{t: "2013-10-17T15:12:12+00:00", value: 13, ok: true, excepted: 13},
		{t: "2013-10-17T15:13:13+00:00", value: 13, ok: false, excepted: 13},
		{t: "2013-10-17T15:10:14+00:00", value: 14, ok: true, excepted: 14}} {
		tt, e := time.Parse(time.RFC3339, test.t)
		if nil != e {
			t.Error(e)
			continue
		}
		value, ok, e := merger.aggregate(test.value, tt)
		if ok != test.ok {
			t.Errorf("[%d - %v]ok != ok", idx, test.t)
			continue
		}
		if !ok {
			continue
		}

		if nil != e {
			t.Error(e)
			continue
		}

		if test.excepted != value {
			t.Errorf("[%d - %v]value != value, %v, %v", idx, test.t, value)
		}
	}
}
