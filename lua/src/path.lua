local env = require "env"
local mod = {}

function mod.getParams(link)
    error("to be implemented by host environment")
end
function mod.setParams(link)
    error("to be implemented by host environment")
end
function mod.hasParams(link)
    error("to be implemented by host environment")
end
function mod.relative(targetLink)
    error("to be implemented by host environment")
end

return mod
