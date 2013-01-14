module("test_clean_path",  package.seeall)

function get(params)
	mj.log(mj.DEBUG, "WD = " .. mj.work_directory)
	return mj.clean_path(params['path'])
end
