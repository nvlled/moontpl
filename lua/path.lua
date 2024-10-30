local path = {}

---@param link string
---@return { [string]: string }
function path.getParams(link) ---
    --- Params of a link or filename is similar to query parameters of a URL,
    --- but with different syntax. 
    --- Example:
    ---    /greet[name=yourname].html
    ---    /showimage[filename=cat.jpg,size=300].html
    ---
    --- Gets the params of a link/filename.
    --- example:
    ---   local params = path.getParams("/showimage[filename=cat.jpg,size=300].html")
    ---   -- params == {filename="cat.jpg", size="300"}
    -- stub
    return {}
end

---@param link string
---@param params { [string]: string }
---@param clear? boolean
function path.setParams(link, params, clear) ---
    --- Adds the params to a link/filename.
    --- If clear is true, then any exiting params will
    --- be removed.
    ---
    --- Example:
    ---     path.setParams("/page[a=1,b=2].html", {c=3,b=22})       
    ---         == "/page[a=1,b=22,cc=3].html"
    ---
    ---     path.setParams("/page[a=1,b=2].html", {c=3,b=22}, true) 
    ---         == "/page[b=22,cc=3].html"
    -- stub
    return ""
end

---@return bool
function path.hasParams(link) ---
    --- Returns true if link has params.
    -- stub
    return false
end

---@param link string
---@return string
function path.relative(link) ---
    --- Converts the link to a relative link (relative to current page).
    --- Example:
    ---     -- current page is /dir/index.html
    ---     path.relative("/file.jpg") == "../file.jpg"
    -- stub
    return ""
end

---@param link string
---@return string
function path.absolute(link) ---
    --- Converts the link to absolute link
    --- Example:
    ---     -- current page is /dir/index.html
    ---     path.absolute("./file.jpg") == "/dir/file.jpg"
    -- stub
    return ""
end

return path
