package ds

import (
	"commons"
	"commons/types"
	"database/sql"
	_ "expvar"
	"flag"
	"fmt"
	"github.com/runner-mei/go-restful"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	_ "runtime/pprof"
	"sync/atomic"
	"testing"
	"time"
)

var (
	models_file = flag.String("models", "etc/mj_models.xml", "the name of models file")
	dbUrl       = flag.String("dburl", "host=127.0.0.1 dbname=ds user=postgres password=mfk sslmode=disable", "the db url")
	drv         = flag.String("db", "postgres", "the db driver")
	goroutines  = flag.Int("connections", 10, "the db connection number")
	address     = flag.String("http", ":7071", "the address of http")

	//test_db    = flag.String("test.db", "sqlite3", "the db driver name for test")
	//test_dbUrl = flag.String("test.dburl", "test.sqlite3.db", "the db url")

	test_db                          = flag.String("test.db", "postgres", "the db driver name for test")
	test_dbUrl                       = flag.String("test.dburl", "host=127.0.0.1 dbname=test user=postgres password=mfk sslmode=disable", "the db url")
	test_address                     = flag.String("test.http", ":7071", "the address of http")
	is_test      int32               = 0
	srv_instance *server             = nil
	ws_instance  *restful.WebService = nil
)

func mainHandle(req *restful.Request, resp *restful.Response) {
	errFile := "_log_/error.html"
	_, err := os.Stat(errFile)
	if err == nil || os.IsExist(err) {

		http.ServeFile(
			resp.ResponseWriter,
			req.Request,
			errFile)

		return
	}
	resp.Write([]byte("Hello, World!"))
}

