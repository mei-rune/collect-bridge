package carrier

import (
	"bytes"
	"commons"
	"commons/types"
	ds "data_store"
	"database/sql"
	"encoding/json"
	"errors"
	"expvar"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/ziutek/mymysql/godrv"
	"io"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	listenAddress          = flag.String("carrier.listen", ":37074", "the address of http")
	db_url                 = flag.String("data_db.url", "host=127.0.0.1 dbname=tpt_data user=tpt password=extreme sslmode=disable", "the db url")
	db_drv                 = flag.String("data_db.driver", "postgres", "the db driver")
	goroutines             = flag.Int("data_db.connections", 10, "the db connection number")
	delayed_job_table_name = flag.String("delayed_job_table_name", "tpt_delayed_jobs", "the table name of delayed job")

	isNumericParams         = true
	server_instance *server = nil
)

var (
	id_column = &types.ColumnDefinition{types.AttributeDefinition{Name: "id",
		Type:       types.GetTypeDefinition("objectId"),
		Collection: types.COLLECTION_UNKNOWN}}

	action_id_column = &types.ColumnDefinition{types.AttributeDefinition{Name: "action_id",
		Type:       types.GetTypeDefinition("integer"),
		Collection: types.COLLECTION_UNKNOWN}}

	status_column = &types.ColumnDefinition{types.AttributeDefinition{Name: "status",
		Type:       types.GetTypeDefinition("integer"),
		Collection: types.COLLECTION_UNKNOWN}}

	previous_status_column = &types.ColumnDefinition{types.AttributeDefinition{Name: "previous_status",
		Type:       types.GetTypeDefinition("integer"),
		Collection: types.COLLECTION_UNKNOWN}}

	event_id_column = &types.ColumnDefinition{types.AttributeDefinition{Name: "event_id",
		Type:       types.GetTypeDefinition("string"),
		Collection: types.COLLECTION_UNKNOWN}}

	sequence_id_column = &types.ColumnDefinition{types.AttributeDefinition{Name: "sequence_id",
		Type:       types.GetTypeDefinition("integer"),
		Collection: types.COLLECTION_UNKNOWN}}

	level_column = &types.ColumnDefinition{types.AttributeDefinition{Name: "level",
		Type:       types.GetTypeDefinition("integer"),
		Collection: types.COLLECTION_UNKNOWN}}

	managed_type_column = &types.ColumnDefinition{types.AttributeDefinition{Name: "managed_type",
		Type:       types.GetTypeDefinition("string"),
		Collection: types.COLLECTION_UNKNOWN}}

	managed_id_column = &types.ColumnDefinition{types.AttributeDefinition{Name: "managed_id",
		Type:       types.GetTypeDefinition("objectId"),
		Collection: types.COLLECTION_UNKNOWN}}

	content_column = &types.ColumnDefinition{types.AttributeDefinition{Name: "content",
		Type:       types.GetTypeDefinition("string"),
		Collection: types.COLLECTION_UNKNOWN}}

	alert_current_value_column = &types.ColumnDefinition{types.AttributeDefinition{Name: "current_value",
		Type:       types.GetTypeDefinition("string"),
		Collection: types.COLLECTION_UNKNOWN}}

	triggered_at_column = &types.ColumnDefinition{types.AttributeDefinition{Name: "triggered_at",
		Type:       types.GetTypeDefinition("datetime"),
		Collection: types.COLLECTION_UNKNOWN}}

	history_current_value_column = &types.ColumnDefinition{types.AttributeDefinition{Name: "current_value",
		Type:       types.GetTypeDefinition("decimal"),
		Collection: types.COLLECTION_UNKNOWN}}

	sampled_at_column = &types.ColumnDefinition{types.AttributeDefinition{Name: "sampled_at",
		Type:       types.GetTypeDefinition("datetime"),
		Collection: types.COLLECTION_UNKNOWN}}

	tpt_alert_history = &types.TableDefinition{Name: "AlertHistory",
		UnderscoreName: "alert_history",
		CollectionName: "tpt_alert_histories",
		Id:             id_column}

	tpt_alert_cookies = &types.TableDefinition{Name: "AlertCookies",
		UnderscoreName: "alert_cookies",
		CollectionName: "tpt_alert_cookies",
		Id:             id_column}

	tpt_history = &types.TableDefinition{Name: "History",
		UnderscoreName: "history",
		CollectionName: "tpt_histories",
		Id:             id_column}
)

