local path = {}

-- PathParams is similar to query parameters of URL,
-- but with different syntax. 
-- example:
--    /greet[name=yourname].html
--    /showimage[filename=cat.jpg,size=300].html

-- Gets the params of a link/filename.
-- example:
--   local params = path.getParams("/showimage[filename=cat.jpg,size=300].html")
--   -- params == {filename="cat.jpg", size="300"}
function path.getParams(link)
    -- stub
    return {}
end

-- Adds the params to a link/filename.
-- If clear is true, then any exiting params will
-- be removed.
-- example:
--   path.setParams("/page[a=1,b=2].html", {c=3,b=22})       == "/page[a=1,b=22,cc=3].html"
--   path.setParams("/page[a=1,b=2].html", {c=3,b=22}, true) == "/page[b=22,cc=3].html"
function path.setParams(link, params, clear)
    -- stub
    return ""
end

-- Returns true if link has params.
function path.hasParams(link)
    -- stub
    return false
end

-- Converts the link to a relative link (relative to current page).
-- example:
--   -- current page is /dir/index.html
--   path.relative("/file.jpg") == "../file.jpg"
function path.relative(link)
    -- stub
    return ""
end

-- Converts the link to absolute link
-- example:
--   -- current page is /dir/index.html
--   path.absolute("./file.jpg") == "/dir/file.jpg"
function path.absolute(link)
    -- stub
    return ""
end

return path
