local mj = require 'mj'
module("routes",  package.seeall)


local route = ml.class()
function route.equal(key, value)

end

mj.route = function()
	return route()
end