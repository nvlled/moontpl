local build = {}

---@param link string
---@return nil
function build.queue(link) ---
    --- Adds the link to the build queue.
    ---
    --- The build queue is the list of lua files
    --- that will be rendered to a file when
    --- the build command is run.
    ---
    --- Example:
    ---   In build.queue("/dir/page.html"),
    ---   a file named $SITEDIR/dir/page.html.lua
    ---   will be added to the build queue.
    --- 
end

function build.queueLocalLinks(node) ---
    local build = require "build"
    local query = require "query"
    local path = require "path"
    local links = query.findLocalLinks(node)
    for _, link in ipairs(links) do
        if path.hasParams(link) then
            build.queue(link)
        end
    end
end

return build
