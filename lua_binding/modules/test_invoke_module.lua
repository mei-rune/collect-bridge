local mj = require 'mj'
module("test_invoke_module",  package.seeall)

function get(params)
	return {value= "get test ok test1whj23"}, nil
end

function put(params)
	return {value= "put test ok test1whj23"}, nil
end

function create(params)
	return {value= "2328"}, "create test ok test1whj23"
end

function delete(params)
	return {value=false}, "delete test ok test1whj23"
end