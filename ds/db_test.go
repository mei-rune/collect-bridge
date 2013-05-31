package ds

import (
	"commons/types"
	"database/sql"
	_ "github.com/lib/pq"
	"os"
	"reflect"
	"testing"
	"time"
)

var (
	personName    = "persons"
	personClsName = "Person"

	person1_attributes = map[string]interface{}{"ID1": int64(12),
		"Name":     "mmm",
		"Name2":    "m22",
		"Age":      int64(23),
		"Mony":     2.01,
		"IP":       types.IPAddress("13.14.17.13"),
		"MAC":      types.PhysicalAddress("22:32:62:82:52:42"),
		"Sex":      "female",
		"Password": types.Password("12344")}

	person1_saved_attributes = map[string]interface{}{"ID1": int64(12),
		"Name":     string("mmm"),
		"Name2":    string("m22"),
		"Age":      int64(23),
		"Mony":     float64(2.01),
		"Sex":      string("female"),
		"Password": types.Password("12344")}

	person1_update_attributes = map[string]interface{}{"ID1": int64(22),
		"Name":  "maa",
		"Name2": "m11",
		"Age":   int64(13)}
)

func init() {
	person1_attributes["Day"], _ = time.Parse(types.DATETIMELAYOUT, "2009-12-12T12:23:23+08:00")
	person1_saved_attributes["Day"], _ = types.GetTypeDefinition("datetime").ToInternal("2009-12-12T12:23:23+08:00")
	person1_saved_attributes["IP"], _ = types.GetTypeDefinition("ipAddress").ToInternal("13.14.17.13")
	person1_saved_attributes["MAC"], _ = types.GetTypeDefinition("physicalAddress").ToInternal("22:32:62:82:52:42")
}

func simpleTest(t *testing.T, cb func(drv string, conn *sql.DB, definitions *types.TableDefinitions)) {

	definitions, err := types.LoadTableDefinitions("etc/test1.xml")
	if nil != err {
		t.Errorf("read file 'test/test1.xml' failed, %s", err.Error())
		t.FailNow()
		return
	}

	datname := os.Getenv("PGDATABASE")
	sslmode := os.Getenv("PGSSLMODE")
	user := os.Getenv("PGUSER")
	password := os.Getenv("PGPASSWORD")

	if datname == "" {
		os.Setenv("PGDATABASE", "pqgotest")
	}

	if sslmode == "" {
		os.Setenv("PGSSLMODE", "disable")
	}
	if user == "" {
		os.Setenv("PGUSER", "postgres")
	}

	if password == "" {
		os.Setenv("PGPASSWORD", "mfk")
	}

	conn, err := sql.Open("postgres", "")
	if err != nil {
		t.Fatal(err)
		return
	}
	defer conn.Close()

	_, err = conn.Exec("CREATE TEMP TABLE persons (ID SERIAL PRIMARY KEY, ID1 int, " +
		"Name varchar(256), " +
		"Name2 varchar(256), " +
		"Age int, " +
		"Day timestamp with time zone, " +
		"Mony numeric(9, 4), " +
		"IP varchar(50), " +
		"MAC varchar(50), " +
		"Sex varchar(10)," +
		"Password varchar(256) )")
	if err != nil {
		t.Fatal(err)
		return
	}

	cb("pg", conn, definitions)
}