func Main() {
	flag.Parse()

	if nil != flag.Args() && 0 != len(flag.Args()) {
		flag.Usage()
		return
	}

	if !commons.FileExists(*models_file) {
		file := filepath.Join("..", *models_file)
		if commons.FileExists(file) {
			*models_file = file
		}
	}
	srv, e := newServer(*drv, *dbUrl, *models_file, *goroutines)
	if nil != e {
		fmt.Println(e)
		return
	}

	defer func() {
		if 1 == atomic.LoadInt32(&is_test) {
			srv_instance = srv
		} else {
			srv.Close()
		}
	}()

	ws := new(restful.WebService)
	ws.Route(ws.GET("/").To(mainHandle))

	ws.Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML) // you can specify this per route as well

	ws.Route(ws.GET("/{parent_type}/{parent_id}/children/{type}/{foreign_key}").To(srv.Children).
		Doc("get a object instance by parent id").
		Param(ws.PathParameter("parent_type", "type of the parent").DataType("string")).
		Param(ws.PathParameter("parent_id", "id of the parent").DataType("string")).
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.PathParameter("foreign_key", "foreign key of the parant").DataType("string"))) // on the response

	ws.Route(ws.POST("/{parent_type}/{parent_id}/children/{type}/{foreign_key}").To(srv.CreateByParent).
		Doc("create a object instance by parent id").
		Param(ws.PathParameter("parent_type", "type of the parent").DataType("string")).
		Param(ws.PathParameter("parent_id", "id of the parent").DataType("string")).
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.PathParameter("foreign_key", "foreign key of the parant").DataType("string"))) // on the response

	ws.Route(ws.GET("/{parent_type}/{parent_id}/children/{type}").To(srv.Children).
		Doc("get a object instance by parent id").
		Param(ws.PathParameter("parent_type", "type of the parent").DataType("string")).
		Param(ws.PathParameter("parent_id", "id of the parent").DataType("string")).
		Param(ws.PathParameter("type", "type of the instance").DataType("string"))) // on the response

	ws.Route(ws.POST("/{parent_type}/{parent_id}/children/{type}").To(srv.CreateByParent).
		Doc("create a object instance by parent id").
		Param(ws.PathParameter("parent_type", "type of the parent").DataType("string")).
		Param(ws.PathParameter("parent_id", "id of the parent").DataType("string")).
		Param(ws.PathParameter("type", "type of the instance").DataType("string"))) // on the response

	ws.Route(ws.GET("/{child_type}/{child_id}/parent/{type}/{foreign_key}").To(srv.Parent).
		Doc("get a object instance by child id").
		Param(ws.PathParameter("child_type", "type of the child").DataType("string")).
		Param(ws.PathParameter("child_id", "id of the child").DataType("string")).
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.PathParameter("foreign_key", "foreign key of the child").DataType("string"))) // on the response

	ws.Route(ws.GET("/{child_type}/{child_id}/parent/{type}").To(srv.Parent).
		Doc("get a object instance by child id").
		Param(ws.PathParameter("child_type", "type of the child").DataType("string")).
		Param(ws.PathParameter("child_id", "id of the child").DataType("string")).
		Param(ws.PathParameter("type", "type of the instance").DataType("string"))) // on the response

	ws.Route(ws.GET("/{type}/{id}").To(srv.FindById).
		Doc("get a object instance").
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.PathParameter("id", "identifier of the instance").DataType("string"))) // on the response

	ws.Route(ws.GET("/{type}").To(srv.FindByParams).
		Doc("get some object instances").
		Param(ws.PathParameter("type", "type of the instance").DataType("string"))) // on the response

	// ws.Route(ws.GET("/{type}/{id}/parent/{parent-type}/").To(srv.FindById).
	// 	Doc("get  a object instance").
	// 	Param(ws.PathParameter("type", "type of the instance").DataType("string")).
	// 	Param(ws.PathParameter("id", "identifier of the instance").DataType("string")).
	// 	Writes(User{})) // on the response

	// ws.Route(ws.GET("/{type}/{id}/children/{children-type}").To(srv.FindById).
	// 	Doc("get a object instance").
	// 	Param(ws.PathParameter("type", "type of the instance").DataType("string")).
	// 	Param(ws.PathParameter("id", "identifier of the instance").DataType("string")).
	// 	Writes(User{})) // on the response

	ws.Route(ws.POST("/{type}").To(srv.Create).
		Doc("create a object").
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.BodyParameter("object", "representation of a object instance").DataType("main.User"))) // from the request

	ws.Route(ws.PUT("/{type}/{id}").To(srv.UpdateById).
		Doc("update a object").
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.PathParameter("id", "identifier of the instance").DataType("string"))) // from the request

	ws.Route(ws.PUT("/{type}").To(srv.UpdateByParams).
		Doc("update some objects").
		Param(ws.PathParameter("type", "type of the instance").DataType("string"))) // from the request

	ws.Route(ws.DELETE("/{type}/{id}").To(srv.DeleteById).
		Doc("delete a object").
		Param(ws.PathParameter("type", "type of the instance").DataType("string")).
		Param(ws.PathParameter("id", "identifier of the instance").DataType("string")))

	ws.Route(ws.DELETE("/{type}").To(srv.DeleteByParams).
		Doc("delete some object").
		Param(ws.PathParameter("type", "type of the instance").DataType("string")))

	restful.Add(ws)

	if 1 == atomic.LoadInt32(&is_test) {
		ws_instance = ws
		//http.Handle("/debug/vars", http.HandlerFunc(expvarHandler))
		//http.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		//http.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		//for _, pf := range rpprof.Profiles() {
		//	http.Handle("/debug/pprof/"+pf.Name(), pprof.Handler(pf.Name()))
		//}
		//http.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	} else {
		log.Println("[ds] serving at '" + *address + "'")
		// mux := http.NewServeMux()
		// mux.Handle("/debug/vars", http.HandlerFunc(expvarHandler))
		// mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		// mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		// for _, pf := range rpprof.Profiles() {
		// 	mux.Handle("/debug/pprof/"+pf.Name(), pprof.Handler(pf.Name()))
		// }
		// mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		http.ListenAndServe(*address, nil)
	}
}

func testBase(t *testing.T, file string, init_cb func(drv string, conn *sql.DB), cb func(db *Client, definitions *types.TableDefinitions)) {
	definitions, err := types.LoadTableDefinitions(file)
	if nil != err {
		t.Errorf("read file '%s' failed, %s", file, err.Error())
		t.FailNow()
		return
	}
	conn, err := sql.Open(*test_db, *test_dbUrl)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer func() {
		if nil != conn {
			conn.Close()
		}
	}()

	init_cb(*test_db, conn)
	conn.Close()
	conn = nil

	*drv = *test_db
	*dbUrl = *test_dbUrl
	*address = *test_address
	*models_file = file
	atomic.StoreInt32(&is_test, 1)

	Main()
	defer restful.ClearRegisteredWebServices()
	var listener net.Listener = nil

	listener, e := net.Listen("tcp", *address)
	if nil != e {
		return
	}

	ch := make(chan string)
	go func() {
		defer func() {
			ch <- "exit"
		}()
		ch <- "ok"
		http.Serve(listener, nil)
	}()

	s := <-ch
	if "ok" != s {
		return
	}

	time.Sleep(10 * time.Microsecond)
	cb(NewClient("http://127.0.0.1"+*test_address), definitions)

	if nil != srv_instance {
		srv_instance.Close()
		srv_instance = nil
	}
	if nil != listener {
		listener.Close()
	}
	restful.ClearRegisteredWebServices()

	<-ch
}

