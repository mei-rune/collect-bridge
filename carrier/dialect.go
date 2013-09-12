package carrier

import (
	"database/sql"
	"strconv"
	"time"
)

type Dialect interface {
	removeAlertCookies(tx *sql.Tx, entity *AlertEntity) error
	saveAlertCookies(tx *sql.Tx, entity *AlertEntity) error
	saveAlertHistory(tx *sql.Tx, entity *AlertEntity) error
	saveHistory(tx *sql.Tx, entity *HistoryEntity) error
	saveNotification(tx *sql.Tx, queue, id interface{}, entity *AlertEntity, now time.Time) error
}

type PostgresqlDialect struct {
	delayed_job_table string
}

func (self *PostgresqlDialect) removeAlertCookies(tx *sql.Tx, entity *AlertEntity) error {
	_, e := tx.Exec("DELETE FROM tpt_alert_cookies WHERE action_id = $1", entity.ActionId)
	return e
}

func (self *PostgresqlDialect) saveAlertCookies(tx *sql.Tx, entity *AlertEntity) error {
	var id int64
	e := tx.QueryRow("SELECT id FROM tpt_alert_cookies WHERE action_id = $1", entity.ActionId).Scan(&id)
	if nil != e {
		if sql.ErrNoRows != e {
			return e
		}
		_, e = tx.Exec(`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, level, content, current_value, triggered_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			entity.ActionId, entity.ManagedType, entity.ManagedId, entity.Status, entity.PreviousStatus, entity.EventId, entity.SequenceId, entity.Level, entity.Content, entity.CurrentValue, entity.TriggeredAt)
		return e
	} else {
		_, e = tx.Exec(`UPDATE tpt_alert_cookies SET status = $1, previous_status = $2, event_id = $3, sequence_id = $4, level = $5, content = $6, current_value = $7, triggered_at = $8  WHERE id = $9`,
			entity.Status, entity.PreviousStatus, entity.EventId, entity.SequenceId, entity.Level, entity.Content, entity.CurrentValue, entity.TriggeredAt, id)
		return e
	}
}

func (self *PostgresqlDialect) saveAlertHistory(tx *sql.Tx, entity *AlertEntity) error {
	_, e := tx.Exec("INSERT INTO tpt_alert_histories_"+strconv.Itoa(entity.TriggeredAt.Year())+months[entity.TriggeredAt.Month()]+
		"(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		entity.ActionId, entity.ManagedType, entity.ManagedId, entity.Status, entity.PreviousStatus, entity.EventId, entity.SequenceId, entity.Content, entity.CurrentValue, entity.TriggeredAt)
	return e
}

func (self *PostgresqlDialect) saveNotification(tx *sql.Tx, queue, id interface{}, entity *AlertEntity, now time.Time) error {
	//fmt.Println(self.delayed_job_table, entity.NotificationData.Priority, 0, queue, entity.NotificationData.PayloadObject, id, now, now, now)
	_, e := tx.Exec(self.delayed_job_table, entity.NotificationData.Priority, 0, queue, entity.NotificationData.PayloadObject, id, now, now, now)
	return e
}

func (self *PostgresqlDialect) saveHistory(tx *sql.Tx, entity *HistoryEntity) error {
	_, e := tx.Exec("INSERT INTO tpt_histories_"+strconv.Itoa(entity.SampledAt.Year())+months[entity.SampledAt.Month()]+
		"(action_id, managed_type, managed_id, current_value, sampled_at) VALUES ($1, $2, $3, $4, $5)",
		entity.ActionId, entity.ManagedType, entity.ManagedId, entity.CurrentValue, entity.SampledAt)
	return e
}

type GenSqlDialect struct {
	delayed_job_table string
}

func (self *GenSqlDialect) removeAlertCookies(tx *sql.Tx, entity *AlertEntity) error {
	_, e := tx.Exec("DELETE FROM tpt_alert_cookies WHERE action_id = ?", entity.ActionId)
	return e
}

func (self *GenSqlDialect) saveAlertCookies(tx *sql.Tx, entity *AlertEntity) error {
	var id int64
	e := tx.QueryRow("SELECT id FROM tpt_alert_cookies WHERE action_id = ?", entity.ActionId).Scan(&id)
	if nil != e {
		if sql.ErrNoRows != e {
			return e
		}
		_, e = tx.Exec(`INSERT INTO tpt_alert_cookies(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, level, content, current_value, triggered_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			entity.ActionId, entity.ManagedType, entity.ManagedId, entity.Status, entity.PreviousStatus, entity.EventId, entity.SequenceId, entity.Level, entity.Content, entity.CurrentValue, entity.TriggeredAt)
		return e
	} else {
		_, e = tx.Exec(`UPDATE tpt_alert_cookies SET status = ?, previous_status = ?, event_id = ?, sequence_id = ?, level = ?, content = ?, current_value = ?, triggered_at = ?  WHERE id = ?`,
			entity.Status, entity.PreviousStatus, entity.EventId, entity.SequenceId, entity.Level, entity.Content, entity.CurrentValue, entity.TriggeredAt, id)
		return e
	}
}

func (self *GenSqlDialect) saveAlertHistory(tx *sql.Tx, entity *AlertEntity) error {
	_, e := tx.Exec("INSERT INTO tpt_alert_histories(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		entity.ActionId, entity.ManagedType, entity.ManagedId, entity.Status, entity.PreviousStatus, entity.EventId, entity.SequenceId, entity.Content, entity.CurrentValue, entity.TriggeredAt)
	return e
}

func (self *GenSqlDialect) saveNotification(tx *sql.Tx, queue, id interface{}, entity *AlertEntity, now time.Time) error {
	_, e := tx.Exec(self.delayed_job_table, entity.NotificationData.Priority, 0, queue, entity.NotificationData.PayloadObject, id, now, now, now)
	return e
}

func (self *GenSqlDialect) saveHistory(tx *sql.Tx, entity *HistoryEntity) error {
	_, e := tx.Exec("INSERT INTO tpt_histories(action_id, managed_type, managed_id, current_value, sampled_at) VALUES (?, ?, ?, ?, ?)",
		entity.ActionId, entity.ManagedType, entity.ManagedId, entity.CurrentValue, entity.SampledAt)
	return e
}

type MySqlDialect struct {
	GenSqlDialect
}

func (self *MySqlDialect) saveAlertCookies(tx *sql.Tx, entity *AlertEntity) error {
	entity.TriggeredAt = entity.TriggeredAt.UTC()
	return self.GenSqlDialect.saveAlertCookies(tx, entity)
}

func (self *MySqlDialect) saveAlertHistory(tx *sql.Tx, entity *AlertEntity) error {
	entity.TriggeredAt = entity.TriggeredAt.UTC()
	return self.GenSqlDialect.saveAlertHistory(tx, entity)
}

func (self *MySqlDialect) saveHistory(tx *sql.Tx, entity *HistoryEntity) error {
	entity.SampledAt = entity.SampledAt.UTC()
	return self.GenSqlDialect.saveHistory(tx, entity)
}

func (self *MySqlDialect) saveNotification(tx *sql.Tx, queue, id interface{}, entity *AlertEntity, now time.Time) error {
	entity.TriggeredAt = entity.TriggeredAt.UTC()
	now = now.UTC()
	return self.GenSqlDialect.saveNotification(tx, queue, id, entity, now)
}

type MsSqlDialect struct {
	GenSqlDialect
}

func (self *MsSqlDialect) saveAlertCookies(tx *sql.Tx, entity *AlertEntity) error {
	t := entity.TriggeredAt
	entity.TriggeredAt = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())
	return self.GenSqlDialect.saveAlertCookies(tx, entity)
}

func (self *MsSqlDialect) saveAlertHistory(tx *sql.Tx, entity *AlertEntity) error {
	t := entity.TriggeredAt
	entity.TriggeredAt = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())
	return self.GenSqlDialect.saveAlertHistory(tx, entity)
}

func (self *MsSqlDialect) saveHistory(tx *sql.Tx, entity *HistoryEntity) error {
	t := entity.SampledAt
	entity.SampledAt = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())
	return self.GenSqlDialect.saveHistory(tx, entity)
}

func (self *MsSqlDialect) saveNotification(tx *sql.Tx, queue, id interface{}, entity *AlertEntity, now time.Time) error {
	now = now.UTC()
	now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, now.Location())
	return self.GenSqlDialect.saveNotification(tx, queue, id, entity, now)
	//_, e := tx.Exec(self.delayed_job_table, entity.NotificationData.Priority, 0, queue, entity.NotificationData.PayloadObject, id, now, now)
	//return e
}
