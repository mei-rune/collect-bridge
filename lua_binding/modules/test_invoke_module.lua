local mj = require 'mj'
module("test_invoke_module",  package.seeall)

function get(params)
	return {value= "get test ok test1whj23"}
end

function put(params)
	return {value= "put test ok test1whj23"}
end

function create(params)
	return {value= "2328", error_message= "create test ok test1whj23"}
end

function delete(params)
	return {value=false , error_message= "delete test ok test1whj23"}
end