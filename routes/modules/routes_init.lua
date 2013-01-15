module("routes",  package.seeall)

local route = ml.class()
function route.equal(key, value)

end

local init_files = ml.Array(mj.enumerate_files(mj.join_path(mj.execute_directory, "modules")))
if(mj.work_directory ~= mj.execute_directory) then
  init_files = init_files .. mj.enumerate_files(mj.join_path(mj.work_directory, "modules"))
end