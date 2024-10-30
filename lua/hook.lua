local hook = {}

---@type function(html.Node): html.Node
function hook.onPageRender(node)
    --- This function will be called when a page
    --- is rendered, right before it is convereted to string.
    --- Node is an HTML node (see html module). The node
    --- can be modified directly, or the function can return
    --- a new modified copy.
    --- 
    --- Example:
    ---     local hook = require("hook")
    ---     hook.onPageRender = function(node)
    ---         -- add the text "hello" inside node
    ---         table.insert(node.children, "hello")
    ---         return node
    ---     end

    -- stub
end

return hook
