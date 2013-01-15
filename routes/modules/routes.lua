module("routes",  package.seeall)

local pcall = pcall
mj.routes = {}

local route = ml.class()
function route:equal(key, value)
  return self
end
function route:start_with(f)
  return self
end
function route:end_with(f)
  return self
end
function route:contains(f)
  return self
end
function route:and_with(f)
  return self
end

function sandbox(f, sb_env)
  if not f then return nil, "sandbox function not valid" end


  sb_env.route()
  local orig_env = _ENV
  _ENV = sb_env

  local pcall_res, message = pcall( f )
  local modified_env = _ENV
  _ENV = orig_env
  if pcall_res then
    return nil, modified_env
  end

  if nil == message then
    return "excute failed!", nil
  end

  return message, nil
end


function load_routefile(file)
  local res, rt = nil, {}
  rt.mj= mj
  rt.ml= ml
  rt.route= route
  
  res, rt = sandbox(loadfile(file), rt)
  if nil ~= res then
    error(res)
  end

  local name = rt.name
  local level = rt.level
  local categories = rt.categories
  local match = rt.match
  local action = rt.action

  if nil == name or "" == name then
    error("load '" .. file .."' failed, name is required")
  end
  if nil == level or "" == level then
    error("load '" .. file .."' failed, level is required")
  end 
  if nil == categories or "" == categories then
    error("load '" .. file .."' failed, categories is required")
  end 
  if nil == match or "" == match then
    error("load '" .. file .."' failed, match is required")
  end 
  if nil == action or "" == action then
    error("load '" .. file .."' failed, action is required")
  end
  
  return rt
end