func SrvTest(t *testing.T, file string, cb func(client *Client, definitions *types.TableDefinitions)) {
	testBase(t, file, func(drv string, conn *sql.DB) {
		sql_str := `
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS redis_commands;
DROP TABLE IF EXISTS actions;
DROP TABLE IF EXISTS metric_triggers;
DROP TABLE IF EXISTS triggers;
DROP TABLE IF EXISTS wbem_params;
DROP TABLE IF EXISTS ssh_params;
DROP TABLE IF EXISTS snmp_params;
DROP TABLE IF EXISTS endpoint_params;
DROP TABLE IF EXISTS access_params;
DROP TABLE IF EXISTS addresses;
DROP TABLE IF EXISTS interfaces;
DROP TABLE IF EXISTS links;
DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS managed_objects;
DROP TABLE IF EXISTS attributes;


CREATE TABLE devices (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
  name          varchar(250),
  description   varchar(2000),
  created_at    timestamp,
  updated_at    timestamp,

  address       varchar(250),
  manufacturer  integer,
  catalog       integer,
  oid           varchar(250),
  services      integer,
  location      varchar(2000),
  type          varchar(100)
);


CREATE TABLE links (
	id                      INTEGER PRIMARY KEY AUTOINCREMENT,
  name                    varchar(250),
  description             varchar(2000),
  created_at              timestamp,
  updated_at              timestamp,

  custom_speed_up         integer,
  custom_speed_down       integer,
  device1                 integer,
  ifIndex1                integer,
  device2                 integer,
  ifIndex2                integer,
  sampling_direct         integer
);

CREATE TABLE  interfaces (
	id                      INTEGER PRIMARY KEY AUTOINCREMENT,
  name                    varchar(250),
  description             varchar(2000),
  created_at              timestamp,
  updated_at              timestamp,
  ifIndex                 integer,
  ifDescr                 varchar(2000),
  ifType                  integer,
  ifMtu                   integer,
  ifSpeed                 integer,
  ifPhysAddress           varchar(50),
  device_id               integer
) ;

CREATE TABLE addresses (
	id                      INTEGER PRIMARY KEY AUTOINCREMENT,
  name                    varchar(250),
  description             varchar(2000),
  created_at              timestamp,
  updated_at              timestamp,
  address                 varchar(50),
  ifIndex                 integer,
  netmask                 varchar(50),
  bcastAddress            integer,
  reasmMaxSize            integer,
  device_id               integer
);



CREATE TABLE snmp_params (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  managed_object_id  integer,

  port               integer,
  version            varchar(50),


  read_community VARCHAR(50),
  write_community VARCHAR(50),

  sec_model VARCHAR(50),    -- usm
  read_sec_name VARCHAR(50),
  read_auth_pass VARCHAR(50),
  read_priv_pass VARCHAR(50),

  write_sec_name VARCHAR(50),
  write_auth_pass VARCHAR(50),
  write_priv_pass VARCHAR(50),

  max_msg_size INTEGER,
  context_name VARCHAR(50),
  identifier VARCHAR(50),
  engine_id VARCHAR(50)
) ;

CREATE TABLE ssh_params (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  managed_object_id  integer,
  description        varchar(2000),

  address            varchar(50),
  port               integer,
  user_name          varchar(50),
  user_password      varchar(250)
);

CREATE TABLE wbem_params (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  managed_object_id  integer,
  description        varchar(2000),
  url                varchar(2000),
  user_name          varchar(50),
  user_password      varchar(250)
) ;


CREATE TABLE triggers (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  name          varchar(250),
  expression    varchar(250),
  attachment    varchar(2000),
  description   varchar(2000),

  parent_type   varchar(250),
  parent_id     integer
);

CREATE TABLE metric_triggers (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  name               varchar(250),
  expression         varchar(250),
  attachment         varchar(2000),
  description        varchar(2000),

  parent_type        varchar(250),
  parent_id          integer,
  metric             varchar(250),
  managed_object_id  integer
) ;

CREATE TABLE actions (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  name               varchar(250),
  description        varchar(2000),
  parent_type        varchar(250),
  parent_id          integer
);

CREATE TABLE redis_commands (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  name               varchar(250),
  description        varchar(2000),
  parent_type        varchar(250),
  parent_id          integer,

  command            varchar(10),
  arg0               varchar(200),
  arg1               varchar(200),
  arg2               varchar(200),
  arg3               varchar(200),
  arg4               varchar(200)
);


CREATE TABLE alerts (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  name               varchar(250),
  description        varchar(2000),
  parent_type        varchar(250),
  parent_id          integer,

  max_repeated       integer,
  expression_style   varchar(50),
  expression_code    varchar(2000)
);


DROP TABLE IF EXISTS documents;
create table documents (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  name               varchar(100),
  type               varchar(100), 
  page_count         integer, 
  author             varchar(100), 
  bytes              integer,
  journal_id         integer,
  isbn               varchar(100),
  compressed_format  varchar(10), 
  website_id         integer, 
  user_id            integer,
  printer_id         integer,
  publish_at         integer
);

DROP TABLE IF EXISTS websites;
CREATE TABLE websites (id  INTEGER PRIMARY KEY AUTOINCREMENT, url varchar(200));

DROP TABLE IF EXISTS printers;
CREATE TABLE printers (id  INTEGER PRIMARY KEY AUTOINCREMENT, name varchar(200));

DROP TABLE IF EXISTS topics;
CREATE TABLE topics (id  INTEGER PRIMARY KEY AUTOINCREMENT, name varchar(200));

-- tables for CLOB
DROP TABLE IF EXISTS zip_files;
CREATE TABLE zip_files (id  INTEGER PRIMARY KEY AUTOINCREMENT, body text, document_id integer);
`

		sql_file := drv + "_test.sql"
		if "postgres" == drv && *IsPostgresqlInherit {
			sql_file = drv + "_inherit_test.sql"
		}

		if !commons.FileExists(sql_file) {
			file := "../ds/" + drv + "_test.sql"
			if "postgres" == drv && *IsPostgresqlInherit {
				file = "../ds/" + drv + "_inherit_test.sql"
			}

			if commons.FileExists(file) {
				sql_file = file
			}
		}
		if r, err := os.Open(sql_file); nil == err {
			all, err := ioutil.ReadAll(r)
			if nil != err {
				t.Fatal(err)
				t.FailNow()
				return
			}
			sql_str = string(all)
			t.Log("load " + sql_file)
		}

		_, err := conn.Exec(sql_str)
		if err != nil {
			t.Fatal(err)
			t.FailNow()
			return
		}

		if "sqlite3" == drv {
			_, err = conn.Exec(`CREATE TABLE IF NOT EXISTS alerts (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  name               varchar(250),
  description        varchar(2000),
  parent_type        varchar(250),
  parent_id          integer,

  max_repeated       integer,
  expression_style   varchar(50),
  expression_code    varchar(2000)
);`)
			if err != nil {
				t.Fatal(err)
				t.FailNow()
				return
			}
		}

	}, cb)
}

