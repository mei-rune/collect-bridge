function receive ()
    local action, params = coroutine.yield()
    return action, params
end

function send (co, ...)
    local action, params = coroutine.yield(co, ...)
    return action, params
end

function execute_task (action, task)
  return coroutine.create(function()
    return "test ok"
    end)
end

function loop ()
  print("lua enter looping")
  local action, params = receive()  -- get new value
  while "__exit__" ~= action do
    print("lua vm receive - '"..action.."'\n")
    co = execute_task(action, params)
    action, params = send(co)
  end
  print("lua exit looping")
end

print("welcome to lua vm")
loop ()