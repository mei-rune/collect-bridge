package main

import (
	"flag"
	"license"
	"net/http/httptest"
	"strings"
	"testing"
)

func simpleTest(t *testing.T, auth map[string]interface{}, cb func()) {
	srv := httptest.NewServer(&License{meta: auth})
	defer srv.Close()
	t.Log(srv.URL)
	flag.Set("license_srv", srv.URL+"/")
	cb()
}

func TestMatchNode(t *testing.T) {
	simpleTest(t, map[string]interface{}{"node": 12}, func() {
		if _, e := license.IsMatchNode("a", 13); nil != e {
			if !strings.Contains(e.Error(), "NONEXCEPTED") {
				t.Error(e)
			}
		} else {
			t.Error("no error")
		}
		if _, e := license.IsMatchNode("a", 11); nil != e {
			t.Error(e)
		}
	})

	simpleTest(t, map[string]interface{}{"node": -1}, func() {
		if _, e := license.IsMatchNode("a", 13); nil != e {
			t.Error(e)
		}
		if _, e := license.IsMatchNode("a", 11); nil != e {
			t.Error(e)
		}
		if _, e := license.IsMatchNode("a", 199999999991); nil != e {
			t.Error(e)
		}
	})

	simpleTest(t, map[string]interface{}{}, func() {
		if _, e := license.IsMatchNode("a", 13); nil != e {
			if !strings.Contains(e.Error(), "401:") {
				t.Error(e)
			}
		} else {
			t.Error("no error")
		}
		if _, e := license.IsMatchNode("a", 11); nil != e {
			if !strings.Contains(e.Error(), "401:") {
				t.Error(e)
			}
		} else {
			t.Error("no error")
		}
		if _, e := license.IsMatchNode("a", 0); nil != e {
			if !strings.Contains(e.Error(), "401:") {
				t.Error(e)
			}
		} else {
			t.Error("no error")
		}
	})
	simpleTest(t, map[string]interface{}{}, func() {
		if _, e := license.IsMatchNode("a", 0); nil != e {
			if !strings.Contains(e.Error(), "401:") {
				t.Error(e)
			}
		}
	})

	simpleTest(t, map[string]interface{}{"node": "asdf"}, func() {
		if _, e := license.IsMatchNode("a", 13); nil != e {
			if !strings.Contains(e.Error(), "401:") {
				t.Error(e)
			}
		} else {
			t.Error("no error")
		}
		if _, e := license.IsMatchNode("a", 0); nil != e {
			if !strings.Contains(e.Error(), "401:") {
				t.Error(e)
			}
		} else {
			t.Error("no error")
		}
		if _, e := license.IsMatchNode("a", 199999999991); nil != e {
			if !strings.Contains(e.Error(), "401:") {
				t.Error(e)
			}
		} else {
			t.Error("no error")
		}
	})
}

func TestIsEnabledModule(t *testing.T) {
	simpleTest(t, map[string]interface{}{"node": 12,
		"module_a":   "enabled",
		"module_b":   "disabled",
		"module_all": "enabled"}, func() {
		if _, e := license.IsEnabledModule("a"); nil != e {
			t.Error(e)
		}
		if b, e := license.IsEnabledModule("b"); nil != e || !b {
			if !strings.Contains(e.Error(), "401:") {
				t.Error(e)
			}
		} else {
			t.Error("no error")
		}
		if _, e := license.IsEnabledModule("c"); nil != e {
			t.Error(e)
		}
	})

	simpleTest(t, map[string]interface{}{"node": 12,
		"module_a":   "enabled",
		"module_b":   "disabled",
		"module_all": "enabled"}, func() {
		if _, e := license.IsEnabledModule("a"); nil != e {
			t.Error(e)
		}
		if _, e := license.IsEnabledModule("b"); nil != e {
			if !strings.Contains(e.Error(), "401:") {
				t.Error(e)
			}
		} else {
			t.Error("no error")
		}
		if _, e := license.IsEnabledModule("c"); nil != e {
			t.Error(e)
		}
	})

	simpleTest(t, map[string]interface{}{"node": 12,
		"module_a": "enabled",
		"module_b": "disabled"}, func() {
		if _, e := license.IsEnabledModule("a"); nil != e {
			t.Error(e)
		}
		if _, e := license.IsEnabledModule("b"); nil != e {
			if !strings.Contains(e.Error(), "401:") {
				t.Error(e)
			}
		} else {
			t.Error("no error")
		}
		if _, e := license.IsEnabledModule("c"); nil != e {
			if !strings.Contains(e.Error(), "401:") {
				t.Error(e)
			}
		} else {
			t.Error("no error")
		}
	})
}

func TestSetInvalid(t *testing.T) {
	simpleTest(t, map[string]interface{}{"node": 12,
		"module_a":   "enabled",
		"module_b":   "disabled",
		"module_all": "enabled"}, func() {
		license.SetInvalidLicense()
		if _, e := license.IsEnabledModule("b"); nil != e {
			if !strings.Contains(e.Error(), "401:") {
				t.Error(e)
			}
		} else {
			t.Error("no error")
		}

		if _, e := license.IsEnabledModule("c"); nil != e {
			if !strings.Contains(e.Error(), "401:") {
				t.Error(e)
			}
		} else {
			t.Error("no error")
		}

		if _, e := license.IsMatchNode("a", 3); nil != e {
			if !strings.Contains(e.Error(), "401:") {
				t.Error(e)
			}
		} else {
			t.Error("no error")
		}

	})
}