func init() {

	alert_attributes := map[string]*types.ColumnDefinition{id_column.Name: id_column,
		action_id_column.Name:           action_id_column,           // 	ActionId     int64     `json:"action_id"`
		status_column.Name:              status_column,              // 	Status       int64     `json:"status"`
		previous_status_column.Name:     previous_status_column,     //
		event_id_column.Name:            event_id_column,            //
		sequence_id_column.Name:         sequence_id_column,         //
		content_column.Name:             content_column,             //
		alert_current_value_column.Name: alert_current_value_column, // 	CurrentValue string    `json:"current_value"`
		triggered_at_column.Name:        triggered_at_column,        // 	TriggeredAt  time.Time `json:"triggered_at"`
		managed_type_column.Name:        managed_type_column,        // 	ManagedType  string    `json:"managed_type"`
		managed_id_column.Name:          managed_id_column}          // 	ManagedId    int64     `json:"managed_id"`

	tpt_alert_history.OwnAttributes = alert_attributes
	tpt_alert_history.Attributes = alert_attributes

	cookies_attributes := map[string]*types.ColumnDefinition{id_column.Name: id_column,
		action_id_column.Name:           action_id_column,       // 	ActionId     int64     `json:"action_id"`
		status_column.Name:              status_column,          // 	Status       int64     `json:"status"`
		previous_status_column.Name:     previous_status_column, //
		event_id_column.Name:            event_id_column,        //
		sequence_id_column.Name:         sequence_id_column,     //
		level_column.Name:               level_column,
		content_column.Name:             content_column,             //
		alert_current_value_column.Name: alert_current_value_column, // 	CurrentValue string    `json:"current_value"`
		triggered_at_column.Name:        triggered_at_column,        // 	TriggeredAt  time.Time `json:"triggered_at"`
		managed_type_column.Name:        managed_type_column,        // 	ManagedType  string    `json:"managed_type"`
		managed_id_column.Name:          managed_id_column}          // 	ManagedId    int64     `json:"managed_id"`

	tpt_alert_cookies.OwnAttributes = cookies_attributes
	tpt_alert_cookies.Attributes = cookies_attributes

	history_attributes := map[string]*types.ColumnDefinition{id_column.Name: id_column,
		action_id_column.Name:             action_id_column,             // 	ActionId       int64     `json:"action_id"`
		history_current_value_column.Name: history_current_value_column, // 	CurrentValue float64   `json:"current_value"`
		sampled_at_column.Name:            sampled_at_column,            // 	SampledAt    time.Time `json:"sampled_at"`
		managed_type_column.Name:          managed_type_column,          // 	ManagedType  string    `json:"managed_type"`
		managed_id_column.Name:            managed_id_column}            // 	ManagedId    int64     `json:"managed_id"`

	tpt_history.OwnAttributes = history_attributes
	tpt_history.Attributes = history_attributes
}

type request_object struct {
	c        chan error
	request  *http.Request
	response http.ResponseWriter
	cb       func(s *server, ctx *context, response http.ResponseWriter, request *http.Request)
}

type server struct {
	c          chan *request_object
	wait       sync.WaitGroup
	last_error *expvar.String
}

type context struct {
	db      *sql.DB
	dialect Dialect
	drv     string
	url     string
}

func (self *server) Close() {
	close(self.c)
	self.wait.Wait()
}

func (self *server) serve(i int, ctx *context) {

	is_running := true
	defer func() {
		if e := recover(); nil != e {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("[panic]%v", e))
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}
			msg := buffer.String()
			self.last_error.Set(msg)
			log.Println(msg)

			if !is_running {
				os.Exit(-1)
			}
		}

		log.Println("carrir is exit.")
		self.wait.Done()
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for is_running {
		select {
		case <-ticker.C:
			//fmt.Println("tick")
		case req, ok := <-self.c:
			if !ok {
				is_running = false
			} else {
				self.call(ctx, req)
			}
		}
	}
}

func (self *server) call(ctx *context, obj *request_object) {
	defer func() {
		if e := recover(); nil != e {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("[panic]%v", e))
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}

			obj.c <- errors.New(buffer.String())
			close(obj.c)
		}
	}()
	self.run(ctx, obj)

	obj.c <- nil
	close(obj.c)
}

