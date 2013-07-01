package ds

import (
	"commons/types"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"reflect"
	"strings"
	"testing"
)

func serverTest(t *testing.T, file string, cb func(db *Client, definitions *types.TableDefinitions)) {
	testBase(t, file, func(drv string, conn *sql.DB) {
		_, err := conn.Exec("DROP TABLE IF EXISTS tpt_people")
		if err != nil {
			t.Fatal(err)
			t.FailNow()
			return
		}

		_, err = conn.Exec("DROP TABLE IF EXISTS tpt_employees")
		if err != nil {
			t.Fatal(err)
			t.FailNow()
			return
		}

		_, err = conn.Exec("DROP TABLE IF EXISTS tpt_managers")
		if err != nil {
			t.Fatal(err)
			t.FailNow()
			return
		}

		primaryKey_decl := "INTEGER PRIMARY KEY AUTOINCREMENT"
		timezone := ""
		if "postgres" == *test_db {
			primaryKey_decl = "SERIAL PRIMARY KEY"
			timezone = " with time zone"
		}

		_, err = conn.Exec("CREATE TABLE tpt_employees (ID " + primaryKey_decl + ", ID1 int, " +
			"Name varchar(256), " +
			"Name2 varchar(256), " +
			"Age int, " +
			"Day timestamp " + timezone + ", " +
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

		_, err = conn.Exec("CREATE TABLE tpt_managers (ID " + primaryKey_decl + ", ID1 int, " +
			"Name varchar(256), " +
			"Name2 varchar(256), " +
			"Age int, " +
			"Day timestamp " + timezone + ", " +
			"Mony numeric(9, 4), " +
			"IP varchar(50), " +
			"MAC varchar(50), " +
			"Sex varchar(10)," +
			"company_test_id integer," +
			"company_id integer," +
			"Password varchar(256)," +
			"Job varchar(256) )")
		if err != nil {
			t.Fatal(err)
			return
		}

		_, err = conn.Exec("CREATE TABLE tpt_people (ID " + primaryKey_decl + ", ID1 int, " +
			"Name varchar(256), " +
			"Name2 varchar(256), " +
			"Age int, " +
			"Day timestamp " + timezone + ", " +
			"Mony numeric(9, 4), " +
			"IP varchar(50), " +
			"MAC varchar(50), " +
			"Sex varchar(10)," +
			"Password varchar(256) )")
		if err != nil {
			t.Fatal(err)
			t.FailNow()
			return
		}
	}, cb)
}

func convert(table *types.TableDefinition, values map[string]interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range values {
		attribute := table.GetAttribute(k)
		if nil == attribute {
			continue
		}
		res[k], _ = attribute.Type.ToInternal(v)
	}
	return res
}
func TestSrvInsert(t *testing.T) {
	serverTest(t, "etc/test1.xml", func(db *Client, definitions *types.TableDefinitions) {
		person := definitions.Find("Person")
		if nil == person {
			t.Error("Person is not defined")
			return
		}

		id, err := db.Create("person", person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		result, err := db.FindById("person", id)
		if nil != err {
			t.Error(err)
		} else {
			if (len(person1_attributes) + 2) != len(result) {
				t.Errorf("(len(person1_attributes)+2) != len(result), excepted is %d, actual is %d.",
					len(person1_attributes), len(result))
			}

			for k, v2 := range convert(person, result) {
				if "id" == k {
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

func TestSrvUpdateById(t *testing.T) {
	serverTest(t, "etc/test1.xml", func(db *Client, definitions *types.TableDefinitions) {
		person := definitions.Find("Person")
		if nil == person {
			t.Error("Person is not defined")
			return
		}

		id, err := db.Create("person", person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		t.Log(id)

		err = db.UpdateById("person", id, person1_update_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		result, err := db.FindById("person", id)
		if nil != err {
			t.Error(err)
		} else {
			if (len(person1_attributes) + 2) != len(result) {
				t.Errorf("(len(person1_attributes)+2) != len(result), excepted is %d, actual is %d.",
					len(person1_attributes), len(result))
			}

			for k, v2 := range convert(person, result) {
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

func TestSrvUpdateByParams(t *testing.T) {
	serverTest(t, "etc/test1.xml", func(db *Client, definitions *types.TableDefinitions) {
		person := definitions.Find("Person")
		if nil == person {
			t.Error("Person is not defined")
			return
		}

		id, err := db.Create("person", person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		t.Log(id)

		affected, err := db.UpdateBy("person", map[string]string{"id": fmt.Sprint(id)}, person1_update_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		if 1 != affected {
			t.Errorf("affected row is not equals 1, actual is %v", affected)
			return
		}

		result, err := db.FindById("person", id)
		if nil != err {
			t.Error(err)
		} else {
			if (len(person1_attributes) + 2) != len(result) {
				t.Errorf("(len(person1_attributes)+2) != len(result), excepted is %d, actual is %d.",
					len(person1_attributes), len(result))
			}

			for k, v2 := range convert(person, result) {
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

func TestSrvFindById(t *testing.T) {
	serverTest(t, "etc/test1.xml", func(db *Client, definitions *types.TableDefinitions) {
		person := definitions.Find("Person")
		if nil == person {
			t.Error("Person is not defined")
			return
		}

		id, err := db.Create("person", person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		db_attributes, err := db.FindById("person", id)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		for k, v2 := range convert(person, db_attributes) {
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

// func TestSrvWhere(t *testing.T) {
// 	serverTest(t, "etc/test1.xml", func(db *Client, definitions *types.TableDefinitions) {
// 		person := definitions.Find("Person")
// 		if nil == person {
// 			t.Error("Person is not defined")
// 			return
// 		}

// 		id, err := db.Create("person", person1_attributes)
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
// 		for k, v2 := range convert(person, db_attributes) {
// 			if person.Id.Name == k {
// 				continue
// 			}

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

func TestSrvFindByParams(t *testing.T) {
	serverTest(t, "etc/test1.xml", func(db *Client, definitions *types.TableDefinitions) {
		person := definitions.Find("Person")
		if nil == person {
			t.Error("Person is not defined")
			return
		}

		id, err := db.Create("person", person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		results, err := db.FindBy("person", map[string]string{"id": fmt.Sprint(id)})
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		if 1 != len(results) {
			t.Errorf("result is empty")
			return
		}

		db_attributes := results[0]
		for k, v2 := range convert(person, db_attributes) {
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

func TestSrvDeleteById(t *testing.T) {
	serverTest(t, "etc/test1.xml", func(db *Client, definitions *types.TableDefinitions) {
		person := definitions.Find("Person")
		if nil == person {
			t.Error("Person is not defined")
			return
		}

		id, err := db.Create("person", person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		err = db.DeleteById("person", id)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		_, err = db.FindById("person", id)
		if nil == err {
			t.Errorf("delete failed, becase refind sucessed")
			return
		}
		if !strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			t.Error("actual is " + err.Error())
			t.Error("excepted is " + sql.ErrNoRows.Error())
		}
	})
}

func TestSrvDeleteByParams(t *testing.T) {
	serverTest(t, "etc/test1.xml", func(db *Client, definitions *types.TableDefinitions) {
		person := definitions.Find("Person")
		if nil == person {
			t.Error("Person is not defined")
			return
		}

		id, err := db.Create("person", person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		affected, err := db.DeleteBy("person", map[string]string{"id": fmt.Sprint(id)})
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		if 1 != affected {
			t.Errorf("affected row is not equals 1, actual is %v", affected)
			return
		}

		_, err = db.FindById("person", id)
		if nil == err {
			t.Errorf(err.Error())
			return
		}
		if !strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			t.Error(err)
		}
	})
}
