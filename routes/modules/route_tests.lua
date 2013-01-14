module("route_tests",  package.seeall)


local test_suites= {["test"]= function( ... )
		-- body
	end }

function get(params)
	if "unit_test" == params["target"]	then
		for i,f in pairs(test_suites) do
			f()
			mj.log(mj.DEBUG, i .. " is ok.")
		end
		return "ok", nil
	end
	return "it is unit test for routes.", "it is unit test for routes."
end
