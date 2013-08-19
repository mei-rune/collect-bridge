package carrier

import (
	"bitbucket.org/runner_mei/goose"
	"database/sql"
	"fmt"
	"github.com/runner-mei/delayed_job"
	"math"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
	"time"
)

var usageTmpl = template.Must(template.New("config").Parse(
	`test:
    driver: {{.driver}}
    open: {{.url}}

development:
    driver: {{.driver}}
    open: {{.url}}

production:
    driver: {{.driver}}
    open: {{.url}}
`))

func merge(t *testing.T, tag, drv, url string) (string, error) {
	file := filepath.Join("db", "dbconf-"+tag+".yml")
	files := []string{file,
		filepath.Join("..", "db", "dbconf-"+tag+".yml"),
		filepath.Join("carrier", "db", "dbconf-"+tag+".yml"),
		filepath.Join("..", "carrier", "db", "dbconf-"+tag+".yml")}
	for _, s := range files {
		if st, e := os.Stat(s); nil == e && !st.IsDir() {
			file = s
		}
	}

	f, e := os.OpenFile(file, os.O_RDWR|os.O_TRUNC, 0)
	if nil != e {
		return "", e
	}
	e = usageTmpl.Execute(f, map[string]interface{}{"driver": drv, "url": url})
	if nil != e {
		return "", e
	}
	return filepath.Dir(file), nil
}

func SimpleTest(t *testing.T, cb func(db *sql.DB)) {
	var e error
	var path string
	var tag string = *db_drv
	switch *db_drv {
	case "mysql":
		path, e = merge(t, tag, "mysql", *db_url)
	case "mymysql":
		tag = "mysql"
		path, e = merge(t, tag, "mysql", *db_url)
	case "postgres":
		tag = "postgresql"
		path, e = merge(t, tag, "postgres", *db_url)
	case "odbc_with_mssql":
		tag = "mssql"
		path, e = merge(t, tag, "mssql", *db_url)
	default:
		t.Error("unsupported driver -- " + *db_drv)
		return
	}
	if nil != e {
		t.Error("create config failed,", e)
		return
	}

	goose.Run("-path="+path, "-tag="+tag, "reset")

	drv := *db_drv
	if strings.HasPrefix(drv, "odbc_with_") {
		drv = "odbc"
	}

	db, e := sql.Open(drv, *db_url)
	if nil != e {
		t.Error("connect to db failed,", e)
		return
	}

	e = Main(true)
	if nil != e {
		t.Error(e)
		return
	}

	delayed_job.SetDbUrl(*db_drv, *db_url)
	cb(db)
}

func SrvTest(t *testing.T, cb func(db *sql.DB, url string)) {
	SrvTest2(t, func(db *sql.DB, db_drv, db_url, url string) {
		cb(db, url)
	})
}

func SrvTest2(t *testing.T, cb func(db *sql.DB, db_drv, db_url, url string)) {
	SimpleTest(t, func(db *sql.DB) {
		e := Main(true)
		if nil != e {
			t.Error(e)
			return
		}
		if nil == server_instance {
			t.Error("server_instance is nil")
			return
		}

		defer server_instance.Close()
		hsrv := httptest.NewServer(server_instance)
		defer hsrv.Close()

		fmt.Println("[carrier-test] serving at " + hsrv.URL)

		cb(db, *db_drv, *db_url, hsrv.URL)
	})
}

func SelectAlertHistories(db *sql.DB) ([]*AlertEntity, error) {

	rows, e := db.Query("select " + prejection_sql + " from tpt_alert_histories")
	if nil != e {
		return nil, e
	}

	alertEntities := make([]*AlertEntity, 0, 2)
	for rows.Next() {
		entity := &AlertEntity{}
		e = rows.Scan(
			&entity.Id,             //Id               int64     `json:"id"`
			&entity.ActionId,       //ActionId         int64     `json:"action_id"`
			&entity.ManagedType,    //ManagedType      string    `json:"managed_type"`
			&entity.ManagedId,      //ManagedId        int64     `json:"managed_id"`
			&entity.Status,         //Status           string    `json:"status"`
			&entity.PreviousStatus, //PreviousStatus   int64     `json:"previous_status"`
			&entity.EventId,        //EventId          string    `json:"event_id"`
			&entity.SequenceId,     //SequenceId       int64     `json:"sequence_id"`
			&entity.Content,        //Content          string    `json:"content"`
			&entity.CurrentValue,   //CurrentValue     string    `json:"current_value"`
			&entity.TriggeredAt)    //TriggeredAt      time.Time `json:"triggered_at"`
		if nil != e {
			return nil, e
		}

		alertEntities = append(alertEntities, entity)
	}
	return alertEntities, nil
}

