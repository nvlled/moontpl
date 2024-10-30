local site = {}

---@param options? { dir: string, lua: boolean, filter: function(string):boolean }
---@return string[]
function site.files(options) ---
    --- List the files found in SITEDIR.
    --- The returned filenames are absolute URL paths.
    --- The options defaults to:
    ---  {
    ---      dir = "/",
    ---      filter = function(pathname) return true end,
    ---      lua = false,
    ---  }
    ---
    --- Example:
    --- Suppose SITEDIR contains the following files:
    --- | index.html.lua
    --- | dir1
    --- | dir1/sample.html.lua
    --- Then
    ---  site.files()              -- returns {"/index.html", "/dir1/sample.html"}
    ---  site.files{"dir="/dir1"}  -- returns { "/dir1/sample.html"}
    ---  site.files{filter = function(p)
    ---      return p:find("index") ~= nil
    ---  end }                     -- returns {"/index.html"}
    return {}
end

return site