func (self *server) run(ctx *context, obj *request_object) {
	defer func() {
		if e := recover(); nil != e {
			obj.response.WriteHeader(http.StatusInternalServerError)
			obj.response.Write([]byte(fmt.Sprintf("[panic]%v", e)))
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				obj.response.Write([]byte(fmt.Sprintf("    %s:%d\r\n", file, line)))
			}
		}
	}()

	obj.cb(self, ctx, obj.response, obj.request)
}

type Notification struct {
	Id            string `json:"id"`
	Priority      int    `json:"priority"`
	Queue         string `json:"queue"`
	PayloadObject string `json:"payload_object"`
}

type AlertEntity struct {
	Id               int64         `json:"id"`
	ActionId         int64         `json:"action_id"`
	Status           int64         `json:"status"`
	PreviousStatus   int64         `json:"previous_status"`
	EventId          string        `json:"event_id"`
	SequenceId       int64         `json:"sequence_id"`
	Level            int64         `json:"level"`
	Content          string        `json:"content"`
	CurrentValue     string        `json:"current_value"`
	TriggeredAt      time.Time     `json:"triggered_at"`
	ManagedType      string        `json:"managed_type"`
	ManagedId        int64         `json:"managed_id"`
	NotificationData *Notification `json:"notification,omitempty"`
}

type HistoryEntity struct {
	Id           int64     `json:"id"`
	ActionId     int64     `json:"action_id"`
	CurrentValue float64   `json:"value"`
	SampledAt    time.Time `json:"sampled_at"`
	ManagedType  string    `json:"managed_type"`
	ManagedId    int64     `json:"managed_id"`
}

var months = []string{"_00", "_01", "_02", "_03", "_04", "_05", "_06", "_07", "_08", "_09", "_10", "_11", "_12", "_13"}

type resultScan interface {
	Scan(dest ...interface{}) error
}

func (self *server) findById(ctx *context, table *types.TableDefinition, projection string,
	scan func(rows resultScan) (interface{}, error), response http.ResponseWriter, request *http.Request) {
	paths := strings.Split(strings.Trim(request.URL.Path, "/"), "/")
	if 2 != len(paths) {
		http.Error(response, "404 page not found, path must is 'alerts' or 'histories', actual path is '"+
			request.URL.Path+"'", http.StatusNotFound)
		return
	}

	self.findOne(ctx, "select "+projection+" from "+table.CollectionName+" where id = "+paths[1], table, projection, scan, response, request)
}

func (self *server) findOne(ctx *context, sqlString string, table *types.TableDefinition, projection string,
	scan func(rows resultScan) (interface{}, error), response http.ResponseWriter, request *http.Request) {
	v, e := scan(ctx.db.QueryRow(sqlString))
	if nil != e {
		if e == sql.ErrNoRows {
			response.WriteHeader(http.StatusNotFound)
		} else {
			response.WriteHeader(http.StatusInternalServerError)
		}
		response.Write([]byte(e.Error()))
		return
	}

	response.WriteHeader(http.StatusOK)

	bs, _ := time.Now().MarshalJSON()
	response.Write([]byte(`{"created_at":`))
	response.Write(bs)
	response.Write([]byte(`, "value":`))
	encoder := json.NewEncoder(response)
	e = encoder.Encode(v)
	if nil != e {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(e.Error()))
	} else {
		response.Write([]byte(`}`))
	}
}

func (self *server) find(ctx *context, table *types.TableDefinition, projection string,
	scan func(rows resultScan) (interface{}, error), response http.ResponseWriter, request *http.Request) {
	where, params, e := ds.BuildWhereWithQueryParams(ctx.drv, table, 1, request.URL.Query())
	if nil != e {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(e.Error()))
		return
	}

	var rows *sql.Rows = nil
	if nil != params {
		rows, e = ctx.db.Query("select "+projection+" from "+table.CollectionName+where, params...)
	} else {
		rows, e = ctx.db.Query("select " + projection + " from " + table.CollectionName + where)
	}

	if nil != e {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(e.Error()))
		return
	}
	defer rows.Close()

	entities := make([]interface{}, 0, 100)
	for rows.Next() {
		if 10000 < len(entities) {
			response.WriteHeader(http.StatusRequestEntityTooLarge)
			response.Write([]byte("result is too large."))
			return
		}

		v, e := scan(rows) //TriggeredAt  time.Time `json:"triggered_at"`
		if nil != e {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(e.Error()))
			return
		}

		entities = append(entities, v)
	}

	e = rows.Err()
	if nil != e {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(e.Error()))
		return
	}

	response.WriteHeader(http.StatusOK)
	bs, _ := time.Now().MarshalJSON()
	response.Write([]byte(`{"created_at":`))
	response.Write(bs)
	response.Write([]byte(`, "value":`))
	if nil == entities || 0 == len(entities) {
		response.Write([]byte(`[]}`))
	} else {
		encoder := json.NewEncoder(response)
		e = encoder.Encode(entities)
		if nil != e {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(e.Error()))
		} else {
			response.Write([]byte(`}`))
		}
	}
}

