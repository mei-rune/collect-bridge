package carrier

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

func httpInvoke(action, url string, msg string, exceptedCode int) (string, error) {
	//fmt.Println(msg)
	req, err := http.NewRequest(action, url, bytes.NewBufferString(msg))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if nil != e {
		return "", e
	}

	// Install closing the request body (if any)
	defer func() {
		if nil != resp.Body {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != exceptedCode {
		resp_body, e := ioutil.ReadAll(resp.Body)
		if nil != e {
			return "", fmt.Errorf("%v: error", resp.StatusCode)
		}

		if nil == resp_body || 0 == len(resp_body) {
			return "", fmt.Errorf("%v: error", resp.StatusCode)
		}

		return "", fmt.Errorf("%v: %v", resp.StatusCode, string(resp_body))
	}

	resp_body, e := ioutil.ReadAll(resp.Body)
	if nil != e {
		return "", fmt.Errorf("%v: read body error", resp.StatusCode)
	}
	return string(resp_body), nil
}

func TestAlertsServer(t *testing.T) {
	SrvTest(t, func(db *sql.DB, url string) {
		nowt := time.Now()
		now, _ := nowt.MarshalJSON()

		_, e := httpInvoke("PUT", url+"/alerts", `[{"action_id": 123,
        "status": 1,
        "current_value": "23",
        "triggered_at": `+string(now)+`,
        "managed_type": "mo",
        "managed_id": 123}]`, 200)
		if nil != e {
			t.Error(e)
			return
		}
		entities, e := SelectAlertHistories(db)
		if nil != e {
			t.Error(e)
			return
		}

		if nil == entities || 1 != len(entities) {
			t.Error("nil == entities || 1 != len(entities)")
			return
		}

		AssertAlerts(t, entities[0], 123, 1, "23", nowt, "mo", 123)

		entities, e = SelectAlertCookies(db)
		if nil != e {
			t.Error(e)
			return
		}

		if nil == entities || 1 != len(entities) {
			t.Error("nil == entities || 1 != len(entities)")
			return
		}

		AssertAlerts(t, entities[0], 123, 1, "23", nowt, "mo", 123)

	})
}

func TestAlertsServer2(t *testing.T) {
	SrvTest(t, func(db *sql.DB, url string) {
		nowt, _ := time.Parse(time.RFC3339, "2013-07-13T14:13:28.7024412+08:00")

		_, e := httpInvoke("PUT", url+"/alerts", `[{"action_id":1,
"current_value":"",
"managed_id":1,
"managed_type":"managed_object",
"metric":"sys",
"name":"this is a test alert",
"status":1,
"trigger_id":"1",
"triggered_at":"2013-07-13T14:13:28.7024412+08:00"}]`, 200)
		if nil != e {
			t.Error(e)
			return
		}
		entities, e := SelectAlertHistories(db)
		if nil != e {
			t.Error(e)
			return
		}

		if nil == entities || 1 != len(entities) {
			t.Error("nil == entities || 1 != len(entities)")
			return
		}

		AssertAlerts(t, entities[0], 1, 1, "", nowt, "managed_object", 1)

		entities, e = SelectAlertCookies(db)
		if nil != e {
			t.Error(e)
			return
		}

		if nil == entities || 1 != len(entities) {
			t.Error("nil == entities || 1 != len(entities)")
			return
		}

		AssertAlerts(t, entities[0], 1, 1, "", nowt, "managed_object", 1)

	})
}

func TestAlertsCookies(t *testing.T) {
	SrvTest(t, func(db *sql.DB, url string) {
		nowt := time.Now()
		now, _ := nowt.MarshalJSON()

		_, e := httpInvoke("PUT", url+"/alerts", `[{"action_id": 123,
        "status": 1,
        "current_value": "23",
        "triggered_at": `+string(now)+`,
        "managed_type": "mo",
        "managed_id": 123}]`, 200)
		if nil != e {
			t.Error(e)
			return
		}
		entities, e := SelectAlertHistories(db)
		if nil != e {
			t.Error(e)
			return
		}

		if nil == entities || 1 != len(entities) {
			t.Error("nil == entities || 1 != len(entities)")
			return
		}

		AssertAlerts(t, entities[0], 123, 1, "23", nowt, "mo", 123)

		entities, e = SelectAlertCookies(db)
		if nil != e {
			t.Error(e)
			return
		}

		if nil == entities || 1 != len(entities) {
			t.Error("nil == entities || 1 != len(entities)")
			return
		}

		AssertAlerts(t, entities[0], 123, 1, "23", nowt, "mo", 123)

		_, e = httpInvoke("PUT", url+"/alerts", `[{"action_id": 123,
        "status": 0,
        "current_value": "65",
        "triggered_at": `+string(now)+`,
        "managed_type": "mo",
        "managed_id": 123}]`, 200)
		if nil != e {
			t.Error(e)
			return
		}
		entities, e = SelectAlertHistories(db)
		if nil != e {
			t.Error(e)
			return
		}

		if nil == entities || 2 != len(entities) {
			t.Error("nil == entities || 2 != len(entities)")
			return
		}

		AssertAlerts(t, entities[0], 123, 1, "23", nowt, "mo", 123)
		AssertAlerts(t, entities[1], 123, 0, "65", nowt, "mo", 123)

		entities, e = SelectAlertCookies(db)
		if nil != e {
			t.Error(e)
			return
		}

		if nil == entities || 0 == len(entities) {
			return
		}

		t.Error("nil != entities || 0 != len(entities)")
	})
}

func TestHistoriesServer(t *testing.T) {
	SrvTest(t, func(db *sql.DB, url string) {
		nowt := time.Now()
		now, _ := nowt.MarshalJSON()

		_, e := httpInvoke("PUT", url+"/histories", `[{"action_id": 123,
        "value": 23,
        "sampled_at": `+string(now)+`,
        "managed_type": "mo",
        "managed_id": 123}]`, 200)
		if nil != e {
			t.Error(e)
			return
		}

		_, e = httpInvoke("PUT", url+"/histories", `[{"action_id": 1243,
        "value": 233,
        "sampled_at": `+string(now)+`,
        "managed_type": "mo2",
        "managed_id": 124}]`, 200)
		if nil != e {
			t.Error(e)
			return
		}

		entities, e := SelectHistories(db)
		if nil != e {
			t.Error(e)
			return
		}

		if nil == entities || 2 != len(entities) {
			t.Error("nil == entities || 2 != len(entities)")
			return
		}

		AssertHistories(t, entities[0], 123, 23, nowt, "mo", 123)
		AssertHistories(t, entities[1], 1243, 233, nowt, "mo2", 124)
	})
}

func countResult(s string) (int, error) {
	var res map[string]interface{} = nil
	e := json.Unmarshal([]byte(s), &res)
	if nil != e {
		return 0, e
	}
	entities := res["value"]
	if values, ok := entities.([]interface{}); ok {
		return len(values), nil
	}
	return 0, fmt.Errorf("it is not a result - %v", s)
}

func TestFind(t *testing.T) {
	SrvTest(t, func(db *sql.DB, url string) {
		for _, test := range []struct{ tname, uname string }{{tname: "tpt_alert_cookies", uname: "alert_cookies"},
			{tname: "tpt_alert_histories", uname: "alerts"},
			{tname: "tpt_histories", uname: "histories"}} {

			for i := 0; i < 10; i++ {
				var e error
				if strings.HasPrefix(test.tname, "tpt_alert") {
					_, e = db.Exec(fmt.Sprintf("insert into %v(action_id, managed_type, managed_id, status, current_value, triggered_at) values(%v, 'mo', 1%v, 1, 'ss', now())", test.tname, i, i%2))
				} else {
					_, e = db.Exec(fmt.Sprintf("insert into %v(action_id, managed_type, managed_id, current_value, sampled_at) values(%v, 'mo', 1%v, 1, now())", test.tname, i, i%2))
				}
				if nil != e {
					t.Error(e)
					return
				}
			}

			s, e := httpInvoke("GET", url+"/"+test.uname, "", 200)
			if nil != e {
				t.Error(e)
				return
			}

			count, e := countResult(s)
			if nil != e {
				t.Error(e)
				return
			}

			if 10 != count {
				t.Error(`excepted is 10`)
				t.Error(`actual is`, count, "--", s)
			}

			s, e = httpInvoke("GET", url+"/"+test.uname+"?@action_id=1", "", 200)
			if nil != e {
				t.Error(e)
				return
			}

			count, e = countResult(s)
			if nil != e {
				t.Error(e)
				return
			}

			if 1 != count {
				t.Error(`excepted is 1`)
				t.Error(`actual is`, count, "--", s)
			}
		}
	})
}

func TestCount(t *testing.T) {
	SrvTest(t, func(db *sql.DB, url string) {
		for _, test := range []struct{ tname, uname string }{{tname: "tpt_alert_cookies", uname: "alert_cookies"},
			{tname: "tpt_alert_histories", uname: "alerts"},
			{tname: "tpt_histories", uname: "histories"}} {

			for i := 0; i < 10; i++ {
				var e error
				if strings.HasPrefix(test.tname, "tpt_alert") {
					_, e = db.Exec(fmt.Sprintf("insert into %v(action_id, managed_type, managed_id, status, current_value, triggered_at) values(%v, 'mo', 1%v, 1, 'ss', now())", test.tname, i, i%2))
				} else {
					_, e = db.Exec(fmt.Sprintf("insert into %v(action_id, managed_type, managed_id, current_value, sampled_at) values(%v, 'mo', 1%v, 1, now())", test.tname, i, i%2))
				}
				if nil != e {
					t.Error(e)
					return
				}
			}

			s, e := httpInvoke("GET", url+"/"+test.uname+"/count", "", 200)
			if nil != e {
				t.Error(e)
				return
			}

			if !strings.Contains(s, `"value":10`) {
				t.Error(`excepted contains "value":10`)
				t.Error(`actual is`, s)
			}

			s, e = httpInvoke("GET", url+"/"+test.uname+"/count?@action_id=1", "", 200)
			if nil != e {
				t.Error(e)
				return
			}

			if !strings.Contains(s, `"value":1`) {
				t.Error(`excepted contains "value":1`)
				t.Error(`actual is`, s)
			}

			s, e = httpInvoke("GET", url+"/"+test.uname+"/count?@managed_id=11", "", 200)
			if nil != e {
				t.Error(e)
				return
			}

			if !strings.Contains(s, `"value":5`) {
				t.Error(`excepted contains "value":5`)
				t.Error(`actual is`, s)
			}
		}
	})
}

func TestRemove(t *testing.T) {
	SrvTest(t, func(db *sql.DB, url string) {
		for _, test := range []struct{ tname, uname string }{{tname: "tpt_alert_cookies", uname: "alert_cookies"},
			{tname: "tpt_alert_histories", uname: "alerts"},
			{tname: "tpt_histories", uname: "histories"}} {

			for i := 0; i < 10; i++ {
				var e error
				if strings.HasPrefix(test.tname, "tpt_alert") {
					_, e = db.Exec(fmt.Sprintf("insert into %v(action_id, managed_type, managed_id, status, current_value, triggered_at) values(%v, 'mo', 1%v, 1, 'ss', now())", test.tname, i, i%2))
				} else {
					_, e = db.Exec(fmt.Sprintf("insert into %v(action_id, managed_type, managed_id, current_value, sampled_at) values(%v, 'mo', 1%v, 1, now())", test.tname, i, i%2))
				}
				if nil != e {
					t.Error(e)
					return
				}
			}

			count := int64(0)
			e := db.QueryRow("select count(*) from " + test.tname).Scan(&count)
			if nil != e {
				t.Error(e)
				return
			}

			if count != 10 {
				t.Error("count is 10 before delete it, actual is ", count)
			}

			s, e := httpInvoke("DELETE", url+"/"+test.uname+"?@action_id=1", "", 200)
			if nil != e {
				t.Error(e)
				return
			}

			if !strings.Contains(s, `"effected":1`) {
				t.Error(`excepted contains "value":1`)
				t.Error(`actual is`, s)
			}

			count = int64(0)
			e = db.QueryRow("select count(*) from " + test.tname + " where action_id=1").Scan(&count)
			if nil != e {
				t.Error(e)
				return
			}

			if count != 0 {
				t.Error("count with action_id = 1 is not 0 after delete it, actual is ", count)
			}

			count = int64(0)
			e = db.QueryRow("select count(*) from " + test.tname).Scan(&count)
			if nil != e {
				t.Error(e)
				return
			}

			if count != 9 {
				t.Error("count is not 9 after delete it, actual is ", count)
			}

			s, e = httpInvoke("DELETE", url+"/"+test.uname+"?@managed_id=11", "", 200)
			if nil != e {
				t.Error(e)
				return
			}

			if !strings.Contains(s, `"effected":4`) {
				t.Error(`excepted contains "effected":4`)
				t.Error(`actual is`, s)
			}

			count = int64(0)
			e = db.QueryRow("select count(*) from " + test.tname + " where managed_id=11").Scan(&count)
			if nil != e {
				t.Error(e)
				return
			}

			if count != 0 {
				t.Error("count with managed_id = 11 is not 0 after delete it, actual is ", count)
			}

			count = int64(0)
			e = db.QueryRow("select count(*) from " + test.tname).Scan(&count)
			if nil != e {
				t.Error(e)
				return
			}

			if count != 5 {
				t.Error("count is not 4 after delete it, actual is ", count)
			}

			s, e = httpInvoke("DELETE", url+"/"+test.uname, "", 200)
			if nil != e {
				t.Error(e)
				return
			}

			if !strings.Contains(s, `"effected":5`) {
				t.Error(`excepted contains "value":5`)
				t.Error(`actual is`, s)
			}

			count = int64(0)
			e = db.QueryRow("select count(*) from " + test.tname).Scan(&count)
			if nil != e {
				t.Error(e)
				return
			}

			if count != 0 {
				t.Error("count is not 0 after delete it, actual is ", count)
			}
		}
	})
}
