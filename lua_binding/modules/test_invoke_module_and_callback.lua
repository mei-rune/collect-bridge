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
    mj.log(mj.SYSTEM, "create result is invode")
	res = mj.execute(params['dumy'], 'create', params)
	  if nil == res then 
    mj.log(mj.SYSTEM, "create result is nil")
  else 
    mj.log(mj.SYSTEM, "create result is:")
    for key, value in pairs(res) do
      mj.log(mj.SYSTEM,  (key or "nil") .. "=" .. (value or "nil"))
    end
    mj.log(mj.SYSTEM, "====== end ======")
  end

	return res
end

function delete(params)
	return mj.execute(params['dumy'], 'delete', params)
end