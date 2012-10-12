package main

import (
	"testing"
)

const (
	s1 = `
	`
)

func TestSpawn(t *testing.T) {
	drv := NewLuaDriver("")
	drv.Start()
	drv.Stop()
}

func doFunc(b bool, t *testing.T) {
	if b {
		defer func() {
			t.Error("it is faile")
		}()
	}
}

func TestDefer(t *testing.T) {
	doFunc(false, t)
}
