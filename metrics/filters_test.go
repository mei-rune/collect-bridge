package metrics

import (
	"strings"
	"testing"
)

var matchers_for_tests = []struct {
	method    string
	arguments []string
	params    string

	matchResult bool
	createError string
}{{"equal", []string{"", "aa"}, "aa", true, ""},
	{"equal", []string{"", "aa"}, "a", false, ""},
	{"equal", nil, "a", false, "wrong number of arguments (0 for 2)"},
	{"equal", []string{}, "a", false, "wrong number of arguments (0 for 2)"},
	{"equal", []string{"aa", "aa"}, "a", false, "wrong number of arguments (3 for 2)"},
	{"in", []string{"", "aa", "bb", "cc"}, "aa", true, ""},
	{"in", []string{"", "aa", "bb", "cc"}, "a", false, ""},
	{"in", []string{"", "aa", "bb", "cc"}, "cc", true, ""},
	{"in", nil, "a", false, "wrong number of arguments (0 for 2+)"},
	{"in", []string{}, "a", false, "wrong number of arguments (0 for 2+)"},
	{"in", []string{"", "a", "b", "c", "d"}, "a", true, "wrong number of arguments (0 for 2+)"},
	{"start_with", []string{"", "aaa"}, "aaaccc", true, ""},
	{"start_with", []string{"", "c"}, "aaaccc", false, ""},
	{"start_with", []string{"", "aaaa"}, "aaaccc", false, ""},
	{"start_with", nil, "a", false, "wrong number of arguments (0 for 2)"},
	{"start_with", []string{}, "a", false, "wrong number of arguments (0 for 2)"},
	{"start_with", []string{"", "aa", "aa"}, "a", false, "wrong number of arguments (3 for 2)"},
	{"end_with", []string{"", "ccc"}, "aaaccc", true, ""},
	{"end_with", []string{"", "c"}, "aaaccc", true, ""},
	{"end_with", []string{"", "cccc"}, "aaaccc", false, ""},
	{"end_with", nil, "a", false, "wrong number of arguments (0 for 2)"},
	{"end_with", []string{}, "a", false, "wrong number of arguments (0 for 2)"},
	{"end_with", []string{"", "aa", "aa"}, "a", false, "wrong number of arguments (3 for 2)"},
	{"contains", []string{"", "abc"}, "aaabccc", true, ""},
	{"contains", []string{"", "aab"}, "aaabccc", true, ""},
	{"contains", []string{"", "bccc"}, "aaabccc", true, ""},
	{"contains", []string{"", "aaab"}, "aaabccc", true, ""},
	{"contains", []string{"", "aaabb"}, "aaabccc", false, ""},
	{"contains", nil, "a", false, "wrong number of arguments (0 for 2)"},
	{"contains", []string{}, "a", false, "wrong number of arguments (0 for 2)"},
	{"contains", []string{"", "aa", "aa"}, "a", false, "wrong number of arguments (3 for 2)"},
	{"match", []string{"", "a{3}bc{3}"}, "aaabccc", true, ""},
	{"match", nil, "a", false, "wrong number of arguments (0 for 2)"},
	{"match", []string{}, "a", false, "wrong number of arguments (0 for 2)"},
	{"match", []string{"", "aa", "aa"}, "a", false, "wrong number of arguments (3 for 2)"},
	{"match2xxxxx", []string{"", "a{3}bc{3s}"}, "aaabccc", true, "Unsupported method - 'match2xxxxx'"}}

type matcher_defs_for_test struct {
	method    string
	arguments []string
}

var all_matchers_for_tests = []struct {
	matchers    []matcher_defs_for_test
	params      map[string]string
	matchResult bool
	matchError  string
	debuging    bool
}{ // test ok
	{[]matcher_defs_for_test{{"equal", []string{"a1", "aa"}},
		{"in", []string{"a2", "aa", "bb", "cc"}},
		{"start_with", []string{"a3", "aaa"}}},
		map[string]string{"a1": "aa", "a2": "cc", "a3": "aaaccc"}, true, "", false},
	{[]matcher_defs_for_test{{"equal", []string{"a1", "aa"}},
		{"in", []string{"a2", "aa", "bb", "cc"}},
		{"start_with", []string{"a3", "aaa"}}},
		map[string]string{"a1": "aa", "a2": "cc", "a3": "aaaccc"}, true, "", true},

	// test failed
	{[]matcher_defs_for_test{{"equal", []string{"a1", "aa"}},
		{"in", []string{"a2", "aa", "bb", "cc"}},
		{"start_with", []string{"a3", "aaa"}}},
		map[string]string{"a1": "a", "a2": "bc", "a3": "aaccc"}, false, "", false},
	// test failed with debug
	{[]matcher_defs_for_test{{"equal", []string{"a1", "aa"}},
		{"in", []string{"a2", "aa", "bb", "cc"}},
		{"start_with", []string{"a3", "aaa"}}},
		map[string]string{"a1": "a", "a2": "bc", "a3": "aaccc"}, false, "match failed:\n'a1' is not match - equal(a1,aa)!\n'a2' is not match - in(a2,aa,bb,cc)!\n'a3' is not match - start_with(a3,aaa)!", true}}

func TestMatch(t *testing.T) {
	for i, data := range matchers_for_tests {
		matcher, e := NewMatcher(data.method, data.arguments)
		if nil != e {
			if e.Error() != data.createError {
				t.Errorf("Create matcher %d failed, excepted is %v, actual is %v", i, data.createError, e.Error())
			}
			continue
		}

		ok := matcher.f(matcher, nil, data.params)
		if ok != data.matchResult {
			t.Errorf("match %d failed, excepted is %v, actual is %v -- %s, arguments is %s", i, data.matchResult, ok, matcher.Description, strings.Join(matcher.Arguments, ","))
		}
	}
}
func TestMatchs(t *testing.T) {
	for i, data := range all_matchers_for_tests {
		matchers := NewMatchers()
		for j, def := range data.matchers {
			matcher, e := NewMatcher(def.method, def.arguments)
			if nil != e {
				t.Errorf("Create matcher %d:%d failed, %v", i, j, e.Error())
				continue
			}
			matchers = append(matchers, matcher)
		}

		params := &context{params: data.params,
			managed_type: "unknow_type",
			managed_id:   "unknow_id",
			mo:           empty_mo,
			alias:        map[string]string{}}

		ok, e := matchers.Match(params, data.debuging)
		if ok != data.matchResult {
			t.Errorf("match '%d' failed, excepted is %v, actual is %v", i, data.matchResult, ok)
		}
		if nil == e {
			if "" != data.matchError {
				t.Errorf("match %d failed, excepted is %v, actual is nil -- params is %v", i, data.matchError, data.params)
			}
		} else if e.Error() != data.matchError {
			t.Errorf("match %d failed, excepted is %v, actual is %v -- params is %v", i, data.matchError, e, data.params)
		}
	}
}
