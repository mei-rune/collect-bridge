package ds

import (
	"bytes"
	"commons"
	"commons/types"
	"database/sql"
	"fmt"
	"github.com/runner-mei/go-restful"
	"log"
	"runtime"
	"sync/atomic"
)

type server struct {
	drv          string
	dbUrl        string
	goroutines   int
	definitions  *types.TableDefinitions
	activedCount int32

	ch chan func(srv *server, db *session) bool
}

func NewServer(drv, dbUrl, file string, goroutines int) (*server, error) {
	definitions, e := types.LoadTableDefinitions(file)
	if nil != e {
		return nil, fmt.Errorf("read file '%s' failed, %s", file, e.Error())
	}

	if !is_test {
		log.Println("[ds] load '" + file + "'")
	}

	if 0 >= goroutines {
		return nil, fmt.Errorf("goroutines must is greate 0")
	}

	srv := &server{drv: drv,
		dbUrl:        dbUrl,
		goroutines:   goroutines,
		definitions:  definitions,
		activedCount: 0,
		ch:           make(chan func(srv *server, db *session) bool)}

	conns := make([]*sql.DB, 0, goroutines)
	for i := 0; i < srv.goroutines; i++ {
		db, e := sql.Open(srv.drv, srv.dbUrl)
		if nil != e {
			for _, conn := range conns {
				conn.Close()
			}
			return nil, e
		}

		conns = append(conns, db)
	}

	for _, conn := range conns {
		go srv.run(conn)
	}

	return srv, nil
}

func (self *server) Close() {
	for i := 0; i < self.goroutines; i++ {
		self.ch <- nil
	}
}
func (self *server) run(db *sql.DB) {
	defer func() {
		if e := recover(); nil != e {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("[panic] crashed with error - %s\r\n", e))
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}
			log.Println(buffer.String())
		}

		db.Close()
		atomic.AddInt32(&self.activedCount, -1)
	}()
	atomic.AddInt32(&self.activedCount, 1)

	sess := newSession(self.drv, db, self.definitions)

	for {
		f := <-self.ch
		if nil == f {
			break
		}
		if !f(self, sess) {
			break
		}
	}
	if !is_test {
		log.Println("server exit")
	}
}

func (self *server) call(req *restful.Request,
	resp *restful.Response,
	cb func(srv *server, db *session) commons.Result) {
	if 0 >= atomic.LoadInt32(&self.activedCount) {
		resp.WriteErrorString(commons.InternalErrorCode, "SERVER CLOSED")
		return
	}

	result_ch := make(chan commons.Result)
	defer close(result_ch)

	self.ch <- func(srv *server, db *session) bool {
		defer func() {
			if e := recover(); nil != e {
				var buffer bytes.Buffer
				buffer.WriteString(fmt.Sprintf("[panic] crashed with error - %s\r\n", e))
				for i := 1; ; i += 1 {
					_, file, line, ok := runtime.Caller(i)
					if !ok {
						break
					}
					buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
				}
				result_ch <- commons.ReturnError(commons.InternalErrorCode, buffer.String())
			}
		}()
		result_ch <- cb(srv, db)
		return true
	}

	res := <-result_ch
	if res.HasError() {
		resp.WriteErrorString(res.ErrorCode(), res.ErrorMessage())
	} else {
		if -1 != res.LastInsertId() {
			resp.WriteHeader(commons.CreatedCode)
		}
		resp.WriteEntity(res)
	}
}

func (self *server) FindById(req *restful.Request, resp *restful.Response) {
	self.call(req, resp, func(srv *server, db *session) commons.Result {
		t := req.PathParameter("type")
		if 0 == len(t) {
			return commons.ReturnError(commons.IsRequiredCode, "'type' is required.")
		}
		table := self.definitions.FindByUnderscoreName(t)
		if nil == table {
			return commons.ReturnError(commons.BadRequestCode, "table '"+t+"' is not exists.")
		}

		t = req.PathParameter("id")
		if 0 == len(t) {
			return commons.ReturnError(commons.IsRequiredCode, "'id' is required.")
		}

		if "@count" == t {
			params := make(map[string]string)
			for k, v := range req.Request.URL.Query() {
				params[k] = v[len(v)-1]
			}
			res, e := db.count(table, params)
			if nil != e {
				return commons.ReturnError(commons.InternalErrorCode, e.Error())
			}
			return commons.Return(res)
		}

		id, e := table.Id.Type.Parse(t)
		if nil != e {
			return commons.ReturnError(commons.BadRequestCode, fmt.Sprintf("'id' is not a '%v', actual value is '%v'",
				table.Id.Type.Name(), t))
		}

		res, e := db.findById(table, id, req.QueryParameter("includes"))
		if nil != e {
			if sql.ErrNoRows == e {
				return commons.ReturnError(commons.NotFoundCode, e.Error())
			}
			return commons.ReturnError(commons.InternalErrorCode, e.Error())
		} else {
			return commons.Return(res)
		}
	})
}

