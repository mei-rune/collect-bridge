package carrier

import (
	"bytes"
	"commons"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func httpInvoke(action, url string, msg string, exceptedCode int) (string, error) {
	//fmt.Println(msg)
	req, err := http.NewRequest(action, url, bytes.NewBufferString(msg))
	if err != nil {
		return "", commons.NewApplicationError(500, err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if nil != e {
		return "", commons.NewApplicationError(500, e.Error())
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
			return "", commons.NewApplicationError(resp.StatusCode, fmt.Sprintf("%v: error", resp.StatusCode))
		}

		if nil == resp_body || 0 == len(resp_body) {
			return "", commons.NewApplicationError(resp.StatusCode, fmt.Sprintf("%v: error", resp.StatusCode))
		}

		return "", commons.NewApplicationError(resp.StatusCode, fmt.Sprintf("%v: %v", resp.StatusCode, string(resp_body)))
	}

	resp_body, e := ioutil.ReadAll(resp.Body)
	if nil != e {
		return "", commons.NewApplicationError(resp.StatusCode, fmt.Sprintf("%v: read body error", resp.StatusCode))
	}
	return string(resp_body), nil
}

func TestAlertsServer(t *testing.T) {
	SrvTest(t, func(db *sql.DB, url string) {
		nowt := time.Now()
		now, _ := nowt.MarshalJSON()

		_, e := httpInvoke("PUT", url+"/alerts", `[{"action_id": 123,
        "status": 1,
        "previous_status": 0,
        "event_id": "123",
        "sequence_id": 1,
        "content": "content is alerted",
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

		AssertAlerts(t, entities[0], 123, 1, 0, "123", 1, "content is alerted", "23", nowt, "mo", 123)

		entities, e = SelectAlertCookies(db)
		if nil != e {
			t.Error(e)
			return
		}

		if nil == entities || 1 != len(entities) {
			t.Error("nil == entities || 1 != len(entities)")
			return
		}

		AssertAlerts(t, entities[0], 123, 1, 0, "123", 1, "content is alerted", "23", nowt, "mo", 123)

		_, e = httpInvoke("PUT", url+"/alerts", `[{"action_id": 123,
        "status": 2,
        "previous_status": 0,
        "event_id": "123",
        "sequence_id": 1,
        "content": "content is alerted",
        "current_value": "23",
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

		AssertAlerts(t, entities[0], 123, 1, 0, "123", 1, "content is alerted", "23", nowt, "mo", 123)
		AssertAlerts(t, entities[1], 123, 2, 0, "123", 1, "content is alerted", "23", nowt, "mo", 123)

		entities, e = SelectAlertCookies(db)
		if nil != e {
			t.Error(e)
			return
		}

		if nil == entities || 1 != len(entities) {
			t.Error("nil == entities || 1 != len(entities)")
			return
		}

		AssertAlerts(t, entities[0], 123, 2, 0, "123", 1, "content is alerted", "23", nowt, "mo", 123)

	})
}

func TestAlertsServer2(t *testing.T) {
	SrvTest(t, func(db *sql.DB, url string) {
		nowt := time.Now()
		now, _ := nowt.MarshalJSON()

		_, e := httpInvoke("PUT", url+"/alerts", `[{"action_id":1,
      "content": "content is alerted",
			"current_value":"",
			"managed_id":1,
			"managed_type":"managed_object",
			"metric":"sys",
			"name":"this is a test alert",
			"status":1,
      "previous_status": 0,
      "event_id": "123",
      "sequence_id": 1,
			"trigger_id":"1",
			"triggered_at":`+string(now)+`}]`, 200)
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

		AssertAlerts(t, entities[0], 1, 1, 0, "123", 1, "content is alerted", "", nowt, "managed_object", 1)

		entities, e = SelectAlertCookies(db)
		if nil != e {
			t.Error(e)
			return
		}

		if nil == entities || 1 != len(entities) {
			t.Error("nil == entities || 1 != len(entities)")
			return
		}

		AssertAlerts(t, entities[0], 1, 1, 0, "123", 1, "content is alerted", "", nowt, "managed_object", 1)

	})
}

func TestAlertsCookies(t *testing.T) {
	SrvTest(t, func(db *sql.DB, url string) {
		nowt := time.Now()
		now, _ := nowt.MarshalJSON()

		_, e := httpInvoke("PUT", url+"/alerts", `[{"action_id": 123,
        "status": 1,
        "previous_status": 0,
        "event_id": "123",
        "sequence_id": 1,
        "content": "content is alerted",
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

		AssertAlerts(t, entities[0], 123, 1, 0, "123", 1, "content is alerted", "23", nowt, "mo", 123)

		entities, e = SelectAlertCookies(db)
		if nil != e {
			t.Error(e)
			return
		}

		if nil == entities || 1 != len(entities) {
			t.Error("nil == entities || 1 != len(entities)")
			return
		}

		AssertAlerts(t, entities[0], 123, 1, 0, "123", 1, "content is alerted", "23", nowt, "mo", 123)

		_, e = httpInvoke("PUT", url+"/alerts", `[{"action_id": 123,
        "status": 0,
        "previous_status": 1,
        "event_id": "123",
        "sequence_id": 1,
        "content": "content is alerted",
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

		AssertAlerts(t, entities[0], 123, 1, 0, "123", 1, "content is alerted", "23", nowt, "mo", 123)
		AssertAlerts(t, entities[1], 123, 0, 1, "123", 1, "content is alerted", "65", nowt, "mo", 123)

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

func getId(s string) (int, error) {
	var res map[string]interface{} = nil
	e := json.Unmarshal([]byte(s), &res)
	if nil != e {
		return 0, e
	}
	entities := res["value"]
	if values, ok := entities.([]interface{}); ok && 1 <= len(values) {
		if m, ok := values[0].(map[string]interface{}); ok {
			return commons.GetInt(m, "id")
		}
		return len(values), nil
	}
	return 0, fmt.Errorf("it is not a result - %v", s)
}

func TestFind(t *testing.T) {
	SrvTest(t, func(db *sql.DB, url string) {

		now := time.Now()
		nows_str := "_" + strconv.Itoa(now.Year()) + months[now.Month()]
		if "postgres" != *db_drv {
			nows_str = ""
		}

		for _, test := range []struct{ tname, uname string }{{tname: "tpt_alert_cookies", uname: "alert_cookies"},
			{tname: "tpt_alert_histories" + nows_str, uname: "alerts"},
			{tname: "tpt_histories" + nows_str, uname: "histories"}} {

			now_func := "now()"
			if "odbc_with_mssql" == *db_drv {
				now_func = "GETDATE()"
			}

			for i := 0; i < 10; i++ {
				var e error
				if strings.HasPrefix(test.tname, "tpt_alert") {
					_, e = db.Exec(fmt.Sprintf("insert into %v(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) values(%v, 'mo', 1%v, 1, 0, '123', 1, 'content is alerted', 'ss', "+now_func+")", test.tname, i, i%2))
				} else {
					_, e = db.Exec(fmt.Sprintf("insert into %v(action_id, managed_type, managed_id, current_value, sampled_at) values(%v, 'mo', 1%v, 1, "+now_func+")", test.tname, i, i%2))
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

func TestFindById(t *testing.T) {
	SrvTest(t, func(db *sql.DB, url string) {

		now := time.Now()
		nows_str := "_" + strconv.Itoa(now.Year()) + months[now.Month()]
		if "postgres" != *db_drv {
			nows_str = ""
		}

		for _, test := range []struct{ tname, uname string }{{tname: "tpt_alert_cookies", uname: "alert_cookies"},
			{tname: "tpt_alert_histories" + nows_str, uname: "alerts"},
			{tname: "tpt_histories" + nows_str, uname: "histories"}} {

			now_func := "now()"
			if "odbc_with_mssql" == *db_drv {
				now_func = "GETDATE()"
			}

			var e error
			if strings.HasPrefix(test.tname, "tpt_alert") {
				_, e = db.Exec(fmt.Sprintf("insert into %v(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) values(%v, 'msssssssso', 1%v, 1, 0, '123', 1, 'content is alerted', 'ss', "+now_func+")", test.tname, 2, 2%2))
			} else {
				_, e = db.Exec(fmt.Sprintf("insert into %v(action_id, managed_type, managed_id, current_value, sampled_at) values(%v, 'msssssssso', 1%v, 1, "+now_func+")", test.tname, 2, 2%2))
			}
			if nil != e {
				t.Error("test["+test.tname+"] failed", e)
				return
			}

			s, e := httpInvoke("GET", url+"/"+test.uname, "", 200)
			if nil != e {
				t.Error("test["+test.tname+"] failed", e)
				return
			}

			id, e := getId(s)
			if nil != e {
				t.Error("test["+test.tname+"] failed", e)
				return
			}

			s, e = httpInvoke("GET", url+"/"+test.uname+"/"+fmt.Sprint(id), "", 200)
			if nil != e {
				t.Error("test["+test.tname+"] failed", e)
				return
			}

			if !strings.Contains(s, "msssssssso") {
				t.Error(`excepted contains 'msssssssso'`)
				t.Error(`actual is`, s)
			}

			s, e = httpInvoke("GET", url+"/"+test.uname+"/"+fmt.Sprint(id)+"/", "", 200)
			if nil != e {
				t.Error("test["+test.tname+"] failed", e)
				return
			}

			if !strings.Contains(s, "msssssssso") {
				t.Error(`excepted contains 'msssssssso'`)
				t.Error(`actual is`, s)
			}
		}
	})
}

func TestFindByIdWithNotFound(t *testing.T) {
	SrvTest(t, func(db *sql.DB, url string) {

		now := time.Now()
		nows_str := "_" + strconv.Itoa(now.Year()) + months[now.Month()]
		if "postgres" != *db_drv {
			nows_str = ""
		}

		for _, test := range []struct{ tname, uname string }{{tname: "tpt_alert_cookies", uname: "alert_cookies"},
			{tname: "tpt_alert_histories" + nows_str, uname: "alerts"},
			{tname: "tpt_histories" + nows_str, uname: "histories"}} {

			s, e := httpInvoke("GET", url+"/"+test.uname+"/123", "", 200)
			if nil != e {
				if err, ok := e.(commons.RuntimeError); !ok || err.Code() != 404 {
					t.Error(e)
					continue
				} else {
					t.Log(e)
				}
			} else {
				t.Error(`excepted contains "code":404`)
				t.Error(`actual is`, s)
			}

			s, e = httpInvoke("GET", url+"/"+test.uname+"/123/", "", 200)
			if nil != e {
				if err, ok := e.(commons.RuntimeError); !ok || err.Code() != 404 {
					t.Error(e)
				} else {
					t.Log(e)
				}
			} else {
				t.Error(`excepted contains "code":404`)
				t.Error(`actual is`, s)
			}

		}
	})
}

func TestCount(t *testing.T) {
	SrvTest(t, func(db *sql.DB, url string) {

		now := time.Now()
		nows_str := "_" + strconv.Itoa(now.Year()) + months[now.Month()]
		if "postgres" != *db_drv {
			nows_str = ""
		}

		for _, test := range []struct{ tname, uname string }{{tname: "tpt_alert_cookies", uname: "alert_cookies"},
			{tname: "tpt_alert_histories" + nows_str, uname: "alerts"},
			{tname: "tpt_histories" + nows_str, uname: "histories"}} {

			now_func := "now()"
			if "odbc_with_mssql" == *db_drv {
				now_func = "GETDATE()"
			}

			for i := 0; i < 10; i++ {
				var e error
				if strings.HasPrefix(test.tname, "tpt_alert") {
					_, e = db.Exec(fmt.Sprintf("insert into %v(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) values(%v, 'mo', 1%v, 1, 0, '123', 1, 'content is alerted', 'ss', "+now_func+")", test.tname, i, i%2))
				} else {
					_, e = db.Exec(fmt.Sprintf("insert into %v(action_id, managed_type, managed_id, current_value, sampled_at) values(%v, 'mo', 1%v, 1, "+now_func+")", test.tname, i, i%2))
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

		now := time.Now()
		nows_str := "_" + strconv.Itoa(now.Year()) + months[now.Month()]
		if "postgres" != *db_drv {
			nows_str = ""
		}

		for _, test := range []struct{ tname, uname string }{{tname: "tpt_alert_cookies", uname: "alert_cookies"},
			{tname: "tpt_alert_histories" + nows_str, uname: "alerts"},
			{tname: "tpt_histories" + nows_str, uname: "histories"}} {

			now_func := "now()"
			if "odbc_with_mssql" == *db_drv {
				now_func = "GETDATE()"
			}

			for i := 0; i < 10; i++ {
				var e error
				if strings.HasPrefix(test.tname, "tpt_alert") {
					_, e = db.Exec(fmt.Sprintf("insert into %v(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) values(%v, 'mo', 1%v, 1, 0, '123', 1, 'content is alerted', 'ss', "+now_func+")", test.tname, i, i%2))
				} else {
					_, e = db.Exec(fmt.Sprintf("insert into %v(action_id, managed_type, managed_id, current_value, sampled_at) values(%v, 'mo', 1%v, 1, "+now_func+")", test.tname, i, i%2))
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

func TestFindByActionId(t *testing.T) {
	SrvTest(t, func(db *sql.DB, url string) {
		for _, test := range []struct{ tname, uname string }{{tname: "tpt_alert_cookies", uname: "alert_cookies"}} {

			now_func := "now()"
			if "odbc_with_mssql" == *db_drv {
				now_func = "GETDATE()"
			}

			for i := 1; i < 12; i++ {
				_, e := db.Exec(fmt.Sprintf("insert into tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) values(%v, 'dd', 1%v, 1, 0, '123', 1, 'content is alerted', 'ss', "+now_func+")", i, 2%2))

				if nil != e {
					t.Error("test["+test.tname+"] failed", e)
					return
				}
			}

			_, e := db.Exec(fmt.Sprintf("insert into tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) values(%v, 'msssssssso', 1%v, 1, 0, '123', 1, 'content is alerted', 'ss', "+now_func+")", 123, 2%2))
			if nil != e {
				t.Error("test tpt_alert_cookies by action_id failed", e)
				return
			}
			s, e := httpInvoke("GET", url+"/alert_cookies/@123", "", 200)
			if nil != e {
				t.Error("test tpt_alert_cookies by action_id failed", e)
				return
			}

			if !strings.Contains(s, "msssssssso") {
				t.Error(`excepted contains 'msssssssso'`)
				t.Error(`actual is`, s)
			}

			s, e = httpInvoke("GET", url+"/alert_cookies/@123/", "", 200)
			if nil != e {
				t.Error("test tpt_alert_cookies by action_id failed", e)
				return
			}

			if !strings.Contains(s, "msssssssso") {
				t.Error(`excepted contains 'msssssssso'`)
				t.Error(`actual is`, s)
			}
		}
	})
}

func TestFindByActionIdWithNotFound(t *testing.T) {
	SrvTest(t, func(db *sql.DB, url string) {
		s, e := httpInvoke("GET", url+"/alert_cookies/@123", "", 200)
		if nil != e {
			if err, ok := e.(commons.RuntimeError); !ok || err.Code() != 404 {
				t.Error(e)
			} else {
				t.Log(e)
			}
		} else {
			t.Error(`excepted contains "code":404`)
			t.Error(`actual is`, s)
		}

		s, e = httpInvoke("GET", url+"/alert_cookies/@123/", "", 200)
		if nil != e {
			if err, ok := e.(commons.RuntimeError); !ok || err.Code() != 404 {
				t.Error(e)
			} else {
				t.Log(e)
			}
		} else {
			t.Error(`excepted contains "code":404`)
			t.Error(`actual is`, s)
		}
	})
}
