package data_store

import (
	"commons"
	"commons/types"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"github.com/runner-mei/go-restful"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	pprof "net/http/pprof"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

var (
	models_file = flag.String("ds.models", "etc/tpt_models.xml", "the name of models file")
	db_url      = flag.String("db.url", "host=127.0.0.1 dbname=tpt_extreme user=tpt password=extreme sslmode=disable", "the db url")
	db_drv      = flag.String("db.driver", "postgres", "the db driver")
	goroutines  = flag.Int("db.connections", 10, "the db connection number")
	address     = flag.String("ds.listen", ":7071", "the address of http")

	Container    *restful.Container = restful.DefaultContainer
	is_test      int32              = 0
	srv_instance *server            = nil
)

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

	files := []string{*models_file,
		filepath.Join("..", *models_file),
		"conf/tpt_models.xml",
		"../conf/tpt_models.xml",
		"lib/models/tpt_models.xml",
		"../lib/models/tpt_models.xml"}
	found := false
	for _, file := range files {
		if commons.FileExists(file) {
			*models_file = file
			found = true
			break
		}
	}

	if !found {
		fmt.Println("models file is not exists, search path is:")
		for _, file := range files {
			fmt.Println("    ", file)
		}
		return
	}

	srv, e := newServer(*db_drv, *db_url, *models_file, *goroutines)
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

	restful.DefaultResponseMimeType = restful.MIME_JSON
	ws := new(restful.WebService)
	ws.Route(ws.GET("/").To(mainHandle))

	ws.Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON) // you can specify this per route as well

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

	Container.Add(ws)

	if restful.DefaultContainer != Container {
		Container.Handle("/debug/vars", http.HandlerFunc(expvarHandler))
		Container.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		Container.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		Container.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		Container.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	}

	if 1 == atomic.LoadInt32(&is_test) {
		log.Println("[ds-test] serving at '" + *address + "'")
	} else {
		log.Println("[ds] serving at '" + *address + "'")
		http.ListenAndServe(*address, nil)
	}
}

func testBase(t *testing.T, file string, init_cb func(drv string, conn *sql.DB), cb func(db *Client, definitions *types.TableDefinitions)) {
	definitions, err := types.LoadTableDefinitions(file)
	if nil != err {
		t.Errorf("read file '%s' failed, %s", file, err.Error())
		return
	}
	conn, err := sql.Open(*db_drv, *db_url)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer func() {
		if nil != conn {
			conn.Close()
		}
	}()

	init_cb(*db_drv, conn)
	conn.Close()
	conn = nil

	*models_file = file
	atomic.StoreInt32(&is_test, 1)

	Container = restful.NewContainer()
	Main()

	hsrv := httptest.NewServer(Container)

	log.Println("[ds-test] serving at '" + hsrv.URL + "'")
	defer hsrv.Close()

	time.Sleep(10 * time.Microsecond)
	cb(NewClient(hsrv.URL), definitions)

	if nil != srv_instance {
		srv_instance.Close()
		srv_instance = nil
	}
}

func SrvTest(t *testing.T, file string, cb func(client *Client, definitions *types.TableDefinitions)) {
	testBase(t, file, func(drv string, conn *sql.DB) {
		sql_str := `
DROP TABLE IF EXISTS tpt_alerts;
DROP TABLE IF EXISTS tpt_redis_commands;
DROP TABLE IF EXISTS tpt_actions;
DROP TABLE IF EXISTS tpt_metric_triggers;
DROP TABLE IF EXISTS tpt_triggers;
DROP TABLE IF EXISTS tpt_wbem_params;
DROP TABLE IF EXISTS tpt_ssh_params;
DROP TABLE IF EXISTS tpt_snmp_params;
DROP TABLE IF EXISTS tpt_endpoint_params;
DROP TABLE IF EXISTS tpt_access_params;
DROP TABLE IF EXISTS tpt_addresses;
DROP TABLE IF EXISTS tpt_device_port;
DROP TABLE IF EXISTS tpt_network_links;
DROP TABLE IF EXISTS tpt_network_devices;
DROP TABLE IF EXISTS tpt_managed_objects;
DROP TABLE IF EXISTS tpt_attributes;


CREATE TABLE tpt_networkdevices (
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

CREATE TABLE tpt_network_device_ports (
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


CREATE TABLE tpt_networklinks (
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

CREATE TABLE tpt_addresses (
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



CREATE TABLE tpt_snmp_params (
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

CREATE TABLE tpt_ssh_params (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  managed_object_id  integer,
  description        varchar(2000),

  address            varchar(50),
  port               integer,
  user_name          varchar(50),
  user_password      varchar(250)
);

CREATE TABLE tpt_wbem_params (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  managed_object_id  integer,
  description        varchar(2000),
  url                varchar(2000),
  user_name          varchar(50),
  user_password      varchar(250)
) ;


CREATE TABLE tpt_triggers (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  name          varchar(250),
  expression    varchar(250),
  attachment    varchar(2000),
  description   varchar(2000),

  parent_type   varchar(250),
  parent_id     integer
);

CREATE TABLE tpt_metric_triggers (
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

CREATE TABLE tpt_actions (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  name               varchar(250),
  description        varchar(2000),
  parent_type        varchar(250),
  parent_id          integer
);

CREATE TABLE tpt_redis_commands (
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


CREATE TABLE tpt_alerts (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  name               varchar(250),
  description        varchar(2000),
  parent_type        varchar(250),
  parent_id          integer,

  delay_times       integer,
  expression_style   varchar(50),
  expression_code    varchar(2000)
);


DROP TABLE IF EXISTS documents;
CREATE TABLE tpt_documents (
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
CREATE TABLE tpt_websites (id  INTEGER PRIMARY KEY AUTOINCREMENT, url varchar(200));

DROP TABLE IF EXISTS printers;
CREATE TABLE tpt_printers (id  INTEGER PRIMARY KEY AUTOINCREMENT, name varchar(200));

DROP TABLE IF EXISTS topics;
CREATE TABLE tpt_topics (id  INTEGER PRIMARY KEY AUTOINCREMENT, name varchar(200));

-- tables for CLOB
DROP TABLE IF EXISTS zip_files;
CREATE TABLE tpt_zip_files (id  INTEGER PRIMARY KEY AUTOINCREMENT, body text, document_id integer);
`

		sql_file := drv + "_test.sql"
		if "postgres" == drv && *IsPostgresqlInherit {
			sql_file = drv + "_inherit_test.sql"
		}

		if !commons.FileExists(sql_file) {
			file := "../data_store/" + drv + "_test.sql"
			if "postgres" == drv && *IsPostgresqlInherit {
				file = "../data_store/" + drv + "_inherit_test.sql"
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

	}, cb)
}

func createJson(t *testing.T, client *Client, target, msg string) string {
	_, id, e := client.CreateJson(client.Url+"/"+target, []byte(msg))
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
	return createJson(t, client, "network_device", fmt.Sprintf(`{"name":"dd%s", "type":"network_device", "address":"192.168.1.%s", "device_type":%s, "services":2%s}`, factor, factor, factor, factor))
}

func CreateMockSnmpParamsForTest(t *testing.T, client *Client, community string) string {
	return createJson(t, client, "snmp_param", fmt.Sprintf(`{ "port":161, "version":"snmp_v2c", "read_community":"%s"}`, community))
}

func CreateMockSshParamsForTest(t *testing.T, client *Client, pass string) string {
	return createJson(t, client, "ssh_param", fmt.Sprintf(`{ "port":22, "user_name":"a", "user_password":"%s"}`, pass))
}
