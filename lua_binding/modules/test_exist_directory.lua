module("test_exist_directory",  package.seeall)

function get(params)
	return {value= mj.directory_exists(mj.work_directory .. mj.path_separator .. "modules" ..mj.path_separator .. params["path"])}, nil
end
