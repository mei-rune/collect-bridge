package main

import (
	"fmt"
	"testing"
)

const (
	s1 = `
function receive (prod)
    local status, action, params = coroutine.resume(prod)
    return action, params
end

function send (x)
    coroutine.yield(x)
end

function execute_task (action, task)
  return coroutine.create(function()
    return "test ok"
    end)
end

function loop ()
  while true do
    local action, params = receive(p)  -- get new value
    if "__exit__" == action then
      print("lua vm exited\n")
      break
    end
    co = execute_task(action, params)
    send(co)
  end
  print("lua vm exited\n")
end

print("welcome to lua vm\n")
return coroutine.create(loop)

	`
)

func TestSpawn(t *testing.T) {
	drv := NewLuaDriver(s1)
	drv.Start()
	fmt.Println("---------")
	drv.Stop()

	fmt.Println("---------2")
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
