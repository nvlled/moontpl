local env = require "env"
local mod = {}

-- stub functions, to be implemented by host environment

function mod.getParams(link) return {} end

function mod.setParams(link) end

function mod.hasParams(link) return false end

function mod.relative(targetLink) return "" end

return mod
