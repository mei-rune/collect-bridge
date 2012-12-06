package main

import (
	"fmt"
	"testing"
)

func TestCompiler(t *testing.T) {
	//var ok bool
	testSlice := []interface{}{"a", 12, 2}
	if true {
		s, ok := testSlice[0].(string)
		fmt.Println(ok, s)
	}
	// for i, c := range testSlice {
	//	s, ok := c.(string)
	//	fmt.Println(i, ok, s)
	// }
}
