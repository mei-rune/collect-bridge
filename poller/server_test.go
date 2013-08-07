package poller

import (
	"carrier"
	"commons/types"
	ds "data_store"
	"database/sql"
	"fmt"
	"testing"
)

func TestCookiesIsClear(t *testing.T) {
	srvTest(t, func(client *ds.Client, definitions *types.TableDefinitions) {
		carrier.SrvTest(t, func(db *sql.DB, url string) {
			*foreignUrl = url

			_, e := db.Exec(`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (1, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');
      INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (2, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');
      INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (3, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');
      INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (4, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`)

			if nil != e {
				t.Error(e)
				return
			}

			is_test = true
			Runforever()
			count := int64(0)
			e = db.QueryRow(`SELECT count(*) FROM tpt_alert_cookies`).Scan(&count)
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

			_, e := db.Exec(fmt.Sprintf(`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (%v, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');
      INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (112, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');
      INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (113, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');
      INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (114, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`, rule_id))
			if nil != e {
				t.Error(e)
				return
			}

			is_test = true
			Runforever()
			count := int64(0)
			e = db.QueryRow(`SELECT count(*) FROM tpt_alert_cookies`).Scan(&count)
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

			_, e := db.Exec(fmt.Sprintf(`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (%v, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');
      INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (112, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');
      INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (113, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');
      INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (114, 'mo', 1, 1, 0, 'aa', 1, 'abc', 'ww', '2013-08-05 12:12:12');`, rule_id))
			if nil != e {
				t.Error(e)
				return
			}

			is_test = true
			Runforever()
			count := int64(0)
			e = db.QueryRow(`SELECT count(*) FROM tpt_alert_cookies`).Scan(&count)
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
			stats := job.Stats()
			if "1" == fmt.Sprint(stats["last_status"]) {
				t.Error("last status is not eq 1")
			}
			if "0" == fmt.Sprint(stats["repeated"]) {
				t.Error("repeated is not eq 1, test is invalid, it is run")
			}

		})
	})
}
