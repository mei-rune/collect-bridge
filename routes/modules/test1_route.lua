

-- mj.log(mj.SYSTEM, "dump _ENV")
-- for i, s in pairs(_ENV) do
--   if type(s) == "function" then
--   	mj.log(mj.SYSTEM, i .. "= function " .. table.concat(debug.getinfo(s), ", "))
--   elseif type(s) == "table" then
--   	mj.log(mj.SYSTEM, i .. "=" .. table.concat(s, "," ))
--   else
--   	mj.log(mj.SYSTEM, i .. "=" .. s)
--   end
-- end


name = 'interface'

description = [[ xxxxxxx
]]

author = "Diman Todorov"

license = "Same as Nmap--See http://nmap.org/book/man-legal.html"

level = {"system", "12"}

categories = {"default", "safe"}


-- 1. 调试模式下, 会输出每一个插件为什么不匹配
-- 2. 系统会调整匹配条件次序, 如 equal 可能将先于其它操作执行, 每个插件的匹配执行顺序是不确定的.
match = route():equal("ss", "equal(ss)"):start_with("tt","start_with(tt)"):end_with("aa", "end_with(aa)"):contains("cc", "contains(cc)"):match("aa", "match(aa{0, 3})") 

-- :and_with(function(param)
-- end)



action = get {
	schema= "snmp",
	action= "table",
	oid= "1.3.6.7",
}

-- or 
-- action = function(param) {
-- }