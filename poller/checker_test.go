package poller

import (
	"testing"
)

var all_checker_tests = []struct {
	json   string
	value  interface{}
	e      string
	status int
	v      interface{}
}{{json: `{"attribute":"a", "operator":">", "value":"12"}`, value: map[string]interface{}{"a": 13}, status: 1, v: int64(13)},
	{json: `{"attribute":"a", "operator":">", "value":"12"}`, value: map[string]interface{}{"a": 12}, status: 0, v: int64(12)}}

func TestCheckers(t *testing.T) {
	for i, test := range all_checker_tests {
		check, e := makeJsonChecker(test.json)
		if nil != e {
			if 0 == len(test.e) {
				t.Error("make all_checker_tests[%v] failed, %v", i, e)
			} else if test.e != e.Error() {
				t.Error("make all_checker_tests[%v] failed, excepted is '%v', actual is '%v'", i, test.e, e)
			}
			continue
		}

		status, v, e := check.Run(test.value, map[string]interface{}{})
		if nil != e {
			if 0 == len(test.e) {
				t.Errorf("test all_checker_tests[%v] failed, %v", i, e)
			} else if test.e != e.Error() {
				t.Errorf("test all_checker_tests[%v] failed, excepted error is '%v', actual error is '%v'", i, test.e, e)
			}
			continue
		}

		if v != test.v {
			t.Errorf("test all_checker_tests[%v] failed, excepted v is %v, actual v is %v", i, test.v, v)
		}

		if status != test.status {
			t.Errorf("test all_checker_tests[%v] failed, excepted status is %v, actual status is %v", i, test.status, status)
		}
	}
}