func TestSimpleInsertByServer(t *testing.T) {
	simpleTest(t, func(drv string, conn *sql.DB, definitions *types.TableDefinitions) {
		person := definitions.Find("Person")
		if nil == person {
			t.Error("Person is not defined")
			return
		}

		id, err := Insert(drv, conn, person, person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		result, err := FindById(drv, conn, person, id)
		if nil != err {
			t.Error(err)
		} else {
			if (len(person1_attributes) + 1) != len(result) {
				t.Errorf("(len(person1_attributes)+1) != len(result), excepted is %d, actual is %d.",
					len(person1_attributes), len(result))
			}

			for k, v2 := range result {
				if person.Id.Name == k {
					continue
				}

				v1, ok := person1_attributes[k]
				if !ok {
					t.Error("'" + k + "' is not exists.")
				} else if v1 != v2 {
					t.Errorf("'"+k+"' is not equals, excepted is [%T]%v, actual is [%T]%v.",
						v1, v1, v2, v2)
				}
			}
		}
	})
}

func TestSimpleUpdateByServer(t *testing.T) {
	simpleTest(t, func(drv string, conn *sql.DB, definitions *types.TableDefinitions) {
		person := definitions.Find("Person")
		if nil == person {
			t.Error("Person is not defined")
			return
		}

		id, err := Insert(drv, conn, person, person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		t.Log(id)

		err = UpdateById(drv, conn, person, person1_update_attributes, id)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		result, err := FindById(drv, conn, person, id)
		if nil != err {
			t.Error(err)
		} else {
			if (len(person1_attributes) + 1) != len(result) {
				t.Errorf("(len(person1_attributes)+1) != len(result), excepted is %d, actual is %d.",
					len(person1_attributes), len(result))
			}

			for k, v2 := range result {
				if person.Id.Name == k {
					continue
				}

				v1, ok := person1_update_attributes[k]
				if ok {
					if v1 != v2 {
						t.Errorf("'"+k+"' is not equals, excepted is [%T]%v, actual is [%T]%v.",
							v1, v1, v2, v2)
					}
				} else if v1, ok = person1_attributes[k]; !ok {
					t.Error("'" + k + "' is not exists.")
				} else if v1 != v2 {
					t.Errorf("'"+k+"' is not equals, excepted is [%T]%v, actual is [%T]%v.",
						v1, v1, v2, v2)
				}

			}
		}
	})
}

func TestSimpleFindByIdByServer(t *testing.T) {
	simpleTest(t, func(drv string, conn *sql.DB, definitions *types.TableDefinitions) {
		person := definitions.Find("Person")
		if nil == person {
			t.Error("Person is not defined")
			return
		}

		id, err := Insert(drv, conn, person, person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		db_attributes, err := FindById(drv, conn, person, id)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		for k, v2 := range db_attributes {
			if person.Id.Name == k {
				continue
			}

			v1, ok := person1_saved_attributes[k]
			if !ok {
				t.Error("'" + k + "' is not exists.")
			} else if !reflect.DeepEqual(v1, v2) {
				t.Errorf("'"+k+"' is not equals, excepted is [%T]%v, actual is [%T]%v.",
					v1, v1, v2, v2)
			}
		}
	})
}

func TestSimpleQueryByServer(t *testing.T) {
	simpleTest(t, func(drv string, conn *sql.DB, definitions *types.TableDefinitions) {
		person := definitions.Find("Person")
		if nil == person {
			t.Error("Person is not defined")
			return
		}

		id, err := Insert(drv, conn, person, person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		it, err := Where(drv, conn, person, "id = $1", id).Iter()
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		results := make([]map[string]interface{}, 0)
		for {
			res := map[string]interface{}{}
			if !it.Next(res) {
				break
			}

			results = append(results, res)
		}

		if nil != it.Err() {
			t.Error(it.Err())
			return
		}

		if 1 != len(results) {
			t.Errorf("result is empty")
			return
		}

		db_attributes := results[0]
		for k, v2 := range db_attributes {
			if person.Id.Name == k {
				continue
			}

			v1, ok := person1_saved_attributes[k]
			if !ok {
				t.Error("'" + k + "' is not exists.")
			} else if !reflect.DeepEqual(v1, v2) {
				t.Errorf("'"+k+"' is not equals, excepted is [%T]%v, actual is [%T]%v.",
					v1, v1, v2, v2)
			}
		}
	})
}

func TestSimpleDeleteByidByServer(t *testing.T) {

	simpleTest(t, func(drv string, conn *sql.DB, definitions *types.TableDefinitions) {
		person := definitions.Find("Person")
		if nil == person {
			t.Error("Person is not defined")
			return
		}

		id, err := Insert(drv, conn, person, person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		err = DeleteById(drv, conn, person, id)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		_, err = FindById(drv, conn, person, id)
		if nil == err {
			t.Errorf(err.Error())
			return
		}
		if sql.ErrNoRows != err {
			t.Error(err)
		}
	})
}
