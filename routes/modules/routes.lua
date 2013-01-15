module("routes",  package.seeall)

local pcall = pcall
mj.routes = {}

local route = ml.class()
function route:_init()
  self.filters = {}
  return self
end
function route:equal(key, value)
  table.insert(self.filters, {type= "equal", key= key, value= value})
  return self
end
function route:start_with(key, prefix)
  table.insert(self.filters, {type= "start_with", key= key, value= prefix})
  return self
end
function route:end_with(key, suffix)
  table.insert(self.filters, {type= "end_with", key= key, value= suffix})
  return self
end
function route:contains(key, sub)
  table.insert(self.filters, {type= "contains", key= key, value= sub})
  return self
end
function route:match(key, pat)
  table.insert(self.filters, {type= "match", key= key, value= pat})
  return self
end
function route:and_with(f)
  error("and_with is not implemented")
  return self
end

local get  = function(opts)
  return opts
end
local put  = function(opts)
  error("put is not implemented")
end
local create  = function(opts)
  error("create is not implemented")
end
local delete  = function(opts)
  error("delete is not implemented")
end

function load_routefile(file)
  local res, rt = nil, {}
  rt.__index = _G
  ml.update(rt, _G)
  rt.route= route
  rt.get = get
  rt.put = put
  rt.create = create
  rt.delete = delete



  if type(file) ~= "string" then
    error("argument 'file' must is a string")
  end
  local f, message = loadfile(file, "bt", rt)
  if nil == f then
    error(message or "load '" .. file .. "'failed")
  end
  f()

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
  if not route.classof(match) then
    error("load '" .. file .."' failed, match must is a route object")
  end
  if nil == action or "" == action then
    error("load '" .. file .."' failed, action is required")
  end
  
  return rt
end