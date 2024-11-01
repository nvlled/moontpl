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

---@param node html.Node
---@return node html.Node|nil
function page.onRender(node)
    --- This function is called when a lua page is
    --- being rendered or converted to an HTML string.
    --- Specifically, it is called right before node:tostring()
    --- is called.
    --- The purpose of this function is to make changes to
    --- the node before it is rendered.
    page.onRenderDefault(node)
end

---@param node html.Node
---@return node html.Node|nil
function page.onRenderDefault(node)
    --- This function is the default implementation
    --- of page.onRender. It will
    --- be called if page.onRender is nil.
    --- This function is not intended to be called
    --- directly.
    ---[[
    -- default implementation
    local tags = require("runtags")
    local build = require("build")
    if tags.serve then page.appendReloadScript(node) end
    if tags.build then build.queueLocalLinks(node) end
    ---]]
end

---@param node html.Node
---@return nil
function page.appendReloadScript(node)
    --- Appends a reload <script> inside node.
    --- This script reloads the page when a lua file
    --- has been modified or created.
    --- This function is intented to be called inside page.onRender.
    --- 
    --- Example:
    ---     local page = require("page")
    ---     page.onRender = function(node)
    ---         page.appendReloadScript(node)
    ---     end

    local query = require "query"
    local body = query.select(node, "body") or node
    local contents = [[
        window.addEventListener("load", function() {
            new EventSource('/.modified').onmessage = function(event) {
                window.location.reload();
            };
        });
    ]]
    if body.children then
        table.insert(body.children, SCRIPT(contents))
    end
end

return page
