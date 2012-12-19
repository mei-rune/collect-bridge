package mdb

import (
	a "commons/assert"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"reflect"
	"testing"
	"time"
)

var (
	person1_attributes = map[string]interface{}{"ID1": int64(12),
		"Name":     "mmm",
		"Name2":    "m22",
		"Age":      int64(23),
		"Mony":     2.01,
		"IP":       "13.14.17.13",
		"MAC":      "22:32:62:82:52:42",
		"Sex":      "female",
		"Password": "12344"}

	person1_saved_attributes = map[string]interface{}{"ID1": SqlInteger64(12),
		"Name":     SqlString("mmm"),
		"Name2":    SqlString("m22"),
		"Age":      SqlInteger64(23),
		"Mony":     SqlDecimal(2.01),
		"Sex":      SqlString("female"),
		"Password": SqlPassword("12344")}

	person1_update_attributes = map[string]interface{}{"ID1": int64(22),
		"Name":  "maa",
		"Name2": "m11",
		"Age":   int64(13)}
)

func init() {
	person1_attributes["Day"], _ = time.Parse(datetimeType.Layout, "2009-12-12T12:23:23+08:00")
	person1_saved_attributes["Day"], _ = datetimeType.Convert("2009-12-12T12:23:23+08:00")
	person1_saved_attributes["IP"], _ = ipAddressType.Convert("13.14.17.13")
	person1_saved_attributes["MAC"], _ = physicalAddressType.Convert("22:32:62:82:52:42")
}

func newServer(t *testing.T) (*mgo.Session, *mgo.Database, *mdb_server) {

	definitions, err := LoadXml("test/test1.xml")
	if nil != err {
		t.Errorf("read file 'test/test1.xml' failed, %s", err.Error())
		merr, _ := err.(*MutiErrors)
		if nil != merr && nil != merr.errs {
			for _, e := range merr.errs {
				t.Errorf(e.Error())
			}
		}
		t.FailNow()
		return nil, nil, nil
	}

	if nil == definitions.clsDefinitions {
		t.Errorf("unmarshal xml 'test/test1.xml' error, classDefinition is nil")
		return nil, nil, nil
	}
	if 3 != len(definitions.clsDefinitions) {
		t.Errorf("unmarshal xml 'test/test1.xml' error, len of classDefinitions is not 2",
			len(definitions.clsDefinitions))
		t.FailNow()
		return nil, nil, nil
	}

	sess, err := mgo.Dial("127.0.0.1")
	if nil != err {
		t.Errorf("connect to mongo server failed, %s", err.Error())
		t.FailNow()
		return nil, nil, nil
	}
	sess.SetSafe(&mgo.Safe{W: 1, FSync: true, J: true})

	if nil == sess.Safe() {
		t.Errorf("mongo server invoke SetSafe failed")
		t.FailNow()
		return nil, nil, nil
	}

	db := sess.DB("test")

	return sess, db, &mdb_server{definitions: definitions, driver: &mgo_driver{session: db}}
}

