local mj = require 'mj'
module("test_ok_init",  package.seeall)

function get(params)
	return {value= "test init ok"}, nil
end