func (self *server) removeById(ctx *context, table *types.TableDefinition, response http.ResponseWriter, request *http.Request) {
	paths := strings.Split(strings.Trim(request.URL.Path, "/"), "/")
	if 2 != len(paths) {
		http.Error(response, "404 page not found, path must is 'alerts' or 'histories', actual path is '"+
			request.URL.Path+"'", http.StatusNotFound)
		return
	}

	self.removeBySQL(ctx, "DELETE FROM "+table.CollectionName+" WHERE id = "+paths[1], table, response, request)
}

func (self *server) removeBySQL(ctx *context, sqlString string, table *types.TableDefinition, response http.ResponseWriter, request *http.Request) {
	res, e := ctx.db.Exec(sqlString)
	if nil != e {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(e.Error()))
		return
	}
	effected, e := res.RowsAffected()
	if nil != e {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(e.Error()))
		return
	}

	bs, _ := time.Now().MarshalJSON()
	response.Write([]byte(`{"created_at":`))
	response.Write(bs)
	response.Write([]byte(`, "effected":`))
	response.Write([]byte(strconv.FormatInt(effected, 10)))
	response.Write([]byte(`}`))
}

func (self *server) remove(ctx *context, table *types.TableDefinition, response http.ResponseWriter, request *http.Request) {
	query_params := make(map[string]string)
	for k, v := range request.URL.Query() {
		query_params[k] = v[len(v)-1]
	}
	where, params, e := ds.BuildWhere(ctx.drv, table, 1, query_params)
	if nil != e {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(e.Error()))
		return
	}

	var res sql.Result
	if nil != params {
		res, e = ctx.db.Exec("DELETE FROM "+table.CollectionName+where, params...)
	} else {
		res, e = ctx.db.Exec("DELETE FROM " + table.CollectionName + where)
	}

	if nil != e {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(e.Error()))
		return
	}
	effected, e := res.RowsAffected()
	if nil != e {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(e.Error()))
		return
	}

	bs, _ := time.Now().MarshalJSON()
	response.Write([]byte(`{"created_at":`))
	response.Write(bs)
	response.Write([]byte(`, "effected":`))
	response.Write([]byte(strconv.FormatInt(effected, 10)))
	response.Write([]byte(`}`))
}

func (self *server) count(ctx *context, table *types.TableDefinition, response http.ResponseWriter, request *http.Request) {
	query_params := make(map[string]string)
	for k, v := range request.URL.Query() {
		query_params[k] = v[len(v)-1]
	}
	where, params, e := ds.BuildWhere(ctx.drv, table, 1, query_params)
	if nil != e {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(e.Error()))
		return
	}

	count := int64(0)
	if nil != params {
		e = ctx.db.QueryRow("select count(*) from "+table.CollectionName+where, params...).Scan(&count)
	} else {
		e = ctx.db.QueryRow("select count(*) from " + table.CollectionName + where).Scan(&count)
	}
	if nil != e {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(e.Error()))
		return
	}
	bs, _ := time.Now().MarshalJSON()
	response.Write([]byte(`{"created_at":`))
	response.Write(bs)
	response.Write([]byte(`, "value":`))
	response.Write([]byte(strconv.FormatInt(count, 10)))
	response.Write([]byte(`}`))
}

const alert_prejection_sql = " id, action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at "
const cookies_prejection_sql = " id, action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, level, content, current_value, triggered_at "

