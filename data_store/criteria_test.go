package data_store

import (
	"bytes"
	"commons/types"
	"reflect"
	"strings"
	"testing"
)

var test_statements = []struct {
	query  map[string]string
	sql    string
	params []interface{}
	err    string
}{{query: map[string]string{"@id": "1"}, sql: " id = $1", params: []interface{}{int64(1)}, err: ""},
	/* 1  */ {query: map[string]string{"@id": "[gt]1"}, sql: " id > $1", params: []interface{}{int64(1)}, err: ""},
	/* 2  */ {query: map[string]string{"@id": "[gte]1"}, sql: " id >= $1", params: []interface{}{int64(1)}, err: ""},
	/* 3  */ {query: map[string]string{"@id": "[eq]1"}, sql: " id = $1", params: []interface{}{int64(1)}, err: ""},
	/* 4  */ {query: map[string]string{"@id": "[lt]1"}, sql: " id < $1", params: []interface{}{int64(1)}, err: ""},
	/* 5  */ {query: map[string]string{"@id": "[lte]1"}, sql: " id <= $1", params: []interface{}{int64(1)}, err: ""},
	/* 6  */ {query: map[string]string{"@id": "[ne]1"}, sql: " id != $1", params: []interface{}{int64(1)}, err: ""},
	/* 7  */ {query: map[string]string{"@id": "[in]1,2,3"}, sql: " id IN ( 1,2,3 )", params: nil, err: ""},
	/* 8  */ {query: map[string]string{"@id": "[nin]1,2,3"}, sql: " id NOT IN ( 1,2,3 )", params: nil, err: ""},
	/* 9  */ {query: map[string]string{"@Name": "[in]1,2,3"}, sql: " Name IN ( '1', '2', '3' )", params: nil, err: ""},
	/* 10 */ {query: map[string]string{"@Name": "[nin]1,2,3"}, sql: " Name NOT IN ( '1', '2', '3' )", params: nil, err: ""},
	/* 11 */ {query: map[string]string{"@id": "[between]1"}, err: "it must has two value"},
	/* 12 */ {query: map[string]string{"@id": "[between]1,2"}, sql: " (id BETWEEN $1 AND $2)", params: []interface{}{int64(1), int64(2)}, err: ""},
	/* 13 */ {query: map[string]string{"@Name": "[between]1,2"}, sql: " (Name BETWEEN $1 AND $2)", params: []interface{}{"1", "2"}, err: ""},
	/* 14 */ {query: map[string]string{"@id": "[is]null"}, sql: " id IS NULL", params: nil, err: ""},
	/* 15 */ {query: map[string]string{"@id": "[eq]1", "@Name": "aa"}, sql: " id = $1 AND Name = $2", params: []interface{}{int64(1), "aa"}, err: ""},
	/* 16 */ {query: map[string]string{"@id": "[eq]1", "@Name": "aa", "@Age": "2"}, sql: " id = $1 AND Name = $2 AND Age = $3", params: []interface{}{int64(1), "aa", int64(2)}, err: ""},
	/* 17 */ {query: map[string]string{"@id": "[eq]1", "@Name": "[between]1,2", "@Age": "2"}, sql: " id = $1 AND (Name BETWEEN $2 AND $3) AND Age = $4", params: []interface{}{int64(1), "1", "2", int64(2)}, err: ""}}

func TestCriteria(t *testing.T) {
	definitions, err := types.LoadTableDefinitions("etc/test1.xml")
	if nil != err {
		t.Errorf("read file 'test/test1.xml' failed, %s", err.Error())
		t.FailNow()
		return
	}
	person := definitions.Find("Person")
	if nil == person {
		t.Error("Person is not defined")
		return
	}

	for i, m := range test_statements {
		var buffer bytes.Buffer
		builder := &whereBuilder{table: person,
			idx:          1,
			isFirst:      true,
			buffer:       &buffer,
			operators:    default_operators,
			add_argument: (*whereBuilder).appendNumericArguments}

		e := builder.build(m.query)
		if nil != e {
			if 0 == len(m.err) {
				t.Errorf("index is %v, sql is %v, error is %v\r\n", i, m.sql, e)
				continue
			}

			if !strings.Contains(e.Error(), m.err) {
				t.Errorf("index is %v, sql is %v, error is %v\r\n", i, m.sql, e)
				continue
			}
			continue
		}

		if m.sql != buffer.String() {
			t.Errorf("index is %v, excepted sql is '%v', actual sql is '%v'\r\n", i, m.sql, buffer.String())
		}

		if !reflect.DeepEqual(m.params, builder.params) {
			t.Errorf("index is %v, excepted params is '%v', actual params is '%v'\r\n", i, m.params, builder.params)
		}
	}
}

func TestSplit(t *testing.T) {
	s1, s2 := split("[eq]12")
	if "eq" != s1 {
		t.Error("excepted is '%v', actual is '%v'", "eq", s1)
	}

	if "12" != s2 {
		t.Error("excepted is '%v', actual is '%v'", "12", s2)
	}

	s1, s2 = split("[eq]")
	if "eq" != s1 {
		t.Error("excepted is '%v', actual is '%v'", "eq", s1)
	}

	if "" != s2 {
		t.Error("excepted is '%v', actual is '%v'", "", s2)
	}

	s1, s2 = split("[]12")
	if "" != s1 {
		t.Error("excepted is '%v', actual is '%v'", "", s1)
	}

	if "12" != s2 {
		t.Error("excepted is '%v', actual is '%v'", "12", s2)
	}

	s1, s2 = split("[12")
	if "eq" != s1 {
		t.Error("excepted is '%v', actual is '%v'", "eq", s1)
	}

	if "[12" != s2 {
		t.Error("excepted is '%v', actual is '%v'", "[12", s2)
	}

	s1, s2 = split("]12")
	if "eq" != s1 {
		t.Error("excepted is '%v', actual is '%v'", "eq", s1)
	}

	if "]12" != s2 {
		t.Error("excepted is '%v', actual is '%v'", "]12", s2)
	}
}
