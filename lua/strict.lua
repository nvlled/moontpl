local strict = {}
local globalVars = {}
local disabled = false
local initialized = false

---@return nil
function strict.enable() ---
    --- Enables strict mode, where global variables
    --- must first be declare()'d' before using them,
    --- otherwise it will throw an error.
    --- 
    --- Example:
    ---   require("strict").enable()
    disabled = false
    if not initialized then
        setmetatable(_G, {
            __newindex=function(table, key, value)
                if not globalVars[key] and not disabled then
                    error("undefined variable: " .. key, 2)
                end
                rawset(table, key, value)
            end;
            __index=function(table, key)
                if not globalVars[key] and not disabled then
                    error("undefined variable: " .. key, 2)
                end
                return nil
            end;
        })
        initialized = true
    end
    return strict
end

---@return nil
function strict.disable() ---
    --- Disables strict mode.
    --- 
    --- Example:
    ---   require("strict").disable()
    disabled = true
end

---@param ... string[]
function strict.declare(...) ---
    --- Declares global variables.
    ---
    --- Example:
    ---   local strict = require("strict")
    ---   strict.enable()
    ---   strict.declare("var1", "var2")
    ---   var1 = 1
    ---   var2 = 2
    ---   var3 = 3    -- will error
    ---   print(var4) -- will error
    for _, varname in ipairs(arg) do
        globalVars[varname] = true
    end
    return strict
end

return strict
