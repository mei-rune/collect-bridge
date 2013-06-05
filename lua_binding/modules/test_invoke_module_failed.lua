module("test_invoke_module_failed",  package.seeall)

function get(params)
	return {error_message= "get error for test_invoke_module_failed"}
end

function put(params)
	return {error_message= "put error for test_invoke_module_failed"}
end

function create(params)
	return {error_message="record not found"}
end

function delete(params)
	return {error_message="delete failed"}
end