func (self *server) findAlertCookies(ctx *context, response http.ResponseWriter, request *http.Request) {
	self.find(ctx, tpt_alert_cookies, cookies_prejection_sql,
		func(rows resultScan) (interface{}, error) {
			entity := &AlertEntity{}
			return entity, rows.Scan(
				&entity.Id,             //Id               int64     `json:"id"`
				&entity.ActionId,       //ActionId         int64     `json:"action_id"`
				&entity.ManagedType,    //ManagedType      string    `json:"managed_type"`
				&entity.ManagedId,      //ManagedId        int64     `json:"managed_id"`
				&entity.Status,         //Status           string    `json:"status"`
				&entity.PreviousStatus, //PreviousStatus   int64     `json:"previous_status"`
				&entity.EventId,        //EventId          string    `json:"event_id"`
				&entity.SequenceId,     //SequenceId       int64     `json:"sequence_id"`
				&entity.Level,          //Level            int64     `json:"level"`
				&entity.Content,        //Content          string    `json:"content"`
				&entity.CurrentValue,   //CurrentValue     string    `json:"current_value"`
				&entity.TriggeredAt)    //TriggeredAt      time.Time `json:"triggered_at"`
		}, response, request)
}
func (self *server) findAlertHistories(ctx *context, response http.ResponseWriter, request *http.Request) {
	self.find(ctx, tpt_alert_history, alert_prejection_sql,
		func(rows resultScan) (interface{}, error) {
			entity := &AlertEntity{}
			return entity, rows.Scan(
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
		}, response, request)
}

func (self *server) findAlertCookiesBy(ctx *context, response http.ResponseWriter, request *http.Request) {
	paths := strings.Split(strings.Trim(request.URL.Path, "/"), "/")
	if 2 != len(paths) {
		http.Error(response, "404 page not found, path must is 'alerts' or 'histories', actual path is '"+
			request.URL.Path+"'", http.StatusNotFound)
		return
	}
	scan := func(rows resultScan) (interface{}, error) {
		entity := &AlertEntity{}
		return entity, rows.Scan(
			&entity.Id,             //Id               int64     `json:"id"`
			&entity.ActionId,       //ActionId         int64     `json:"action_id"`
			&entity.ManagedType,    //ManagedType      string    `json:"managed_type"`
			&entity.ManagedId,      //ManagedId        int64     `json:"managed_id"`
			&entity.Status,         //Status           string    `json:"status"`
			&entity.PreviousStatus, //PreviousStatus   int64     `json:"previous_status"`
			&entity.EventId,        //EventId          string    `json:"event_id"`
			&entity.SequenceId,     //SequenceId       int64     `json:"sequence_id"`
			&entity.Level,          //Level            int64     `json:"level"`
			&entity.Content,        //Content          string    `json:"content"`
			&entity.CurrentValue,   //CurrentValue     string    `json:"current_value"`
			&entity.TriggeredAt)    //TriggeredAt      time.Time `json:"triggered_at"`
	}
	id := paths[1]

	if '@' == id[0] {
		self.findOne(ctx, "select "+cookies_prejection_sql+" from "+tpt_alert_cookies.CollectionName+" where action_id = "+id[1:], tpt_alert_cookies, cookies_prejection_sql,
			scan, response, request)
	} else {
		self.findOne(ctx, "select "+cookies_prejection_sql+" from "+tpt_alert_cookies.CollectionName+" where id = "+id, tpt_alert_cookies, cookies_prejection_sql,
			scan, response, request)
	}
}
func (self *server) findAlertHistoriesBy(ctx *context, response http.ResponseWriter, request *http.Request) {
	self.findById(ctx, tpt_alert_history, alert_prejection_sql,
		func(rows resultScan) (interface{}, error) {
			entity := &AlertEntity{}
			return entity, rows.Scan(
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
		}, response, request)
}

func (self *server) countAlertCookies(ctx *context, response http.ResponseWriter, request *http.Request) {
	self.count(ctx, tpt_alert_cookies, response, request)
}
func (self *server) countAlertHistories(ctx *context, response http.ResponseWriter, request *http.Request) {
	self.count(ctx, tpt_alert_history, response, request)
}
func (self *server) countHistories(ctx *context, response http.ResponseWriter, request *http.Request) {
	self.count(ctx, tpt_history, response, request)
}

func (self *server) removeAlertCookies(ctx *context, response http.ResponseWriter, request *http.Request) {
	self.remove(ctx, tpt_alert_cookies, response, request)
}
func (self *server) removeAlertHistories(ctx *context, response http.ResponseWriter, request *http.Request) {
	self.remove(ctx, tpt_alert_history, response, request)
}
func (self *server) removeHistories(ctx *context, response http.ResponseWriter, request *http.Request) {
	self.remove(ctx, tpt_history, response, request)
}

func (self *server) removeAlertCookiesBy(ctx *context, response http.ResponseWriter, request *http.Request) {
	paths := strings.Split(strings.Trim(request.URL.Path, "/"), "/")
	if 2 != len(paths) {
		http.Error(response, "404 page not found, path must is 'alerts' or 'histories', actual path is '"+
			request.URL.Path+"'", http.StatusNotFound)
		return
	}
	id := paths[1]

	if '@' == id[0] {
		self.removeBySQL(ctx, "DELETE FROM "+tpt_alert_cookies.CollectionName+" WHERE action_id = "+id[1:], tpt_alert_cookies, response, request)
	} else {
		self.removeBySQL(ctx, "DELETE FROM "+tpt_alert_cookies.CollectionName+" WHERE id = "+id, tpt_alert_cookies, response, request)
	}
}
func (self *server) removeAlertHistoriesBy(ctx *context, response http.ResponseWriter, request *http.Request) {
	self.removeById(ctx, tpt_alert_history, response, request)
}
func (self *server) removeHistoriesBy(ctx *context, response http.ResponseWriter, request *http.Request) {
	self.removeById(ctx, tpt_history, response, request)
}

func (self *server) findHistories(ctx *context, response http.ResponseWriter, request *http.Request) {
	self.find(ctx, tpt_history, " id, action_id, managed_type, managed_id, current_value, sampled_at ",
		func(rows resultScan) (interface{}, error) {
			entity := &HistoryEntity{}
			return entity, rows.Scan(
				&entity.Id,           //Id           int64     `json:"id"`
				&entity.ActionId,     //ActionId     int64     `json:"action_id"`
				&entity.ManagedType,  //ManagedType  string    `json:"managed_type"`
				&entity.ManagedId,    //ManagedId    int64     `json:"managed_id"`
				&entity.CurrentValue, //CurrentValue string    `json:"current_value"`
				&entity.SampledAt)    //TriggeredAt  time.Time `json:"triggered_at"`
		}, response, request)
}

func (self *server) findHistoriesBy(ctx *context, response http.ResponseWriter, request *http.Request) {
	self.findById(ctx, tpt_history, " id, action_id, managed_type, managed_id, current_value, sampled_at ",
		func(rows resultScan) (interface{}, error) {
			entity := &HistoryEntity{}
			return entity, rows.Scan(
				&entity.Id,           //Id           int64     `json:"id"`
				&entity.ActionId,     //ActionId     int64     `json:"action_id"`
				&entity.ManagedType,  //ManagedType  string    `json:"managed_type"`
				&entity.ManagedId,    //ManagedId    int64     `json:"managed_id"`
				&entity.CurrentValue, //CurrentValue string    `json:"current_value"`
				&entity.SampledAt)    //TriggeredAt  time.Time `json:"triggered_at"`
		}, response, request)
}

func (self *server) onAlerts(ctx *context, response http.ResponseWriter, request *http.Request) {
	var entities []AlertEntity
	decoder := json.NewDecoder(request.Body)
	decoder.UseNumber()
	e := decoder.Decode(&entities)
	if nil != e {
		response.WriteHeader(http.StatusBadRequest)
		io.WriteString(response, "it is not a valid json string, ")
		io.WriteString(response, e.Error())
		return
	}
	isCommited := false
	tx, e := ctx.db.Begin()
	if nil != e {
		response.WriteHeader(http.StatusInternalServerError)
		io.WriteString(response, "start a transcation failed, ")
		io.WriteString(response, e.Error())
		return
	}

	defer func() {
		if !isCommited {
			tx.Rollback()
		}
	}()

	now := time.Now()

	for _, entity := range entities {
		if 0 == entity.Status {
			e := ctx.dialect.removeAlertCookies(tx, &entity)
			if nil != e && sql.ErrNoRows != e {
				response.WriteHeader(http.StatusInternalServerError)
				io.WriteString(response, "remove cookies failed, ")
				io.WriteString(response, e.Error())
				return
			}
		} else {
			e := ctx.dialect.saveAlertCookies(tx, &entity)
			if nil != e {
				response.WriteHeader(http.StatusInternalServerError)
				io.WriteString(response, "save cookies failed, ")
				io.WriteString(response, e.Error())
				return
			}
		}

		//fmt.Println("save history with action_id was", entity.ActionId, " and status is", entity.Status)
		e = ctx.dialect.saveAlertHistory(tx, &entity)
		if nil != e {
			response.WriteHeader(http.StatusInternalServerError)
			io.WriteString(response, "save alert failed, ")
			io.WriteString(response, e.Error())
			return
		}

		if nil != entity.NotificationData {
			var queue, id interface{}
			if 0 != len(entity.NotificationData.Queue) {
				queue = entity.NotificationData.Queue
			}

			if 0 != len(entity.NotificationData.Id) {
				id = entity.NotificationData.Id
			}

			e = ctx.dialect.saveNotification(tx, queue, id, &entity, now)
			if nil != e {
				response.WriteHeader(http.StatusInternalServerError)
				io.WriteString(response, "save notification failed, ")
				io.WriteString(response, e.Error())
				return
			}
		}
	}

	isCommited = true
	e = tx.Commit()
	if nil != e && sql.ErrNoRows != e {
		response.WriteHeader(http.StatusInternalServerError)
		io.WriteString(response, "commit transcation failed, ")
		io.WriteString(response, e.Error())
		return
	}
	isCommited = true
}

func (self *server) onHistories(ctx *context, response http.ResponseWriter, request *http.Request) {
	var entities []HistoryEntity
	decoder := json.NewDecoder(request.Body)
	decoder.UseNumber()
	e := decoder.Decode(&entities)
	if nil != e {
		response.WriteHeader(http.StatusBadRequest)
		io.WriteString(response, "it is not a valid json string, ")
		io.WriteString(response, e.Error())
		return
	}
	isCommited := false
	tx, e := ctx.db.Begin()
	if nil != e {
		response.WriteHeader(http.StatusInternalServerError)
		io.WriteString(response, "start a transcation failed, ")
		io.WriteString(response, e.Error())
		return
	}
	defer func() {
		if !isCommited {
			tx.Rollback()
		}
	}()

	for _, entity := range entities {
		e := ctx.dialect.saveHistory(tx, &entity)
		if nil != e {
			response.WriteHeader(http.StatusInternalServerError)
			io.WriteString(response, "save history failed")
			io.WriteString(response, e.Error())
			return
		}
	}

	isCommited = true
	e = tx.Commit()
	if nil != e && sql.ErrNoRows != e {
		response.WriteHeader(http.StatusInternalServerError)
		io.WriteString(response, "commit transcation failed")
		io.WriteString(response, e.Error())
		return
	}
	isCommited = true
}

func expvarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}

func (self *server) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	defer func() {
		if nil != request.Body {
			request.Body.Close()
		}
	}()

	// fmt.Println(request.Method, request.URL.Path)

	var cb func(s *server, ctx *context, response http.ResponseWriter, request *http.Request) = nil
	switch request.Method {
	case "PUT":
		switch request.URL.Path {
		case "/alerts", "/alerts/":
			cb = (*server).onAlerts
		case "/histories", "/histories/":
			cb = (*server).onHistories
		default:
			http.Error(response, "404 page not found, path must is 'alerts' or 'history', actual path is '"+
				request.URL.Path+"'", http.StatusNotFound)
			return
		}
	case "GET":
		switch request.URL.Path {
		case "/alert_cookies", "/alert_cookies/":
			cb = (*server).findAlertCookies
		case "/alerts", "/alerts/":
			cb = (*server).findAlertHistories
		case "/alert_cookies/count", "/alert_cookies/count/":
			cb = (*server).countAlertCookies
		case "/alerts/count", "/alerts/count/":
			cb = (*server).countAlertHistories
		case "/histories", "/histories/":
			cb = (*server).findHistories
		case "/histories/count", "/histories/count/":
			cb = (*server).countHistories
		default:

			if strings.HasPrefix(request.URL.Path, "/debug/") {
				switch request.URL.Path {
				case "/debug/vars":
					expvarHandler(response, request)
				case "/debug/pprof", "/debug/pprof/":
					pprof.Index(response, request)
				case "/debug/pprof/cmdline":
					pprof.Cmdline(response, request)
				case "/debug/pprof/profile":
					pprof.Profile(response, request)
				case "/debug/pprof/symbol", "/debug/pprof/symbol/":
					pprof.Symbol(response, request)
				default:
					if strings.HasPrefix(request.URL.Path, "/debug/pprof/") {
						pprof.Index(response, request)
						return
					}
					http.NotFound(response, request)
				}
				return
			}

			paths := strings.Split(strings.Trim(request.URL.Path, "/"), "/")
			if 2 != len(paths) {
				http.Error(response, "404 page not found, path must is 'alerts' or 'histories', actual path is '"+
					request.URL.Path+"'", http.StatusNotFound)
				return
			}
			switch paths[0] {
			case "alert_cookies":
				cb = (*server).findAlertCookiesBy
			case "alerts":
				cb = (*server).findAlertHistoriesBy
			case "histories":
				cb = (*server).findHistoriesBy
			default:
				http.Error(response, "404 page not found, path must is 'alerts' or 'histories', actual path is '"+
					request.URL.Path+"'", http.StatusNotFound)
				return
			}
		}
	case "DELETE":
		switch request.URL.Path {
		case "/alert_cookies", "/alert_cookies/":
			cb = (*server).removeAlertCookies
		case "/alerts", "/alerts/":
			cb = (*server).removeAlertHistories
		case "/histories", "/histories/":
			cb = (*server).removeHistories
		default:
			paths := strings.Split(strings.Trim(request.URL.Path, "/"), "/")
			if 2 != len(paths) {
				http.Error(response, "404 page not found, path must is 'alerts' or 'histories', actual path is '"+
					request.URL.Path+"'", http.StatusNotFound)
				return
			}
			switch paths[0] {
			case "alert_cookies":
				cb = (*server).removeAlertCookiesBy
			case "alerts":
				cb = (*server).removeAlertHistoriesBy
			case "histories":
				cb = (*server).removeHistoriesBy
			default:
				http.Error(response, "404 page not found, path must is 'alerts' or 'histories', actual path is '"+
					request.URL.Path+"'", http.StatusNotFound)
				return
			}
		}
	default:
		http.Error(response, "404 page not found, method must is 'PUT'", http.StatusNotFound)
		return
	}

	c := make(chan error, 1)
	self.c <- &request_object{c: c, response: response, request: request, cb: cb}
	e := <-c
	if nil != e {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(e.Error()))
	}
}

