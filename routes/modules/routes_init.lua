module("routes_init",  package.seeall)

print(package.path)
print(package.cpath)

require 'routes'

function filename(name)
  pa = ml.splitpath(name)
  return ml.splitext(pa)
end

for i, file in ipairs(mj.enumerate_scripts(".*_route%.lua$")) do
  mj.log(mj.SYSTEM, "load route file -- '" .. file .. "'")
  mj.routes[filename(file)] = routes.load_routefile(file)
end

mj.log(mj.SYSTEM, "load route file finished.")