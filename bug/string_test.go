package main

import (
	"fmt"
	"testing"
)

func PrintStrings(ifs []interface{}) {
	fmt.Println(ifs)
}
func TestString(t *testing.T) {
	PrintStrings([]string{"1", "2"})
}
