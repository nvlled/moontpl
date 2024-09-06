require "web"

local strict = require("strict").enable()
local path = require "path"

local x= 1

--local mymod = require "mymod"
--mymod.foo()
--mymod.bar()

--table.insert(package.loaders, function(moduleName)
--	print("searching for module:", moduleName)
--end)
print("num loaders", #package.loaders)
print("path", package.path)

require("mymodule")

return DIV {
	_title = "home title",
	_desc = "home desc",
	"blah"
}