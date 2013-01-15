
-- name = 'interface'

description = [[ xxxxxxx
]]

for i, s in pairs(_ENV) do
  if type(s) == "function" then
  print(i .. "=" .. s())
  else
  print(i .. "=" .. s)
  end
end

author = "Diman Todorov"

license = "Same as Nmap--See http://nmap.org/book/man-legal.html"

level = {"system", 12}

categories = {"default", "safe"}


-- 1. 调试模式下, 会输出每一个插件为什么不匹配
-- 2. 系统会调整匹配条件次序, 如 equal 可能将先于其它操作执行, 每个插件的匹配执行顺序是不确定的.
match = route():equal("ss", ""):start_with("tt",""):end_with("aa", "bb"):contains("", ""):match("aa{0, 3}"):and_with(function(param)
end)



action = table {
  
}
-- or 
-- action = function(param) {
-- }