func createJson(t *testing.T, client *Client, target, msg string) string {
	_, id, e := client.CreateJson("http://127.0.0.1:7071/"+target, []byte(msg))
	if nil != e {
		t.Errorf("create %s failed, %v", target, e)
		t.FailNow()
	}
	if nil != client.Warnings {
		t.Error(client.Warnings)
	}
	return id
}

func CreateItForTest(t *testing.T, client *Client, target string, values map[string]interface{}) string {
	id, e := client.Create(target, values)
	if nil != e {
		t.Errorf("create %s failed, %v", target, e)
		t.FailNow()
	}
	if nil != client.Warnings {
		t.Error(client.Warnings)
	}
	return id
}

func CreateItByParentForTest(t *testing.T, client *Client, parnet_type, parent_id, target string, values map[string]interface{}) string {
	id, e := client.CreateByParent(parnet_type, parent_id, target, values)
	if nil != e {
		t.Errorf("create %s failed, %v", target, e)
		t.FailNow()
	}
	if nil != client.Warnings {
		t.Error(client.Warnings)
	}
	return id
}

func CreateMockDeviceForTest(t *testing.T, client *Client, factor string) string {
	return createJson(t, client, "device", fmt.Sprintf(`{"name":"dd%s", "type":"device", "address":"192.168.1.%s", "catalog":%s, "services":2%s, "managed_address":"20.0.8.110"}`, factor, factor, factor, factor))
}

func CreateMockSnmpParamsForTest(t *testing.T, client *Client, community string) string {
	return createJson(t, client, "snmp_param", fmt.Sprintf(`{ "version":"snmp_v2c", "read_community":"%s"}`, community))
}
