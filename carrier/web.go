package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"expvar"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var (
	listenAddress = flag.String("carrier.listen", ":7076", "the address of http")
	dbUrl         = flag.String("carrier.dburl", "host=127.0.0.1 dbname=tpt_extreme user=tpt password=extreme sslmode=disable", "the db url")
	drv           = flag.String("carrier.db", "postgres", "the db driver")
	goroutines    = flag.Int("carrier.connections", 10, "the db connection number")

	server_instance *server = nil
	is_test                 = false
)

type data_object struct {
	c        chan error
	request  *http.Request
	response http.ResponseWriter
	cb       func(s *server, ctx *context, response http.ResponseWriter, request *http.Request)
}

type server struct {
	c          chan *data_object
	wait       sync.WaitGroup
	last_error *expvar.String
}

type context struct {
	db  *sql.DB
	drv string
	url string
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

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for is_running {
		select {
		case <-ticker.C:
			fmt.Println("tick")
		case req, ok := <-self.c:
			if !ok {
				is_running = false
			} else {
				self.call(ctx, req)
			}
		}
	}
}

func (self *server) call(ctx *context, obj *data_object) {
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

func (self *server) run(ctx *context, obj *data_object) {
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

type alertEntity struct {
	RuleId       int64     `json:"action_id"`
	Status       string    `json:"status"`
	CurrentValue string    `json:"current_value"`
	TriggeredAt  time.Time `json:"triggered_at"`
	ManagedType  string    `json:"managed_type"`
	ManagedId    int64     `json:"managed_id"`
}

type historyEntity struct {
	RuleId       int64     `json:"action_id"`
	CurrentValue string    `json:"current_value"`
	SampledAt    time.Time `json:"sampled_at"`
	ManagedType  string    `json:"managed_type"`
	ManagedId    int64     `json:"managed_id"`
}

var months = []string{"_00", "_01", "_02", "_03", "_04", "_05", "_06", "_07", "_08", "_09", "_10", "_11", "_12", "_13"}

func (self *server) onAlerts(ctx *context, response http.ResponseWriter, request *http.Request) {
	var entities []alertEntity
	decoder := json.NewDecoder(request.Body)
	decoder.UseNumber()
	e := decoder.Decode(&entities)
	if nil != e {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte("it is not a valid json string, "))
		response.Write([]byte(e.Error()))
		return
	}
	isCommited := false
	tx, e := ctx.db.Begin()
	if nil != e {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(e.Error()))
		return
	}

	defer func() {
		if !isCommited {
			tx.Rollback()
		}
	}()

	for _, entity := range entities {
		var id int64
		e := tx.QueryRow("SELECT id FROM tpt_alert_cookies WHERE rule_id = $1", entity.RuleId).Scan(&id)
		if nil != e {
			if sql.ErrNoRows != e {
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte(e.Error()))
				return
			}

			_, e = tx.Exec(`INSERT INTO tpt_alert_cookies(rule_id, managed_type, managed_id, status, current_value, triggered_at)
    		VALUES ($1, $2, $3, $4, $5, $6)`, entity.RuleId, entity.ManagedType, entity.ManagedId,
				entity.Status, entity.CurrentValue, entity.TriggeredAt)
		} else {
			_, e = tx.Exec(`UPDATE tpt_alert_cookies SET status = $1, current_value = $2, triggered_at = $3  WHERE id = $4`,
				entity.Status, entity.CurrentValue, entity.TriggeredAt, id)
		}
		if nil != e {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(e.Error()))
			return
		}

		_, e = tx.Exec("INSERT INTO tpt_alert_history_"+strconv.Itoa(entity.TriggeredAt.Year())+months[entity.TriggeredAt.Month()]+
			"(rule_id, managed_type, managed_id, status, current_value, triggered_at) VALUES ($1, $2, $3, $4, $5, $6)",
			entity.RuleId, entity.ManagedType, entity.ManagedId, entity.Status, entity.CurrentValue, entity.TriggeredAt)
		if nil != e {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(e.Error()))
			return
		}
	}

	isCommited = true
	e = tx.Commit()
	if nil != e && sql.ErrNoRows != e {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(e.Error()))
		return
	}
	isCommited = true
}

func (self *server) onHistory(ctx *context, response http.ResponseWriter, request *http.Request) {
	var entities []historyEntity
	decoder := json.NewDecoder(request.Body)
	decoder.UseNumber()
	e := decoder.Decode(&entities)
	if nil != e {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte("it is not a valid json string, "))
		response.Write([]byte(e.Error()))
		return
	}
	isCommited := false
	tx, e := ctx.db.Begin()
	if nil != e {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(e.Error()))
		return
	}
	defer func() {
		if !isCommited {
			tx.Rollback()
		}
	}()

	for _, entity := range entities {
		_, e = tx.Exec("INSERT INTO tpt_alert_history_"+strconv.Itoa(entity.SampledAt.Year())+months[entity.SampledAt.Month()]+
			"(rule_id, managed_type, managed_id, current_value, sampled_at) VALUES ($1, $2, $3, $4, $5, $6)",
			entity.RuleId, entity.ManagedType, entity.ManagedId, entity.CurrentValue, entity.SampledAt)
		if nil != e {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(e.Error()))
			return
		}
	}

	isCommited = true
	e = tx.Commit()
	if nil != e && sql.ErrNoRows != e {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(e.Error()))
		return
	}
	isCommited = true
}

func (self *server) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	defer func() {
		if nil != request.Body {
			request.Body.Close()
		}
	}()

	if "PUT" != request.Method {
		http.Error(response, "404 page not found, method must is 'PUT'", http.StatusNotFound)
		return
	}

	var cb func(s *server, ctx *context, response http.ResponseWriter, request *http.Request) = nil
	switch request.URL.Path {
	case "alerts":
		cb = (*server).onAlerts
	case "history":
		cb = (*server).onHistory
	default:
		http.Error(response, "404 page not found, path must is 'alerts' or 'history', actual path is '"+
			request.URL.Path+"'", http.StatusNotFound)
		return
	}

	c := make(chan error, 1)
	self.c <- &data_object{c: c, response: response, request: request, cb: cb}
	e := <-c
	if nil != e {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(e.Error()))
	}
}

func main() {

	if 0 >= *goroutines {
		fmt.Println("goroutines must is greate 0")
		return
	}
	if "sqlite3" == *drv {
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

	srv := &server{c: make(chan *data_object, 3000),
		last_error: varString}

	contexts := make([]*context, 0, *goroutines)
	for i := 0; i < *goroutines; i++ {
		db, e := sql.Open(*drv, *dbUrl)
		if nil != e {
			for _, conn := range contexts {
				conn.db.Close()
			}
			fmt.Println("connect to db failed,", e)
			return
		}

		contexts = append(contexts, &context{drv: *drv, url: *dbUrl, db: db})
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
}
