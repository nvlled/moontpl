local hook = {}

-- This function will be called when a page
-- is rendered, right before it is convereted to string.
-- Node is an HTML node (see html module). The node
-- can be modified directly, or the function can return
-- a new modified copy.
function hook.onPageRender(node)
    -- stub
end

return hook
