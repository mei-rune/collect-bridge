
local mj = {}

mj.DEBUG = 9000
mj.INFO = 6000
mj.WARN = 4000
mj.ERROR = 2000
mj.FATAL = 1000
mj.SYSTEM = 0

function mj:receive ()
    local action, params = coroutine.yield()
    return action, params
end

function mj:send_and_recv ( ...)
    local action, params = coroutine.yield( ...)
    return action, params
end

function mj:log(level, msg)
  if "number" ~= type(level) then
    return nil, "'params' is not a table."
  end

  coroutine.yield("log", level, msg)
end

function mj:execute(schema, action, params)
  if "table" ~= type(params) then
    return nil, "'params' is not a table."
  end
  return coroutine.yield(action, schema, params)
end


function mj:execute_module(module_name, action, params)
  module = require(module_name)
  if nil == module then
    return nil, "module '"..module_name.."' is not exists."
  end
  func = module[action]
  if nil == func then
    return nil, "method '"..action.."' is not implemented in module '"..module_name.."'."
  end
  return func(module, params)
end

function mj:execute_script(action, script, params)
  if 'string' ~= type(script) then
    return nil, "'script' is not a string."
  end
  local env = {["mj"] = self,
   ["action"] = action,
   ['params'] = params}
  setmetatable(env, _ENV)
  func = assert(load(script, nil, 'bt', env))
  return func()
end

function mj:execute_task(action, params)
  --if nil == task then
  --  print("params = nil")
  --end

  return coroutine.create(function()
      if nil == params then
        return nil, "'params' is nil."
      end
      if "table" ~= type(params) then
        return nil, "'params' is not a table, actual is '"..type(params).."'." 
      end
      schema = params["schema"]
      if nil == schema then
        return nil, "'schema' is nil"
      elseif "script" == schema then
        return self:execute_script(action, params["script"], params)
      else
        return self:execute_module(schema, action, params)
      end
    end)
end


function mj:loop()
  self:log(SYSTEM, "lua enter looping")
  local action, params = mj:receive()  -- get new value
  while "__exit__" ~= action do
    self:log(SYSTEM, "lua vm receive - '"..action.."'")

    co = self:execute_task(action, params)
    action, params = self:send_and_recv(co)
  end
  self:log(SYSTEM, "lua exit looping")
end

package.preload["mj"] = mj
mj:log(SYSTEM, "welcome to lua vm")
mj:loop ()