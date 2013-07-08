package poller

import (
	"testing"
)

var all_checker_tests = []struct {
	json   string
	value  interface{}
	e      string
	status int
}{{json: `{"attribute":"a", "operator":">", "value":"12"}`, value: map[string]interface{}{"a": 13}, status: 1},
	{json: `{"attribute":"a", "operator":">", "value":"12"}`, value: map[string]interface{}{"a": 12}, status: 0}}

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

		status, e := check.Run(test.value, map[string]interface{}{})
		if nil != e {
			if 0 == len(test.e) {
				t.Error("test all_checker_tests[%v] failed, %v", i, e)
			} else if test.e != e.Error() {
				t.Error("test all_checker_tests[%v] failed, excepted error is '%v', actual error is '%v'", i, test.e, e)
			}
			continue
		}

		if status != test.status {
			t.Error("test all_checker_tests[%v] failed, excepted is %v, actual is %v", i, test.status, status)
		}
	}
}
