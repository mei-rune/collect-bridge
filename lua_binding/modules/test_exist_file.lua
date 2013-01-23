module("test_exist_file",  package.seeall)

function get(params)
	return {value= mj.file_exists(mj.work_directory .. mj.path_separator .. "modules" ..mj.path_separator .. "enumerate_files"..mj.path_separator .. params["file"])}, nil
end
