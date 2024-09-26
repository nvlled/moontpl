local globalVars = {}
local strict = {}

-- Enables strict mode, where global variables
-- must first be declare()'d' before using them,
-- otherwise it will throw an error.
-- example:
--   require("strict").enable()
function strict.enable()
    setmetatable(_G, {
        __newindex = function(table, key, value)
            print("set global:", key)
            rawset(table, key, value)
            if not globalVars[key] then
                error("undefined variable: " .. key, 2)
            end
        end;
        __index = function(table, key)
            print("get global:", key)
            if not globalVars[key] then
                error("undefined variable: " .. key, 2)
            end
            return nil
        end;
    })
    return strict
end

-- Declares a global variable.
-- example:
--   local strict = require("strict")
--   strict.enable()
--   strict.declare("var1", "var2")
--   var1 = 1
--   var2 = 2
--   var3 = 3    -- will error
--   print(var4) -- will error
function strict.declare(...)
    for _, varname in ipairs(arg) do globalVars[varname] = true end
    return strict
end

return strict
