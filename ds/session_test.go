package ds

import (
	"commons/types"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"reflect"
	"testing"
	"time"
)

var (
	test_db    = flag.String("test.db", "postgres", "the db driver name for test")
	test_dbUrl = flag.String("test.dburl", "host=127.0.0.1 dbname=pqgotest user=postgres password=mfk sslmode=disable", "the db url")

	personName    = "people"
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

func simpleTest(t *testing.T, cb func(db *session, definitions *types.TableDefinitions)) {

	definitions, err := types.LoadTableDefinitions("etc/test1.xml")
	if nil != err {
		t.Errorf("read file 'etc/test1.xml' failed, %s", err.Error())
		t.FailNow()
		return
	}
	conn, err := sql.Open(*test_db, *test_dbUrl)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer conn.Close()

	_, err = conn.Exec("CREATE TEMP TABLE employees (ID SERIAL PRIMARY KEY, ID1 int, " +
		"Name varchar(256), " +
		"Name2 varchar(256), " +
		"Age int, " +
		"Day timestamp with time zone, " +
		"Mony numeric(9, 4), " +
		"IP varchar(50), " +
		"MAC varchar(50), " +
		"Sex varchar(10)," +
		"Password varchar(256)," +
		"company_test_id integer," +
		"Job varchar(256) )")
	if err != nil {
		t.Fatal(err)
		return
	}
	_, err = conn.Exec("CREATE TEMP TABLE managers (ID SERIAL PRIMARY KEY, ID1 int, " +
		"Name varchar(256), " +
		"Name2 varchar(256), " +
		"Age int, " +
		"Day timestamp with time zone, " +
		"Mony numeric(9, 4), " +
		"IP varchar(50), " +
		"MAC varchar(50), " +
		"Sex varchar(10)," +
		"Password varchar(256)," +
		"company_test_id integer," +
		"company_id integer," +
		"Job varchar(256) )")
	if err != nil {
		t.Fatal(err)
		return
	}

	_, err = conn.Exec("CREATE TEMP TABLE people (ID SERIAL PRIMARY KEY, ID1 int, " +
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
	simple := newSession(*test_db, conn, definitions)
	cb(simple, definitions)
}

func TestSimpleInsert(t *testing.T) {
	simpleTest(t, func(db *session, definitions *types.TableDefinitions) {
		person := definitions.Find("Person")
		if nil == person {
			t.Error("Person is not defined")
			return
		}

		id, err := db.insert(person, person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		result, err := db.findById(person, id, "")
		if nil != err {
			t.Error(err)
		} else {
			if (len(person1_attributes) + 2) != len(result) {
				t.Errorf("(len(person1_attributes)+2) != len(result), excepted is %d, actual is %d.",
					len(person1_attributes), len(result))
			}

			for k, v2 := range result {
				if person.Id.Name == k {
					continue
				}

				if "type" == k {
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

func TestSimpleUpdateById(t *testing.T) {
	simpleTest(t, func(db *session, definitions *types.TableDefinitions) {
		person := definitions.Find("Person")
		if nil == person {
			t.Error("Person is not defined")
			return
		}

		id, err := db.insert(person, person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		t.Log(id)

		err = db.updateById(person, id, person1_update_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		result, err := db.findById(person, id, "")
		if nil != err {
			t.Error(err)
		} else {
			if (len(person1_attributes) + 2) != len(result) {
				t.Errorf("(len(person1_attributes)+2) != len(result), excepted is %d, actual is %d.",
					len(person1_attributes), len(result))
			}

			for k, v2 := range result {
				if person.Id.Name == k {
					continue
				}

				if "type" == k {
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

func TestSimpleUpdateByParams(t *testing.T) {
	simpleTest(t, func(db *session, definitions *types.TableDefinitions) {
		person := definitions.Find("Person")
		if nil == person {
			t.Error("Person is not defined")
			return
		}

		id, err := db.insert(person, person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		t.Log(id)

		affected, err := db.update(person, map[string]string{"@id": fmt.Sprint(id)}, person1_update_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		if 1 != affected {
			t.Errorf("affected row is not equals 1, actual is %v", affected)
			return
		}

		result, err := db.findById(person, id, "")
		if nil != err {
			t.Error(err)
		} else {
			if (len(person1_attributes) + 2) != len(result) {
				t.Errorf("(len(person1_attributes)+2) != len(result), excepted is %d, actual is %d.",
					len(person1_attributes), len(result))
			}

			for k, v2 := range result {
				if person.Id.Name == k {
					continue
				}

				if "type" == k {
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

func TestSimpleFindById(t *testing.T) {
	simpleTest(t, func(db *session, definitions *types.TableDefinitions) {
		person := definitions.Find("Person")
		if nil == person {
			t.Error("Person is not defined")
			return
		}

		id, err := db.insert(person, person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		db_attributes, err := db.findById(person, id, "")
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		for k, v2 := range db_attributes {
			if person.Id.Name == k {
				continue
			}

			if "type" == k {
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

// func TestSimpleWhere(t *testing.T) {
// 	simpleTest(t, func(db *session, definitions *types.TableDefinitions) {
// 		person := definitions.Find("Person")
// 		if nil == person {
// 			t.Error("Person is not defined")
// 			return
// 		}

// 		id, err := db.insert(person, person1_attributes)
// 		if nil != err {
// 			t.Errorf(err.Error())
// 			return
// 		}

// 		it, err := db.where(person, "id = $1", id).Iter()
// 		if nil != err {
// 			t.Errorf(err.Error())
// 			return
// 		}

// 		results := make([]map[string]interface{}, 0)
// 		for {
// 			res := map[string]interface{}{}
// 			if !it.Next(res) {
// 				break
// 			}

// 			results = append(results, res)
// 		}

// 		if nil != it.Err() {
// 			t.Error(it.Err())
// 			return
// 		}

// 		if 1 != len(results) {
// 			t.Errorf("result is empty")
// 			return
// 		}

// 		db_attributes := results[0]
// 		for k, v2 := range db_attributes {
// 			if person.Id.Name == k {
// 				continue
// 			}

// if "type" == k {
// 	continue
// }

// 			v1, ok := person1_saved_attributes[k]
// 			if !ok {
// 				t.Error("'" + k + "' is not exists.")
// 			} else if !reflect.DeepEqual(v1, v2) {
// 				t.Errorf("'"+k+"' is not equals, excepted is [%T]%v, actual is [%T]%v.",
// 					v1, v1, v2, v2)
// 			}
// 		}
// 	})
// }

func TestSimpleFindByParams(t *testing.T) {
	simpleTest(t, func(db *session, definitions *types.TableDefinitions) {
		person := definitions.Find("Person")
		if nil == person {
			t.Error("Person is not defined")
			return
		}

		id, err := db.insert(person, person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		results, err := db.find(person, map[string]string{"@id": fmt.Sprint(id)})
		if nil != err {
			t.Errorf(err.Error())
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

			if "type" == k {
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

func TestSimpleDeleteById(t *testing.T) {
	simpleTest(t, func(db *session, definitions *types.TableDefinitions) {
		person := definitions.Find("Person")
		if nil == person {
			t.Error("Person is not defined")
			return
		}

		id, err := db.insert(person, person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		err = db.deleteById(person, id)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		_, err = db.findById(person, id, "")
		if nil == err {
			t.Errorf("delete failed, becase refind sucessed")
			return
		}
		if sql.ErrNoRows != err {
			t.Error(err)
		}
	})
}

func TestSimpleDeleteByParams(t *testing.T) {
	simpleTest(t, func(db *session, definitions *types.TableDefinitions) {
		person := definitions.Find("Person")
		if nil == person {
			t.Error("Person is not defined")
			return
		}

		id, err := db.insert(person, person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		affected, err := db.delete(person, map[string]string{"@id": fmt.Sprint(id)})
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		if 1 != affected {
			t.Errorf("affected row is not equals 1, actual is %v", affected)
			return
		}

		_, err = db.findById(person, id, "")
		if nil == err {
			t.Errorf("delete failed, becase refind sucessed")
			return
		}
		if sql.ErrNoRows != err {
			t.Error(err)
		}
	})
}
