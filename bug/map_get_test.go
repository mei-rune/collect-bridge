package main

import (
	"fmt"
	"testing"
)

func Get(ifs map[string]int, k string) (int, bool) {
	return ifs[k]
}
func TestMapGet(t *testing.T) {
	m := make(map[string]int)
	m["a"] = 1
	if i, _ := Get(m, "a"); 1 != i {
		t.Errorf("test")
	}
}