func (self *server) FindByParams(req *restful.Request, resp *restful.Response) {
	self.call(req, resp, func(srv *server, db *session) commons.Result {
		t := req.PathParameter("type")
		if 0 == len(t) {
			return commons.ReturnError(commons.IsRequiredCode, "'type' is required.")
		}
		defintion := self.definitions.FindByUnderscoreName(t)
		if nil == defintion {
			return commons.ReturnError(commons.BadRequestCode, "table '"+t+"' is not exists.")
		}

		params := make(map[string]string)
		for k, v := range req.Request.URL.Query() {
			params[k] = v[len(v)-1]
		}

		res, e := db.find(defintion, params)
		if nil != e {
			return commons.ReturnError(commons.InternalErrorCode, e.Error())
		}

		return commons.Return(res)
	})
}

func (self *server) Children(req *restful.Request, resp *restful.Response) {
	self.call(req, resp, func(srv *server, db *session) commons.Result {
		t := req.PathParameter("parent_type")
		if 0 == len(t) {
			return commons.ReturnError(commons.IsRequiredCode, "'parent_type' is required.")
		}
		parent_type := self.definitions.FindByUnderscoreName(t)
		if nil == parent_type {
			return commons.ReturnError(commons.BadRequestCode, "table '"+t+"' is not exists.")
		}

		t = req.PathParameter("parent_id")
		if 0 == len(t) {
			return commons.ReturnError(commons.IsRequiredCode, "'parent_id' is required.")
		}

		parent_id, e := parent_type.Id.Type.Parse(t)
		if nil != e {
			return commons.ReturnError(commons.BadRequestCode,
				fmt.Sprintf("'parent_id' is not a '%v', actual value is '%v'",
					parent_type.Id.Type.Name(), t))
		}

		t = req.PathParameter("type")
		if 0 == len(t) {
			return commons.ReturnError(commons.IsRequiredCode, "'type' is required.")
		}
		defintion := self.definitions.FindByUnderscoreName(t)
		if nil == defintion {
			return commons.ReturnError(commons.BadRequestCode, "table '"+t+"' is not exists.")
		}

		res, e := db.children(parent_type, parent_id, defintion, req.PathParameter("foreign_key"))
		if nil != e {
			return commons.ReturnError(commons.InternalErrorCode, e.Error())
		}

		return commons.Return(res)
	})
}

func (self *server) Parent(req *restful.Request, resp *restful.Response) {
	self.call(req, resp, func(srv *server, db *session) commons.Result {
		t := req.PathParameter("child_type")
		if 0 == len(t) {
			return commons.ReturnError(commons.IsRequiredCode, "'child_type' is required.")
		}
		child_type := self.definitions.FindByUnderscoreName(t)
		if nil == child_type {
			return commons.ReturnError(commons.BadRequestCode, "table '"+t+"' is not exists.")
		}

		t = req.PathParameter("child_id")
		if 0 == len(t) {
			return commons.ReturnError(commons.IsRequiredCode, "'child_id' is required.")
		}

		child_id, e := child_type.Id.Type.Parse(t)
		if nil != e {
			return commons.ReturnError(commons.BadRequestCode,
				fmt.Sprintf("'child_id' is not a '%v', actual value is '%v'",
					child_type.Id.Type.Name(), t))
		}

		t = req.PathParameter("type")
		if 0 == len(t) {
			return commons.ReturnError(commons.IsRequiredCode, "'type' is required.")
		}

		defintion := self.definitions.FindByUnderscoreName(t)
		if nil == defintion {
			return commons.ReturnError(commons.BadRequestCode, "table '"+t+"' is not exists.")
		}

		res, e := db.parent(child_type, child_id, defintion, req.PathParameter("foreign_key"))
		if nil != e {
			return commons.ReturnError(commons.InternalErrorCode, e.Error())
		}

		return commons.Return(res)
	})
}

func (self *server) UpdateById(req *restful.Request, resp *restful.Response) {
	self.call(req, resp, func(srv *server, db *session) commons.Result {
		t := req.PathParameter("type")
		if 0 == len(t) {
			return commons.ReturnError(commons.IsRequiredCode, "'type' is required.")
		}
		table := self.definitions.FindByUnderscoreName(t)
		if nil == table {
			return commons.ReturnError(commons.BadRequestCode, "table '"+t+"' is not exists.")
		}

		t = req.PathParameter("id")
		if 0 == len(t) {
			return commons.ReturnError(commons.IsRequiredCode, "'id' is required.")
		}

		id, e := table.Id.Type.Parse(t)
		if nil != e {
			return commons.ReturnError(commons.BadRequestCode,
				fmt.Sprintf("'id' is not a '%v', actual value is '%v'",
					table.Id.Type.Name(), t))
		}

		var attributes map[string]interface{}
		e = req.ReadEntity(&attributes)
		if nil != e {
			return commons.ReturnError(commons.BadRequestCode, "read body failed - "+e.Error())
		}

		e = db.updateById(table, id, attributes)
		if nil != e {
			if sql.ErrNoRows == e {
				return commons.ReturnError(commons.NotFoundCode, e.Error())
			}
			return commons.ReturnError(commons.InternalErrorCode, e.Error())
		} else {
			return commons.Return(true).SetEffected(1)
		}
	})
}

