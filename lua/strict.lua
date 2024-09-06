local globalVars = {}
local mod = {}

function mod.declare(...)
	for _, varname in ipairs(arg) do
		globalVars[varname] = true
	end
	return mod
end

function mod.enable()
	setmetatable(_G, {
		__newindex = function(table, key, value)
			print("set global:", key)
			rawset(table, key, value)
			if not globalVars[key] then
				error("undefined variable: " .. key, 2)
			end
		end,
		__index = function(table, key)
			print("get global:", key)
			if not globalVars[key] then
				error("undefined variable: " .. key, 2)
			end
			return nil
		end,
	})
	return mod
end

return mod
