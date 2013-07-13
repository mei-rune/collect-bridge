package carrier

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func httpInvoke(action, url string, msg string, exceptedCode int) error {
	//fmt.Println(msg)
	req, err := http.NewRequest(action, url, bytes.NewBufferString(msg))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if nil != e {
		return e
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
			return fmt.Errorf("%v: error", resp.StatusCode)
		}

		if nil == resp_body || 0 == len(resp_body) {
			return fmt.Errorf("%v: error", resp.StatusCode)
		}

		return fmt.Errorf("%v: %v", resp.StatusCode, string(resp_body))
	}

	return nil
}

func TestAlertsServer(t *testing.T) {
	SrvTest(t, func(db *sql.DB, url string) {
		nowt := time.Now()
		now, _ := nowt.MarshalJSON()

		e := httpInvoke("PUT", url+"/alerts", `[{"action_id": 123,
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

		e := httpInvoke("PUT", url+"/alerts", `[{"action_id":1,
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

func TestAlertsCookies(t *testing.T) {
	SrvTest(t, func(db *sql.DB, url string) {
		nowt := time.Now()
		now, _ := nowt.MarshalJSON()

		e := httpInvoke("PUT", url+"/alerts", `[{"action_id": 123,
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

		e = httpInvoke("PUT", url+"/alerts", `[{"action_id": 123,
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

		if nil == entities || 1 != len(entities) {
			t.Error("nil == entities || 1 != len(entities)")
			return
		}

		AssertAlerts(t, entities[0], 123, 0, "65", nowt, "mo", 123)

	})
}

func TestHistoriesServer(t *testing.T) {
	SrvTest(t, func(db *sql.DB, url string) {
		nowt := time.Now()
		now, _ := nowt.MarshalJSON()

		e := httpInvoke("PUT", url+"/histories", `[{"action_id": 123,
        "current_value": 23,
        "sampled_at": `+string(now)+`,
        "managed_type": "mo",
        "managed_id": 123}]`, 200)
		if nil != e {
			t.Error(e)
			return
		}

		e = httpInvoke("PUT", url+"/histories", `[{"action_id": 1243,
        "current_value": 233,
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
