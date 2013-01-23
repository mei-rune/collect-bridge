local mj = require 'mj'
module("test_enumerate_files",  package.seeall)

function get(params)
	return {value= mj.invoke_native("io_ext.enumerate_files", mj.work_directory .. mj.path_separator .. "modules" ..mj.path_separator .. "enumerate_files")}, nil
end
