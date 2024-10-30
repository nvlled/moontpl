local query = {}

---@param root html.Node
---@param fn function(html.Node)
function query.eachNode(root, fn)
    --- Recursively visits each node with fn, starting with root node
    local queue = {root}
    while #queue > 0 do
        local node = table.remove(queue)
        if not node then goto continue end

        fn(node)

        for _, child in ipairs(node.children) do
            if child.tag ~= nil then table.insert(queue, child) end
        end
        ::continue::
    end
end

---@param root html.Node
---@param predicate function(html.Node): boolean
---@return html.Node[]
function query.findNodes(root, predicate)
    --- Finds all nodes where predicate(node) is true.
    local queue = {root}
    local result = {}
    while #queue > 0 do
        local node = table.remove(queue)
        if not node then goto continue end

        if predicate(node) then table.insert(result, node) end
        for _, child in ipairs(node.children) do
            if child.tag ~= nil then table.insert(queue, child) end
        end
        ::continue::
    end
    return result
end

---@param node html.Node
---@return string[]
function query.findLocalLinks(node)
    --- Recursively finds all <a> inside node where
    --- its href attribute is a local link.
    local path = require "path"

    if not node then return {} end

    local result = {}
    query.eachNode(node, function(sub)
        if not sub or sub.tag ~= "a" then return end

        local href = sub.attrs.href or ""
        if href:find ".-://" then return end

        table.insert(result, path.absolute(href))
    end)

    return result
end

---@param node html.Node
---@param ... string[]
---@return html.Node
function query.select(node, ...)
    --- Selects a subnode given by the path parameters.
    ---
    --- Example:
    ---     local node = DIV {
    ---         P "paragraph"
    ---         P {
    ---             H1 {
    ---                 "hello"
    ---                 EM "there";
    ---             },
    ---         }
    ---     }
    ---     local small = query.select(node, "p", "h1", "em")
    ---     print(small:tostring())
    ---     -- Output: <em>there</em>
    for _, tag in ipairs(arg) do
        local nextNode = nil
        for _, child in ipairs(node.children or {}) do
            if child.tag == tag then
                nextNode = child
                break
            end
        end
        node = nextNode
        if not node then break end
    end
    return node
end

return query