func TestSimpleInsertByServer(t *testing.T) {

	simpleTest(t, []string{"Person"}, func(sess *mgo.Session, db *mgo.Database, svr *mdb_server) {
		person := svr.definitions.Find("Person")
		a.Assert(t, person, a.NotNil, a.Commentf("person class is not nil"))

		id, err := svr.Create(person, person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			merr, _ := err.(*MutiErrors)
			if nil != merr && nil != merr.errs {
				for _, e := range merr.errs {
					t.Errorf(e.Error())
				}
			}
		}
		if "" == id {
			t.Error("result.id of insert is nil")
		}
		t.Log("id=" + id)

		var result bson.M
		err = db.C("Person").FindId(id).One(&result)
		if nil != err {
			t.Error(err)
		} else {
			if (len(person1_attributes) + 1) != len(result) {
				t.Errorf("(len(person1_attributes)+1) != len(result), excepted is %d, actual is %d.",
					len(person1_attributes), len(result))
			}

			for k, v2 := range result {
				if "_id" == k {
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

	simpleTest(t, []string{"Person"}, func(sess *mgo.Session, db *mgo.Database, svr *mdb_server) {
		person := svr.definitions.Find("Person")
		a.Assert(t, person, a.NotNil, a.Commentf("person class is not nil"))

		id, err := svr.Create(person, person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			merr, _ := err.(*MutiErrors)
			if nil != merr && nil != merr.errs {
				for _, e := range merr.errs {
					t.Errorf(e.Error())
				}
			}
		}
		if "" == id {
			t.Error("result.id of insert is nil")
		}
		t.Log("id=" + id)

		err = svr.Update(person, id, person1_update_attributes)
		if nil != err {
			t.Errorf(err.Error())
			merr, _ := err.(*MutiErrors)
			if nil != merr && nil != merr.errs {
				for _, e := range merr.errs {
					t.Errorf(e.Error())
				}
			}
		}

		var result bson.M
		err = db.C("Person").FindId(id).One(&result)
		if nil != err {
			t.Error(err)
		} else {
			if (len(person1_attributes) + 1) != len(result) {
				t.Errorf("(len(person1_attributes)+1) != len(result), excepted is %d, actual is %d.",
					len(person1_attributes), len(result))
			}

			for k, v2 := range result {
				if "_id" == k {
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

func TestSimpleFindByidByServer(t *testing.T) {
	simpleTest(t, []string{"Person"}, func(sess *mgo.Session, db *mgo.Database, svr *mdb_server) {
		person := svr.definitions.Find("Person")
		a.Assert(t, person, a.NotNil, a.Commentf("person class is not nil"))

		id, err := svr.Create(person, person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			merr, _ := err.(*MutiErrors)
			if nil != merr && nil != merr.errs {
				for _, e := range merr.errs {
					t.Errorf(e.Error())
				}
			}
			return
		}
		if "" == id {
			t.Error("result.id of insert is nil")
			return
		}
		t.Log("id=" + id)

		var result bson.M
		err = db.C("Person").FindId(id).One(&result)
		if nil != err {
			t.Error(err)
			return
		}
		db_attributes, err := svr.FindById(person, id)
		if nil != err {
			t.Errorf(err.Error())
			merr, _ := err.(*MutiErrors)
			if nil != merr && nil != merr.errs {
				for _, e := range merr.errs {
					t.Errorf(e.Error())
				}
			}
			return
		}

		for k, v2 := range db_attributes {
			if "_id" == k {
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

	simpleTest(t, []string{"Person"}, func(sess *mgo.Session, db *mgo.Database, svr *mdb_server) {

		person := svr.definitions.Find("Person")
		a.Assert(t, person, a.NotNil, a.Commentf("person class is not nil"))

		id, err := svr.Create(person, person1_attributes)
		if nil != err {
			t.Errorf(err.Error())
			merr, _ := err.(*MutiErrors)
			if nil != merr && nil != merr.errs {
				for _, e := range merr.errs {
					t.Errorf(e.Error())
				}
			}
			return
		}
		if "" == id {
			t.Error("result.id of insert is nil")
			return
		}
		t.Log("id=" + id)

		var result bson.M
		err = db.C("Person").FindId(id).One(&result)
		if nil != err {
			t.Error(err)
			return
		}

		err = svr.RemoveById(person, id)
		if nil != err {
			t.Errorf(err.Error())
			return
		}

		err = db.C("Person").FindId(id).One(&result)
		if nil == err {
			t.Error("remove failed")
			return
		}

		if "not found" != err.Error() {
			t.Error("remove failed")
		}
	})
}

func simpleTest(t *testing.T, params []string,
	cb func(sess *mgo.Session, db *mgo.Database, svr *mdb_server)) {

	sess, db, svr := newServer(t)
	for _, s := range params {
		db.C(s).DropCollection()
	}
	defer func() {
		for _, s := range params {
			db.C(s).DropCollection()
		}
		sess.Close()
	}()

	cb(sess, db, svr)
}
