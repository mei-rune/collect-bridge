package mdb

import (
	a "commons/assert"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"testing"
)

var (
	person1 = map[string]interface{}{"ID1": int64(12),
		"Name":     "mmm",
		"Name2":    "m22",
		"Age":      int64(23),
		"Day":      "2009-12-12 12:23:23",
		"Mony":     2.01,
		"IP":       "13.14.17.13",
		"MAC":      "22:32:62:82:52:42",
		"Sex":      "female",
		"Password": "12344"}
)

func newServer(t *testing.T) (*mgo.Session, *mgo.Database, *MdbServer) {

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

	return sess, db, &MdbServer{definitions: definitions, driver: &mgo_driver{session: db}}
}

func TestSimpleInsertByServer(t *testing.T) {
	sess, db, svr := newServer(t)
	db.C("Person").DropCollection()

	defer func() {
		db.C("Person").DropCollection()
		sess.Close()
	}()
	person := svr.definitions.Find("Person")
	a.Assert(t, person, a.NotNil, a.Commentf("person class is not nil"))

	id, err := svr.Create(person, person1)
	if nil != err {
		t.Errorf(err.Error())
		merr, _ := err.(*MutiErrors)
		if nil != merr && nil != merr.errs {
			for _, e := range merr.errs {
				t.Errorf(e.Error())
			}
		}
	}
	if nil == id {
		t.Error("result.id of insert is nil")
	}

	var result bson.M
	err = db.C("Person").FindId(id).One(&result)
	if nil != err {
		t.Error(err)
	} else {
		if len(person1) != (len(result) + 1) {
			t.Error("len(person1) != (len(result) + 1)")
		}

		for k, v2 := range result {
			if "_id" == k {
				continue
			}
			v1, ok := person1[k]
			if !ok {
				t.Error("'" + k + "' is not exists.")
			} else if v1 != v2 {
				t.Errorf("'"+k+"' is not equals, excepted is %v, actual is %v.", v1, v2)
			}
		}
	}

	t.Errorf("not implemented")
}

func TestSimpleUpdateByServer(t *testing.T) {
	t.Errorf("not implemented")
}

func TestSimpleFindByidByServer(t *testing.T) {
	t.Errorf("not implemented")
}

func TestSimpleDeleteByidByServer(t *testing.T) {
	t.Errorf("not implemented")
}
