local mj = require 'mj'
module("test_invoke_module_and_callback",  package.seeall)

function get(params)
	mj.log(mj.DEBUG, "this a test log for test_invoke_module_and_callback")
	return mj.execute(params['dumy'], 'get', params)
end

function put(params)
	return mj.execute(params['dumy'], 'put', params)
end

function create(params)
	return mj.execute(params['dumy'], 'create', params)
end

function delete(params)
	return mj.execute(params['dumy'], 'delete', params)
end