package main

import (
	"encoding/xml"
	"testing"
)

type XMLTest struct {
	XMLName xml.Name `xml:"test"`
	Ts      []XMLEntity
}

type XMLEntity struct {
	XMLName xml.Name `xml:"t"`
	Value   string   `xml:"value"`
}

func TestXml(t *testing.T) {
	var t1, t2 XMLTest
	t1.Ts = []XMLEntity{XMLEntity{Value: "a"}, XMLEntity{Value: "b"}}
	txt, _ := xml.Marshal(&t1)
	err := xml.Unmarshal(txt, &t2)
	if nil != err {
		t.Errorf("read xml failed - %v", err)
	}

	if nil == t2.Ts || 0 == len(t2.Ts) {
		t.Log(string(txt))
		t.Error("test xml failed.")
	}
}
