package poller

import (
	"carrier"
	"commons/types"
	ds "data_store"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestLoadCookiesWhileStartServer(t *testing.T) {
	srvTest(t, func(client *ds.Client, definitions *types.TableDefinitions) {
		id := ds.CreateItForTest(t, client, "network_device", mo)
		ds.CreateItByParentForTest(t, client, "network_device", id, "wbem_param", wbem_params)
		ds.CreateItByParentForTest(t, client, "network_device", id, "snmp_param", snmp_params)
		mt_id := ds.CreateItByParentForTest(t, client, "network_device", id, "metric_trigger", metric_trigger_for_cpu)
		action_id := ds.CreateItByParentForTest(t, client, "metric_trigger", mt_id, "alert", map[string]interface{}{
			"id":               "123",
			"name":             "this is a test alert",
			"delay_times":      0,
			"expression_style": "json",
			"expression_code": map[string]interface{}{
				"attribute": "a",
				"operator":  ">=",
				"value":     "0"}})

		carrier.SrvTest(t, func(db *sql.DB, url string) {
			*foreignUrl = url

			for _, s := range []string{fmt.Sprintf(`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (%v, 'mo', 1, 1, 0, 'event_id_sss_aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`, action_id),
				`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (2, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`,
				`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (3, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`,
				`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (4, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`} {
				_, e := db.Exec(s)

				if nil != e {
					t.Error(e)
					return
				}
			}

			is_test = true
			*load_cookies = true
			Runforever()
			if nil == server_test {
				t.Error("load trigger failed.")
				return
			}
			defer func() {
				server_test.Stop()
				server_test = nil
			}()

			if nil == server_test.jobs || 0 == len(server_test.jobs) {
				t.Error("load trigger failed.")
				return
			}

			tr_instance := server_test.jobs[mt_id].(*metricJob).Trigger.(*intervalTrigger)
			tr_instance.callAfter()
			//server_test.jobs[mt_id].(*metricJob).callAfter()
			stats := server_test.jobs[mt_id].Stats()
			bs, e := json.MarshalIndent(stats, "", "  ")
			if nil != e {
				t.Error(e)
				return
			}
			if !strings.Contains(string(bs), "event_id_sss_aa") {
				t.Error("load cookies failed.")
				t.Log(string(bs))
			}
		})
	})
}

func TestLoadCookiesWhileOnTick(t *testing.T) {
	srvTest(t, func(client *ds.Client, definitions *types.TableDefinitions) {
		carrier.SrvTest(t, func(db *sql.DB, url string) {
			*foreignUrl = url
			is_test = true
			*load_cookies = true
			Runforever()
			if nil == server_test {
				t.Error("load trigger failed.")
				return
			}
			defer func() {
				server_test.Stop()
				server_test = nil
			}()

			id := ds.CreateItForTest(t, client, "network_device", mo)
			ds.CreateItByParentForTest(t, client, "network_device", id, "wbem_param", wbem_params)
			ds.CreateItByParentForTest(t, client, "network_device", id, "snmp_param", snmp_params)
			mt_id := ds.CreateItByParentForTest(t, client, "network_device", id, "metric_trigger", metric_trigger_for_cpu)
			action_id := ds.CreateItByParentForTest(t, client, "metric_trigger", mt_id, "alert", map[string]interface{}{
				"id":               "123",
				"name":             "this is a test alert",
				"delay_times":      0,
				"expression_style": "json",
				"expression_code": map[string]interface{}{
					"attribute": "a",
					"operator":  ">=",
					"value":     "0"}})

			for _, s := range []string{`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (2, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`,
				`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (3, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`,
				`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (4, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`,
				fmt.Sprintf(`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (%v, 'mo', 1, 1, 0, 'event_id_sss_aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`, action_id)} {
				_, e := db.Exec(s)

				if nil != e {
					t.Error(e)
					return
				}
			}

			server_test.onIdle()

			if nil == server_test || nil == server_test.jobs || 0 == len(server_test.jobs) {
				t.Error("load trigger failed.")
				return
			}

			tr_instance := server_test.jobs[mt_id].(*metricJob).Trigger.(*intervalTrigger)
			tr_instance.callAfter()
			//server_test.jobs[mt_id].(*metricJob).callAfter()
			stats := server_test.jobs[mt_id].Stats()
			bs, e := json.MarshalIndent(stats, "", "  ")
			if nil != e {
				t.Error(e)
				return
			}
			if !strings.Contains(string(bs), "event_id_sss_aa") {
				t.Error("load cookies failed.")
				t.Log(string(bs))
			}
		})
	})
}

func TestLoadCookiesWhileOnTickWithNotfound(t *testing.T) {
	srvTest(t, func(client *ds.Client, definitions *types.TableDefinitions) {
		carrier.SrvTest(t, func(db *sql.DB, url string) {
			*foreignUrl = url
			is_test = true
			*load_cookies = true
			Runforever()
			if nil == server_test {
				t.Error("load trigger failed.")
				return
			}
			defer func() {
				server_test.Stop()
				server_test = nil
			}()

			id := ds.CreateItForTest(t, client, "network_device", mo)
			ds.CreateItByParentForTest(t, client, "network_device", id, "wbem_param", wbem_params)
			ds.CreateItByParentForTest(t, client, "network_device", id, "snmp_param", snmp_params)
			mt_id := ds.CreateItByParentForTest(t, client, "network_device", id, "metric_trigger", metric_trigger_for_cpu)
			ds.CreateItByParentForTest(t, client, "metric_trigger", mt_id, "alert", map[string]interface{}{
				"id":               "123",
				"name":             "this is a test alert",
				"delay_times":      0,
				"expression_style": "json",
				"expression_code": map[string]interface{}{
					"attribute": "a",
					"operator":  ">=",
					"value":     "0"}})

			server_test.onIdle()

			if nil == server_test || nil == server_test.jobs || 0 == len(server_test.jobs) {
				t.Error("load trigger failed.")
				return
			}

			tr_instance := server_test.jobs[mt_id].(*metricJob).Trigger.(*intervalTrigger)
			tr_instance.callAfter()
			stats := server_test.jobs[mt_id].Stats()
			bs, e := json.MarshalIndent(stats, "", "  ")
			if nil != e {
				t.Error(e)
				return
			}
			if !strings.Contains(string(bs), `event_id": "",`) {
				t.Error("load cookies failed.")
				t.Log(string(bs))
			}
		})
	})
}

func TestCookiesIsClear(t *testing.T) {
	srvTest(t, func(client *ds.Client, definitions *types.TableDefinitions) {
		carrier.SrvTest(t, func(db *sql.DB, url string) {
			*foreignUrl = url

			for _, s := range []string{`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (1, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`,
				`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (2, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`,
				`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (3, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`,
				`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (4, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`} {
				_, e := db.Exec(s)

				if nil != e {
					t.Error(e)
					return
				}
			}

			is_test = true
			Runforever()
			if nil == server_test {
				t.Error("load trigger failed.")
				return
			}
			defer func() {
				server_test.Stop()
				server_test = nil
			}()

			count := int64(0)
			e := db.QueryRow(`SELECT count(*) FROM tpt_alert_cookies`).Scan(&count)
			if nil != e {
				t.Error(e)
				return
			}

			if 0 != count {
				t.Error("excepted count is 0, actual is", count)
				return
			}
		})
	})
}

func TestCookiesNotClear(t *testing.T) {
	srvTest(t, func(client *ds.Client, definitions *types.TableDefinitions) {
		mo_id := ds.CreateItForTest(t, client, "network_device", mo)
		ds.CreateItByParentForTest(t, client, "network_device", mo_id, "wbem_param", wbem_params)
		ds.CreateItByParentForTest(t, client, "network_device", mo_id, "snmp_param", snmp_params)
		mt_id := ds.CreateItByParentForTest(t, client, "network_device", mo_id, "metric_trigger", metric_trigger_for_cpu)
		rule_id := ds.CreateItByParentForTest(t, client, "metric_trigger", mt_id, "alert", map[string]interface{}{
			"name":             "this is a test alert",
			"delay_times":      0,
			"expression_style": "json",
			"expression_code": map[string]interface{}{
				"attribute": "services",
				"operator":  ">=",
				"value":     "0"}})

		carrier.SrvTest(t, func(db *sql.DB, url string) {
			*foreignUrl = url
			for _, s := range []string{fmt.Sprintf(`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (%v, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`, rule_id),
				`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (112, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`,
				`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (113, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`,
				`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (114, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`} {
				_, e := db.Exec(s)
				if nil != e {
					t.Error(e)
					return
				}
			}

			is_test = true
			Runforever()
			if nil == server_test {
				t.Error("load trigger failed.")
				return
			}
			defer func() {
				server_test.Stop()
				server_test = nil
			}()

			count := int64(0)
			e := db.QueryRow(`SELECT count(*) FROM tpt_alert_cookies`).Scan(&count)
			if nil != e {
				t.Error(e)
				return
			}

			if 1 != count {
				t.Error("excepted count is 1, actual is", count)
				return
			}
		})
	})
}

func TestCookiesLoadStatus(t *testing.T) {
	srvTest(t, func(client *ds.Client, definitions *types.TableDefinitions) {
		mo_id := ds.CreateItForTest(t, client, "network_device", mo)
		ds.CreateItByParentForTest(t, client, "network_device", mo_id, "wbem_param", wbem_params)
		ds.CreateItByParentForTest(t, client, "network_device", mo_id, "snmp_param", snmp_params)
		mt_id := ds.CreateItByParentForTest(t, client, "network_device", mo_id, "metric_trigger", map[string]interface{}{
			"name":       "this is a test trigger",
			"type":       "metric_trigger",
			"metric":     "sys",
			"expression": "@every 1000m"})
		rule_id := ds.CreateItByParentForTest(t, client, "metric_trigger", mt_id, "alert", map[string]interface{}{
			"name":             "this is a test alert",
			"delay_times":      1110,
			"expression_style": "json",
			"expression_code": map[string]interface{}{
				"attribute": "services",
				"operator":  ">=",
				"value":     "0"}})

		carrier.SrvTest(t, func(db *sql.DB, url string) {
			*foreignUrl = url

			for _, s := range []string{`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (112, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`,
				`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (113, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`,
				`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (114, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`,
				fmt.Sprintf(`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (%v, 'mo', 1, 1, 0, 'aaccccaaaaa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`, rule_id)} {
				_, e := db.Exec(s)
				if nil != e {
					t.Error(e)
					return
				}
			}

			is_test = true
			Runforever()
			if nil == server_test {
				t.Error("load trigger failed.")
				return
			}
			defer func() {
				server_test.Stop()
				server_test = nil
			}()

			count := int64(0)
			e := db.QueryRow(`SELECT count(*) FROM tpt_alert_cookies`).Scan(&count)
			if nil != e {
				t.Error(e)
				return
			}

			if 1 != count {
				t.Error("excepted count is 1, actual is", count)
				return
			}

			job := server_test.jobs[fmt.Sprint(mt_id)]
			if nil == job {
				t.Error("job is not found")
				return
			}

			tr_instance := server_test.jobs[mt_id].(*metricJob).Trigger.(*intervalTrigger)
			//tr_instance.callAfter()
			action := tr_instance.actions[0]
			action.RunAfter()
			stats := action.Stats()
			if "1" != fmt.Sprint(stats["last_status"]) {
				t.Error("last status is not eq 1, action is", stats["last_status"])
			}
			if "aaccccaaaaa" != fmt.Sprint(stats["event_id"]) {
				t.Error("last status is not eq 'aaccccaaaaa', action is", stats["event_id"])
			}
			t.Log(stats)
		})
	})
}
