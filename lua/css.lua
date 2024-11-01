local strict = require("strict")
local ext = require "ext"

local css = {}

local function underscore2Dash(s)
    return string.gsub(s, "_", "-")
end

local function sortTable(t)
    local result = {}
    for k, v in pairs(t) do
        table.insert(result, { k = k, v = v })
    end
    table.sort(result, function(a, b)
        return a.k < b.k
    end)
    return result
end

local function cssToString(ruleset)
    local buffer = {}
    local n = #ruleset

    for _, rule in ipairs(ruleset) do
        if not rule or rule.selector == "" or ext.len(rule.declarations) == 0 then
            goto continue
        end

        local indent = ""
        if rule.mediaQuery then
            indent = "  "
            table.insert(buffer, "@media " .. rule.mediaQuery .. "{\n")
        end

        table.insert(buffer, indent .. ext.trim(rule.selector) .. " {\n")

        for _, e in ipairs(sortTable(rule.declarations)) do
            local decl = "  " .. e.k .. ": " .. e.v .. ";\n"
            table.insert(buffer, indent .. decl)
        end

        table.insert(buffer, indent .. "}\n")

        if rule.mediaQuery then
            table.insert(buffer, "}\n")
        end

        ::continue::
    end

    return table.concat(buffer, "")
end

local function mediaToString(media)
    local buffer = {}
    for _, ruleset in ipairs(media.rulesets) do
        for line in ext.split(tostring(ruleset), "\n") do
            table.insert(buffer, "  " .. line)
        end
    end
    if #buffer > 0 then
        buffer[#buffer] = ext.trim(buffer[#buffer])
    end

    return "@media " .. media.types .. " {\n" .. table.concat(buffer, "\n")
        .. "}"
end

local cssMeta = { __tostring = cssToString, __textContent = cssToString }

local cssMediaMeta = { __tostring = mediaToString, __textContent = mediaToString }

local function appendSelector(parent, child, nospace)
    local sep = nospace and "" or " "
    if not parent:find "," and not child:find "," then
        return parent .. sep .. child
    end

    local xs = {}
    for h in ext.split(parent, ",") do
        h = h:match "^%s*(.-)%s*$"
        for k in ext.split(child, ",") do
            k = k:match "^%s*(.-)%s*$"
            table.insert(xs, h .. sep .. k)
        end
    end

    return table.concat(xs, ", ")
end

local function _CSS(args, selector)
    if not selector then
        selector = ""
    end

    local rule = { mediaQuery = nil, selector = selector, declarations = {} }
    local result = { type = "css", rule }
    local subRules = {}

    for key, value in pairs(args) do
        if type(key) == "string" then
            if type(value) == "table" then
                local subRules =
                    _CSS(value, appendSelector(selector, key, true))
                for _, s in ipairs(subRules) do
                    table.insert(result, s)
                end
            elseif type(value) == "number" then
                rule.declarations[underscore2Dash(key)] =
                    tostring(value) .. "px"
            elseif key == "@media" then
                rule.mediaQuery = tostring(value)
            else
                rule.declarations[underscore2Dash(key)] = tostring(value)
            end
        elseif type(key) == "number" and type(value) == "table" then
            table.insert(subRules, value)
        else
            error "invalid declaration"
        end
    end

    for _, value in ipairs(subRules) do
        if getmetatable(value) == cssMeta then
            for _, rule in ipairs(value) do
                rule.selector = appendSelector(selector, rule.selector)
                table.insert(result, rule)
            end
        elseif getmetatable(value) == cssMediaMeta then
            for _, ruleset in ipairs(value.rulesets) do
                for _, rule in ipairs(ruleset) do
                    rule.selector = appendSelector(selector, rule.selector)

                    if rule.mediaQuery then
                        rule.mediaQuery =
                            value.types .. " and " .. rule.mediaQuery
                    else
                        rule.mediaQuery = value.types
                    end

                    table.insert(result, rule)
                end
            end
        else
            for k, v in pairs(value) do
                if type(v) == "number" then
                    rule.declarations[underscore2Dash(k)] = tostring(v) .. "px"
                else
                    rule.declarations[underscore2Dash(k)] = v
                end
            end
        end
    end

    return result
end

---@param selector string
---@return function(table): CSSNode
function css.CSS(selector)
    --- Defines a CSS node represents a CSS rule.
    --- It returns a CSS when converted with tostring().
    ---
    --- Example:
    ---     local CSS = require("css").CSS
    ---     local c = CSS 'body' {
    ---         color="blue";
    ---         background_color="#000";
    ---     }
    ---     print(c:tostring())
    ---     -- Output:
    ---     -- body {
    ---     --   color: blue;
    ---     --   background-color: #000;
    ---     -- }
    if type(selector) == "table" then
        local style = _CSS(selector, "")
        setmetatable(style, cssMeta)
        return style
    end

    return function(args)
        local style = _CSS(args, selector)
        setmetatable(style, cssMeta)

        return style
    end
end

---@param types string
---@return function(table): table
function css.CSS_MEDIA(types)
    --- Returns an object that represents defintions
    --- for media queries.
    ---
    --- Example:
    --- local CSS_MEDIA = require("css").CSS_MEDIA
    --- CSS_MEDIA '(orientation: portrait)' {
    ---    CSS ".post.popup" {
    ---        max_width = '80vw !important',
    ---    }
    --- }
    return function(args)
        local media = { types = types, rulesets = args }
        setmetatable(media, cssMediaMeta)
        return media
    end
end

---@return nil
function css.importGlobals()
    --- Adds all exported module functions into the global scope.
    ---
    --- Example:
    ---     -- in web.lua file
    ---     require("css").importGlobals()
    ---
    ---     -- in example.lua file
    ---     require("web")
    ---     -- CSS function can now be used without explicity importing
    ---     CSS '.test' { }
    strict.disable()
    CSS = css.CSS
    CSS_MEDIA = css.CSS_MEDIA
    strict.enable()
end

return css