func SelectAlertCookies(db *sql.DB) ([]*AlertEntity, error) {
	rows, e := db.Query("select " + prejection_sql + " from tpt_alert_cookies")
	if nil != e {
		return nil, e
	}

	alertEntities := make([]*AlertEntity, 0, 2)
	for rows.Next() {
		entity := &AlertEntity{}
		e = rows.Scan(
			&entity.Id,             //Id               int64     `json:"id"`
			&entity.ActionId,       //ActionId         int64     `json:"action_id"`
			&entity.ManagedType,    //ManagedType      string    `json:"managed_type"`
			&entity.ManagedId,      //ManagedId        int64     `json:"managed_id"`
			&entity.Status,         //Status           string    `json:"status"`
			&entity.PreviousStatus, //PreviousStatus   int64     `json:"previous_status"`
			&entity.EventId,        //EventId          string    `json:"event_id"`
			&entity.SequenceId,     //SequenceId       int64     `json:"sequence_id"`
			&entity.Content,        //Content          string    `json:"content"`
			&entity.CurrentValue,   //CurrentValue     string    `json:"current_value"`
			&entity.TriggeredAt)    //TriggeredAt      time.Time `json:"triggered_at"`
		if nil != e {
			return nil, e
		}

		alertEntities = append(alertEntities, entity)
	}
	return alertEntities, nil
}

func SelectHistories(db *sql.DB) ([]*HistoryEntity, error) {
	rows, e := db.Query("select id, action_id, managed_type, managed_id, current_value, sampled_at from tpt_histories")
	if nil != e {
		return nil, e
	}

	historyEntities := make([]*HistoryEntity, 0, 2)
	for rows.Next() {
		var id int64
		entity := &HistoryEntity{}
		e = rows.Scan(&id,
			&entity.ActionId,    //ActionId       int64     `json:"action_id"`
			&entity.ManagedType, //ManagedType  string    `json:"managed_type"`
			&entity.ManagedId,
			&entity.CurrentValue, //CurrentValue string    `json:"current_value"`
			&entity.SampledAt)    //TriggeredAt  time.Time `json:"triggered_at"`
		if nil != e {
			return nil, e
		}

		historyEntities = append(historyEntities, entity)
	}
	return historyEntities, nil
}

func AssertAlerts(t *testing.T, entity *AlertEntity, action_id, status, previousStatus int64, eventId string, sequenceId int64, content, value string, now time.Time, mo_type string, mo_id int64) {

	if entity.ActionId != action_id {
		t.Error(" entity.ActionId != action_id, excepted is ", action_id, ", actual is", entity.ActionId)
	}
	if entity.Status != status {
		t.Error(" entity.Status != status, excepted is ", status, ", actual is", entity.Status)
	}
	if entity.PreviousStatus != previousStatus {
		t.Error(" entity.PreviousStatus != previousStatus, excepted is ", previousStatus, ", actual is", entity.PreviousStatus)
	}
	if entity.EventId != eventId {
		t.Error(" entity.EventId != eventId, excepted is ", eventId, ", actual is", entity.EventId)
	}
	if entity.SequenceId != sequenceId {
		t.Error(" entity.SequenceId != sequenceId, excepted is ", sequenceId, ", actual is", entity.SequenceId)
	}
	if entity.Content != content {
		t.Error(" entity.Content != content, excepted is ", content, ", actual is", entity.Content)
	}
	if entity.CurrentValue != value {
		t.Error(" entity.CurrentValue != value, excepted is ", value, ", actual is", entity.CurrentValue)
	}
	if entity.TriggeredAt.Unix() != now.Unix() {
		t.Error(" entity.TriggeredAt != now, excepted is ", now, ", actual is", entity.TriggeredAt)
	}
	if entity.ManagedType != mo_type {
		t.Error(" entity.ManagedType != mo_type, excepted is ", mo_type, ", actual is", entity.ManagedType)
	}
	if entity.ManagedId != mo_id {
		t.Error(" entity.ManagedId != mo_id, excepted is ", mo_id, ", actual is ", entity.ManagedId)
	}

}

func AssertHistories(t *testing.T, entity *HistoryEntity, action_id int64, value float64, now time.Time, mo_type string, mo_id int64) {
	if entity.ActionId != action_id {
		t.Error(" entity.ActionId != action_id, excepted is ", action_id, ", actual is", entity.ActionId)
	}
	if math.Abs(entity.CurrentValue-value) > 0.002 {
		t.Error(" entity.CurrentValue != value, excepted is ", value, ", actual is", entity.CurrentValue)
	}
	if entity.SampledAt.Unix() != now.Unix() {
		t.Error(" entity.SampledAt != now, excepted is ", now, ", actual is", entity.SampledAt)
	}
	if entity.ManagedType != mo_type {
		t.Error(" entity.ManagedType != mo_type, excepted is ", mo_type, ", actual is", entity.ManagedType)
	}
	if entity.ManagedId != mo_id {
		t.Error(" entity.ManagedId != mo_id, excepted is ", mo_id, ", actual is", entity.ManagedId)
	}

}
