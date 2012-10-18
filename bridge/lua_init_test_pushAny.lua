DEBUG = 9000
INFO = 6000
WARN = 4000
ERROR = 2000
FATAL = 1000
SYSTEM = 0

function receive ()
    local action, params = coroutine.yield()
    return action, params
end

function send_and_recv ( ...)
    local action, params = coroutine.yield( ...)
    return action, params
end

function log(level, msg)
  if "number" ~= type(level) then
    return nil, "'params' is not a table."
  end

  coroutine.yield("log", level, msg)
end

function loop()
  log(SYSTEM, "lua enter looping")
  local action, params = receive()  -- get new value
  while "__exit__" ~= action do
    if nil == params then
      print("lua vm receive - '"..action.."' - nil" )
    else
      print("lua vm receive - '"..action.."' -", params)
    end
    action, params = send_and_recv(action, params)
  end
  log(SYSTEM, "lua exit looping")
end

log(SYSTEM, "welcome to lua vm")
loop ()
