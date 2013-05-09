package main

import (
	"errors"
	"fmt"
	"testing"
)

func Foo(b bool) (i int, err error) {
	defer fmt.Println(err)
	if b {
		err = errors.New("aaaddd")
		i = 4
		return
	}
	i = 5
	return
}

func TestMapGet(t *testing.T) {
	if i, _ := Foo(true); 8 != i {
		t.Error("aaa")
	}
	if i, _ := Foo(false); 12 != i {
		t.Error("aaaddd")
	}
}
