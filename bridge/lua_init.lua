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
      print("lua vm exited\n")
    if "__exit__" == action then
      break
    end
    co = execute_task(action, params)
    send(co)
  end
  print("lua vm exited\n")
end

print("welcome to lua vm\n")
return coroutine.create(loop)