func Main(is_test bool) error {
	flag.Parse()
	if nil != flag.Args() && 0 != len(flag.Args()) {
		flag.Usage()
		return nil
	}
	if e := commons.LoadConfig(nil); nil != e {
		return e
	}

	drv := *db_drv
	if strings.HasPrefix(drv, "odbc_with_") {
		drv = "odbc"
	}

	if 0 >= *goroutines {
		return errors.New("goroutines must is greate 0")
	}

	delayed_job_table1 := "INSERT INTO " + *delayed_job_table_name + "(priority, attempts, queue, handler, handler_id, last_error, run_at, locked_at, locked_by, failed_at, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NULL, $6, NULL, NULL, NULL, $7, $8)"
	delayed_job_table2 := "INSERT INTO " + *delayed_job_table_name + "(priority, attempts, queue, handler, handler_id, last_error, run_at, locked_at, locked_by, failed_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, NULL, ?, NULL, NULL, NULL, ?, ?)"
	var dialect Dialect = &GenSqlDialect{delayed_job_table: delayed_job_table2}
	switch *db_drv {
	case "mysql", "mymysql":
		dialect = &MySqlDialect{GenSqlDialect: GenSqlDialect{delayed_job_table: delayed_job_table2}}
	case "postgres":
		dialect = &PostgresqlDialect{delayed_job_table: delayed_job_table1}
	case "odbc_with_mssql":
		dialect = &MsSqlDialect{GenSqlDialect: GenSqlDialect{delayed_job_table: delayed_job_table2}}
	}

	if "sqlite3" == *db_drv {
		*goroutines = 1
	}

	var varString *expvar.String = nil
	varE := expvar.Get("carrier")
	if nil != varE {
		varString, _ = varE.(*expvar.String)
		if nil == varString {
			varString = expvar.NewString("carrier." + time.Now().String())
		}
	} else {
		varString = expvar.NewString("carrier")
	}

	srv := &server{c: make(chan *request_object, 3000),
		last_error: varString}

	contexts := make([]*context, 0, *goroutines)
	for i := 0; i < *goroutines; i++ {
		db, e := sql.Open(drv, *db_url)
		if nil != e {
			for _, conn := range contexts {
				conn.db.Close()
			}
			return errors.New("connect to db failed," + e.Error())
		}

		contexts = append(contexts, &context{drv: *db_drv, dialect: dialect, url: *db_url, db: db})
	}

	for i := 0; i < *goroutines; i++ {
		go srv.serve(i, contexts[i])
		srv.wait.Add(1)
	}

	server_instance = srv
	if is_test {
		log.Println("[carrier-test] serving at '" + *listenAddress + "'")
	} else {

		log.Println("[carrier] serving at '" + *listenAddress + "'")
		defer srv.Close()
		http.ListenAndServe(*listenAddress, srv)
	}
	return nil
}
