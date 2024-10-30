local page = {}

---@type {[string]: any}
page.input = {} ---
--- `page.input` is used when a lua html page is used as a
--- HTML/CSS templating engine. The page
--- can access the page.input to alter the output.

---@type {[string]: any}
page.data = {} ---
--- `page.data` is set by each lua page.
--- This serves as a metadata about the page itself.
--- The data set here is included in the page.list().
---
--- Example:
---   -- in index.html.lua --
---   local page = require("page")
---   page.data.title = "Home page"
---   page.data.desc = "Welcome"

---@type string
page.PAGE_LINK = "" ---
--- The link to the page currently being run or rendered.
--- 
--- Example:
--- -- in /home/mysite/subdir/page.html.lua
--- require("page").PAGE_LINK == "/subdir/page.html" -- true

---@type string
page.PAGE_FILENAME = "" ---
--- The filename of the page currently being run or rendered.
--- Example:
--- -- in /home/mysite/subdir/page.html.lua
--- require("page").PAGE_LINK == "/home/mysite/subdir/page.html.lua" -- true

---@type PageEntry {absFile: string, relFile: string link: string: data: table}
---@return PageEntry[]
function page.list() ---
    --- Returns the list of pages found in the SITEDIR.
    --- PageEntry.data contains the data set by page.data.
    -- stub
    return {}
end

---@return string[]
function page.files() ---
    --- Returns the list of filenames found in the SITEDIR.
    -- stub
    return {}
end

return page
