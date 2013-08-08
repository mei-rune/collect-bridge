package carrier

import (
	"database/sql"
	"encoding/json"
	"github.com/runner-mei/delayed_job"
	"testing"
	"time"
)

// <class name="RedisCommand" base="Action">
//   <property name="command">
//     <restriction base="string">
//       <maxLength>10</maxLength>
//     </restriction>
//   </property>
//   <property name="arg0">
//     <restriction base="string">
//       <maxLength>200</maxLength>
//     </restriction>
//   </property>
//   <property name="arg1">
//     <restriction base="string">
//       <maxLength>200</maxLength>
//     </restriction>
//   </property>
//   <property name="arg2">
//     <restriction base="string">
//       <maxLength>200</maxLength>
//     </restriction>
//   </property>
//   <property name="arg3">
//     <restriction base="string">
//       <maxLength>200</maxLength>
//     </restriction>
//   </property>
//   <property name="arg4">
//     <restriction base="string">
//       <maxLength>200</maxLength>
//     </restriction>
//   </property>
// </class>

// <class name="DbCommand" base="Action">
//   <property name="drv">
//     <restriction base="string">
//       <maxLength>200</maxLength>
//     </restriction>
//   </property>
//   <property name="url">
//     <restriction base="string">
//       <maxLength>200</maxLength>
//     </restriction>
//   </property>
//   <property name="script">
//     <restriction base="string">
//       <maxLength>2000</maxLength>
//     </restriction>
//   </property>
// </class>

// <class name="Mail" base="Action">
//   <property name="from_address">
//     <restriction base="string">
//       <maxLength>200</maxLength>
//     </restriction>
//   </property>

//   <property name="to_address">
//     <restriction base="string">
//       <maxLength>200</maxLength>
//       <required />
//     </restriction>
//   </property>

//   <property name="cc_address">
//     <restriction base="string">
//       <maxLength>200</maxLength>
//     </restriction>
//   </property>

//   <property name="bcc_address">
//     <restriction base="string">
//       <maxLength>200</maxLength>
//     </restriction>
//   </property>

//   <property name="subject">
//     <restriction base="string">
//       <maxLength>200</maxLength>
//       <required />
//     </restriction>
//   </property>

//   <property name="content_type">
//     <restriction base="string">
//       <!-- html, text -->
//       <maxLength>200</maxLength>
//     </restriction>
//   </property>

//   <property name="content">
//     <restriction base="string">
//       <maxLength>2000</maxLength>
//       <required />
//     </restriction>
//   </property>
// </class>

// <class name="Syslog" base="Action">
//   <property name="facility">
//     <restriction base="string">
//       <maxLength>250</maxLength>
//       <enumeration>
//         <value>kernel</value>
//         <value>user</value>
//         <value>user</value>
//         <value>mail</value>
//         <value>daemon</value>
//         <value>auth</value>
//         <value>syslog</value>
//         <value>lpr</value>
//         <value>news</value>
//         <value>uucp</value>
//         <value>cron</value>
//         <value>authpriv</value>
//         <value>system0</value>
//         <value>system1</value>
//         <value>system2</value>
//         <value>system3</value>
//         <value>system4</value>
//         <value>local0</value>
//         <value>local1</value>
//         <value>local2</value>
//         <value>local3</value>
//         <value>local4</value>
//         <value>local5</value>
//         <value>local6</value>
//         <value>local7</value>
//       </enumeration>
//     </restriction>
//   </property>

//   <property name="severity">
//     <restriction base="string">
//       <maxLength>250</maxLength>
//       <enumeration>
//         <value>emerg</value>
//         <value>alert</value>
//         <value>crit</value>
//         <value>err</value>
//         <value>waining</value>
//         <value>notice</value>
//         <value>info</value>
//         <value>debug</value>
//       </enumeration>
//     </restriction>
//   </property>

//   <property name="tag">
//     <restriction base="string">
//       <maxLength>250</maxLength>
//     </restriction>
//   </property>

//   <property name="content">
//     <restriction base="string">
//       <maxLength>250</maxLength>
//       <required />
//     </restriction>
//   </property>
// </class>

// <class name="ExecCommand" base="Action">
//   <property name="work_directory">
//     <restriction base="string">
//       <maxLength>512</maxLength>
//     </restriction>
//   </property>
//   <property name="prompt">
//     <restriction base="string">
//       <maxLength>250</maxLength>
//     </restriction>
//   </property>
//   <property name="command">
//     <restriction base="string">
//       <maxLength>250</maxLength>
//       <required />
//     </restriction>
//   </property>
// </class>

var notifications = []map[string]interface{}{map[string]interface{}{"name": "test1",
	"description": "",
	"type":        "redis_command",
	"parent_type": "alert",
	"parent_id":   "1",
	"created_at":  "2013-07-13T14:13:28.7024412+08:00",
	"updated_at":  "2013-07-13T14:13:28.7024412+08:00",
	"command":     "SET",
	"arg0":        "$action_id",
	"arg1":        "$current_value"}}

func TestNotifications(t *testing.T) {
	bs, e := json.Marshal(notifications)
	if nil != e {
		t.Error(e)
		return
	}
	id := "TestNotifications"

	rules := json.RawMessage(bs)
	handler := map[string]interface{}{"arguments": map[string]interface{}{"action_id": 1,
		"managed_type":    "managed_object",
		"managed_id":      1,
		"status":          1,
		"previous_status": 0,
		"event_id":        "123",
		"sequence_id":     1,
		"current_value":   "2",
		"triggered_at":    "2013-07-13T14:13:28.7024412+08:00"},
		"type":  "multiplexed",
		"rules": &rules}

	bs, e = json.Marshal(handler)
	if nil != e {
		t.Error(e)
		return
	}

	bs, e = json.Marshal(map[string]interface{}{"id": id, "priority": 0, "payload_object": string(bs)})
	if nil != e {
		t.Error(e)
		return
	}

	nowt := time.Now()
	now, _ := nowt.MarshalJSON()

	js := `[{"action_id":1,
        "content": "content is alerted",
        "current_value":"2",
        "managed_id":1,
        "managed_type":"managed_object",
        "metric":"sys",
        "name":"this is a test alert",
        "status":1,
        "previous_status": 0,
        "event_id": "123",
        "sequence_id": 1,
        "trigger_id":"1",
        "triggered_at":` + string(now) + `,"notification":` + string(bs) + `}]`

	t.Log(js)

	delayed_job.WorkTest(t, func(worker *delayed_job.TestWorker) {
		SrvTest(t, func(db *sql.DB, url string) {
			_, e := httpInvoke("PUT", url+"/alerts", js, 200)
			if nil != e {
				t.Error(e)
				return
			}

			assertCount := func(t *testing.T, sql string, excepted int64) {
				count := int64(-1)
				e := db.QueryRow(sql).Scan(&count)
				if nil != e {
					t.Error(e)
					return
				}

				if count != excepted {
					t.Error("excepted \"", sql, "\" is ", excepted, ", actual is ", count)
				}
			}

			assertCount(t, "SELECT count(*) FROM "+*delayed_job_table_name+" where handler_id = '"+id+"'", 1)

			i, j, e := worker.WorkOff(1)
			if nil != e {
				t.Error(e)
				return
			}
			assertCount(t, "SELECT count(*) FROM "+*delayed_job_table_name+" where handler_id = '"+id+"'", 0)
			if i != 1 {
				t.Log("success is", i, "failed is", j)
				t.Error("excepted job count is 1, excepted is", i)
				return
			}
		})
	})
}
