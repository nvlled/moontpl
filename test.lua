require "web"
local mymod = require "mymod"

local strict = require("strict").enable()
local path = require "path"

local x= 1

mymod.foo()
mymod.bar()

return DIV {
	_title = "home title",
	_desc = "home desc",
	"blah"
}
