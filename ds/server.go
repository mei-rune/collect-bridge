package ds

import (
	"commons"
	"bytes"
	"database/sql"
	"log"
	"runtime"
)

type server struct {
	drv             string
	dbUrl           string
	goroutines      int
	isNumericParams bool

	ch chan func(srv *server, db *sql.DB) bool
}

func IsNumericParams(drv string) bool {
	switch drv {
	case "postgres":
		return true
	default:
		return false
	}
}

func newServer(drv, dbUrl string, goroutines int) (*server, error) {
	srv := &server{drv: drv,
		dbUrl:           dbUrl,
		goroutines:      goroutines,
		isNumericParams: IsNumericParams(drv),
		ch:              make(chan func(srv *server, db *sql.DB) bool)}

	conns := make([]sql.DB, 0, goroutines)
	for i := 0; i < srv.goroutines; i++ {
		db, e := sql.Open(self.drv, self.dbUrl)
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

	return srv
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
	}()

	for {
		f := <-self.ch
		if nil == f {
			break
		}
		if !f(self, db) {
			break
		}
	}
}

func (self *server) run(req *restful.Request, resp *restful.Response) bool) {
  ch := make(chan commons.Result)
  self.ch <- func(srv *server, db *sql.DB) bool {



    return true
  }
}