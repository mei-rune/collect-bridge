
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

function execute(schema, action, params)
  if "table" ~= type(params) then
    return nil, "'params' is not a table."
  end
  return coroutine.yield(action, schema, params)
end

modules = {}

function init_modules(path, modules)
  package.path = package.path + ";" + path

  for k, v in pairs(modules) do
    if "string" == type(k) then
      modules[k] = require (v)
    else
      modules[v] = require (v)
    end
  end
end

function execute_module(module_name, action, params)
  module = modules[module_name]
  if nil == module then
    return nil, "module '"..module_name.."' is not exists."
  end
  func = module[action]
  if nil == func then
    return nil, "method '"..action.."' is not implemented in module '"..module_name.."'."
  end
  return func(params)
end

function execute_script(action, script)
  if 'string' ~= script then
    return nil, "'script' is not a string."
  end
  local env = {["action"] = action }
  setmetatable(env, _ENV)
  local _ENV = env
  func = assert(loadstring())
  return func()
end

function execute_task (action, params)
  --if nil == task then
  --  print("params = nil")
  --end

  return coroutine.create(function()
      if nil == params then
        return nil, "'params' is nil."
      end
      if "table" ~= type(params) then
        return nil, "'params' is not a table"
      end
      schema = params["schema"]
      if nil == schema then
        return nil, "'schema' is nil"
      elseif "script" == schema then
        return execute_script(action, params["script"])
      else
        return execute_module(schema, action, params)
      end
    end)
end


function loop()
  log(SYSTEM, "lua enter looping")
  local action, params = receive()  -- get new value
  while "__exit__" ~= action do
    log(SYSTEM, "lua vm receive - '"..action.."'")

    co = execute_task(action, params)
    action, params = send_and_recv(co)
  end
  log(SYSTEM, "lua exit looping")
end

log(SYSTEM, "welcome to lua vm")
loop ()