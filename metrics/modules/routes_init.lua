module("routes_init",  package.seeall)

require 'routes'
local cjson = require 'cjson'

function filename(name)
  local pa, file = ml.splitpath(name)
  return ml.splitext(file)
end

for i, file in ipairs(mj.enumerate_scripts(".*_route%.lua$")) do
  mj.log(mj.SYSTEM, "load route file -- '" .. file .. "'")
  local rt = routes.load_routefile(file)
  rt.file = file
  table.insert(mj.routes, rt)
end


-- mj.log(mj.SYSTEM, "dump routes")
-- function dump(obj, depth)
--   if depth > 2 then
--	return 
--   end
--   if type(obj) == "function" then
--	mj.log(mj.SYSTEM, depth .. "- function ")
--   elseif type(obj) == "table" then
--	for i, s in pairs(obj) do
--		mj.log(mj.SYSTEM, depth .. i .. ":")
--		dump(s, depth + 1)
--	end
--   --elseif type(obj) == "userdata" then
--   --	mj.log(mj.SYSTEM, "- userdata")
--   else
--	mj.log(mj.SYSTEM, depth .. "- " .. type(obj))
--   end
-- end

-- dump(mj.routes, 0)
mj.log(mj.SYSTEM, cjson.encode(mj.routes))

mj.log(mj.SYSTEM, "load route file finished.")