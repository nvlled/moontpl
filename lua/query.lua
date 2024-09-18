local query = {}

function query.eachNode(root, fn)
	local queue = { root }
	while #queue > 0 do
		local node = table.remove(queue)
		if not node then
			goto continue
		end

		fn(node)

		for _, child in ipairs(node.children) do
			if child.tag ~= nil then
				table.insert(queue, child)
			end
		end
		::continue::
	end
end

function query.findNodes(root, predicate)
	local queue = { root }
	local result = {}
	while #queue > 0 do
		local node = table.remove(queue)
		if not node then
			goto continue
		end

		if predicate(node) then
			table.insert(result, node)
		end
		for _, child in ipairs(node.children) do
			if child.tag ~= nil then
				table.insert(queue, child)
			end
		end
		::continue::
	end
	return result
end

function query.findLocalLinks(node)
	local path = require "path"

	if not node then
		return {}
	end

	local result = {}
	query.eachNode(node, function(sub)
		if not sub or sub.tag ~= "a" then
			return
		end

		local href = sub.attrs.href or ""
		if href:find ".-://" then
			return
		end

		table.insert(result, path.absolute(href))
	end)

	return result
end

return query