func (self *server) UpdateByParams(req *restful.Request, resp *restful.Response) {
	self.call(req, resp, func(srv *server, db *session) commons.Result {
		t := req.PathParameter("type")
		if 0 == len(t) {
			return commons.ReturnError(commons.IsRequiredCode, "'type' is required.")
		}
		defintion := self.definitions.FindByUnderscoreName(t)
		if nil == defintion {
			return commons.ReturnError(commons.BadRequestCode, "table '"+t+"' is not exists.")
		}
		var attributes map[string]interface{}
		e := req.ReadEntity(&attributes)
		if nil != e {
			return commons.ReturnError(commons.BadRequestCode, "read body failed - "+e.Error())
		}
		params := make(map[string]string)
		for k, v := range req.Request.URL.Query() {
			params[k] = v[len(v)-1]
		}
		affected, e := db.update(defintion, params, attributes)
		if nil != e {
			return commons.ReturnError(commons.InternalErrorCode, e.Error())
		} else {
			return commons.Return(true).SetEffected(affected)
		}
	})
}

func (self *server) DeleteById(req *restful.Request, resp *restful.Response) {
	self.call(req, resp, func(srv *server, db *session) commons.Result {
		t := req.PathParameter("type")
		if 0 == len(t) {
			return commons.ReturnError(commons.IsRequiredCode, "'type' is required.")
		}
		table := self.definitions.FindByUnderscoreName(t)
		if nil == table {
			return commons.ReturnError(commons.BadRequestCode, "table '"+t+"' is not exists.")
		}

		t = req.PathParameter("id")
		if 0 == len(t) {
			return commons.ReturnError(commons.IsRequiredCode, "'id' is required.")
		}

		id, e := table.Id.Type.Parse(t)
		if nil != e {
			return commons.ReturnError(commons.BadRequestCode,
				fmt.Sprintf("'id' is not a '%v', actual value is '%v'",
					table.Id.Type.Name(), t))
		}

		e = db.deleteById(table, id)
		if nil != e {
			if sql.ErrNoRows == e {
				return commons.ReturnError(commons.NotFoundCode, e.Error())
			}
			return commons.ReturnError(commons.InternalErrorCode, e.Error())
		} else {
			return commons.Return(true).SetEffected(1)
		}
	})
}

func (self *server) DeleteByParams(req *restful.Request, resp *restful.Response) {
	self.call(req, resp, func(srv *server, db *session) commons.Result {
		t := req.PathParameter("type")
		if 0 == len(t) {
			return commons.ReturnError(commons.IsRequiredCode, "'type' is required.")
		}
		defintion := self.definitions.FindByUnderscoreName(t)
		if nil == defintion {
			return commons.ReturnError(commons.BadRequestCode, "table '"+t+"' is not exists.")
		}

		params := make(map[string]string)
		for k, v := range req.Request.URL.Query() {
			params[k] = v[len(v)-1]
		}
		affected, e := db.delete(defintion, params)
		if nil != e {
			return commons.ReturnError(commons.InternalErrorCode, e.Error())
		} else {
			return commons.Return(true).SetEffected(affected)
		}
	})
}

func (self *server) Create(req *restful.Request, resp *restful.Response) {
	self.call(req, resp, func(srv *server, db *session) commons.Result {
		t := req.PathParameter("type")
		if 0 == len(t) {
			return commons.ReturnError(commons.IsRequiredCode, "'type' is required.")
		}
		defintion := self.definitions.FindByUnderscoreName(t)
		if nil == defintion {
			return commons.ReturnError(commons.BadRequestCode, "table '"+t+"' is not exists.")
		}

		var attributes map[string]interface{}
		e := req.ReadEntity(&attributes)
		if nil != e {
			return commons.ReturnError(commons.BadRequestCode, "read body failed - "+e.Error())
		}

		if "true" == req.QueryParameter("save") {
			params := make(map[string]string)
			for k, v := range req.Request.URL.Query() {
				params[k] = v[len(v)-1]
			}
			action, lastInsertId, e := db.save(defintion, params, attributes)

			var res *commons.SimpleResult
			if nil != e {
				res = commons.ReturnError(commons.InternalErrorCode, e.Error())
			} else {
				res = commons.Return(true).SetLastInsertId(lastInsertId)
			}
			switch action {
			case 0:
				res.SetOption("is_created", false)
			case 1:
				res.SetOption("is_created", true)
			}
			return res
		} else {
			lastInsertId, e := db.insert(defintion, attributes)
			if nil != e {
				return commons.ReturnError(commons.InternalErrorCode, e.Error())
			} else {
				return commons.Return(true).SetLastInsertId(lastInsertId)
			}
		}
	})
}
