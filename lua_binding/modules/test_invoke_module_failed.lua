module("test_invoke_module_failed",  package.seeall)

function get(params)
	return nil, "get error for test_invoke_module_failed"
end

function put(params)
	return nil, "put error for test_invoke_module_failed"
end

function create(params)
	return nil, "record not found"
end

function delete(params)
	return nil, "delete failed"
end