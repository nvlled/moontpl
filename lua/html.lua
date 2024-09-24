local ext = require "ext"

local ctorMeta
local nodeMeta

local function trim(s)
	return s:match "^%s*(.-)%s*$"
end

local function tableLen(t)
	if not t then
		return 0
	end
	local count = 0
	for _, _ in pairs(t) do
		count = count + 1
	end
	return count
end

local function underscore2Dash(s)
	local result = string.gsub(s, "_", "-")
	return result
end

local function attrEscape(attr)
	attr = attr:gsub('"', "&quot;")
	attr = attr:gsub("'", "&#39;")
	return attr
end

local function htmlEscape(html)
	html = html:gsub("&", "&amp;")
	html = html:gsub(">", "&gt;")
	html = html:gsub("<", "&lt;")
	return html
end

local function styleToString(t)
	local declarations = {}
	for key, value in pairs(t) do
		if type(key) == "string" then
			if type(value) == "number" then
				table.insert(declarations, underscore2Dash(key) .. ": " .. tostring(value) .. "px")
			else
				table.insert(declarations, underscore2Dash(key) .. ": " .. value)
			end
		else
			error("invalid declaration: " .. tostring(key))
		end
	end

	return table.concat(declarations, "; ")
end

local function attrsToString(attrs)
	if tableLen(attrs) == 0 then
		return ""
	end
	local entries = {}
	for k, v in pairs(attrs) do
		if type(k) == "string" then
			k = attrEscape(underscore2Dash(k))
			if k == "style" and type(v) == "table" then
				table.insert(entries, underscore2Dash(k) .. "=" .. '"' .. attrEscape(styleToString(v)) .. '"')
			elseif type(v) == "boolean" then
				table.insert(entries, underscore2Dash(k))
			else
				table.insert(entries, underscore2Dash(k) .. "=" .. '"' .. attrEscape(tostring(v)) .. '"')
			end
		end
	end
	return " " .. table.concat(entries, " ")
end

local function nodeTextContent(node)
	if not node or not node.tag then
		return ""
	end

	if node.options.tostring then
		return node.options.tostring(node)
	end

	if node.options.selfClosing then
		if not node.children or #node.children == 0 then
			return ""
		end
	end

	return table.concat(
		ext.map(node.children, function(sub)
			if type(sub) == "string" then
				return node.options.noHTMLEscape and sub or htmlEscape(sub)
			elseif not sub then
				return ""
			end
			return nodeToString(sub, level)
		end),
		""
	)
end

local function nodeToString(node, level)
	if not node or not node.tag then
		return ""
	end

	if node.options.tostring then
		return node.options.tostring(node)
	end

	local prefix = node.options.prefix or ""
	local suffix = node.options.suffix or ""

	if node.options.selfClosing then
		if not node.children or #node.children == 0 then
			local tag = node.tag or ""
			return prefix .. "<" .. tag .. attrsToString(node.attrs) .. "/>" .. suffix
		end
	end

	if not level then
		level = 1
	end

	-- TODO: handle overflow, limited to 3000~ items
	local body = table.concat(
		ext.map(node.children, function(sub)
			if type(sub) == "string" then
				return node.options.noHTMLEscape and sub or htmlEscape(sub)
			elseif not sub then
				return ""
			end
			return nodeToString(sub, level)
		end),
		""
	)

	if node.tag == "" then
		return body
	end

	return prefix .. "<" .. node.tag .. attrsToString(node.attrs) .. ">" .. body .. "</" .. node.tag .. ">" .. suffix
end

local appendChild = function(a, b)
	if type(a) == "function" then
		a = a()
	end
	table.insert(a.children, type(b) == "function" and b() or b)
	return a
end

nodeMeta = {
	__textContent = nodeTextContent,
	__tostring = nodeToString,
	__div = appendChild,
	__pow = appendChild,
}

local function _node(tagName, args, options)
	options = options or {}

	if type(args) == "string" then
		local result = { tag = tagName, attrs = {}, children = { args }, options = options }
		setmetatable(result, nodeMeta)
		return result
	end

	if getmetatable(args) ~= nil then
		args = { args }
	end

	local attrs = {}
	local data = {}
	local children = {}

	if args["data"] then
		data = args["data"]
		args["data"] = nil
	end

	for k, v in pairs(args) do
		if type(k) == "string" then
			if k:sub(1, 2) == "--" then
				options[k:sub(3)] = v
			elseif k:sub(1, 1) == "_" then
				data[k:sub(2)] = v
			else
				attrs[k] = v
			end
		elseif type(k) == "number" then
			while type(v) == "function" and getmetatable(v) ~= ctorMeta do
				v = v()
			end

			if type(v) == "string" then
				table.insert(children, v)
			elseif type(v) == "table" then
				local mt = getmetatable(v)

				if mt == nodeMeta then
					table.insert(children, v)
				elseif mt and mt == ctorMeta then
					table.insert(children, v())
				elseif mt and mt.__tostring then
					table.insert(children, tostring(v))
				else
					error "plain table cannot be a child node"
				end
			elseif v then
				table.insert(children, tostring(v))
			end
		end
	end

	local result = { tag = tagName, attrs = attrs, children = children, options = options, data = data }
	setmetatable(result, nodeMeta)

	return result
end

ctorMeta = {
	__call = function(self, args)
		return self.ctor(args)
	end,
	__pow = function(self, args)
		return self.ctor(args)
	end,
	__div = function(self, args)
		return self.ctor(args)
	end,
	__idiv = function(self, args)
		return self.ctor(args)
	end,
}

local function Node(tagName, options)
	local ctor = function(args)
		args = args or {}
		if getmetatable(args) == ctorMeta then
			args = args {}
		end
		local result = _node(tagName, args, options)
		return result
	end
	return setmetatable({ ctor = ctor }, ctorMeta)
end

local function importGlobals()
	HTML = Node("html", { prefix = "<!DOCTYPE html>" })

	HEAD = Node "head"
	TITLE = Node "title"
	BODY = Node "body"
	SCRIPT = Node("script", { noHTMLEscape = true })
	LINK = Node("link", { selfClosing = true })
	STYLE = Node("style", { noHTMLEscape = true })
	META = Node("meta", { selfClosing = true })

	A = Node "a"
	BASE = Node("base", { selfClosing = true })

	P = Node "p"
	DIV = Node "div"
	SPAN = Node "span"

	DETAILS = Node "details"
	SUMMARY = Node "summary"

	B = Node "b"
	I = Node "i"
	EM = Node "em"
	STRONG = Node "strong"
	SMALL = Node "small"
	S = Node "s"
	PRE = Node "pre"
	CODE = Node "code"
	BLOCKQUOTE = Node "blockquote"

	OL = Node "ol"
	UL = Node "ul"
	LI = Node "li"

	FORM = Node "form"
	INPUT = Node("input", { selfClosing = true })
	TEXTAREA = Node "textarea"
	BUTTON = Node "button"
	LABEL = Node "label"
	SELECT = Node "select"
	OPTION = Node "option"

	TABLE = Node "table"
	THEAD = Node "thead"
	TBODY = Node "tbody"
	COL = Node("col", { selfClosing = true })
	TR = Node "tr"
	TD = Node "td"

	SVG = Node "svg"

	BR = Node("br", { selfClosing = true })
	HR = Node("hr", { selfClosing = true })

	H1 = Node "h1"
	H2 = Node "h2"
	H3 = Node "h3"
	H4 = Node "h4"
	H5 = Node "h5"
	H6 = Node "h6"

	IMG = Node("img", { selfClosing = true })
	AREA = Node("area", { selfClosing = true })

	VIDEO = Node "video"
	IFRAME = Node "iframe"
	EMBED = Node("embed", { selfClosing = true })
	TRACK = Node("track", { selfClosing = true })
	SOURCE = Node("source", { selfClosing = true })

	FRAGMENT = Node ""
end

local inlineNodes = {
	a = true,
	b = true,
	i = true,
	s = true,
	span = true,
	small = true,
	em = true,
	strong = true,
	img = true,
	code = true,
	label = true,
}

-- TODO: shit kludgey code, please refactor later
function toMarkdown(node, parent, level, index, preserveLineBreaks)
	if node.tag == "pre" then
		preserveLineBreaks = true
	end

	if not level then
		level = 0
	end
	if not node then
		return ""
	end

	local function prevNewline()
		local prevSibling = parent and parent.children[index - 1]
		if not prevSibling then
			return ""
		end
		if type(prevSibling) == "string" or inlineNodes[prevSibling.tag] then
			return "\n\n"
		end
		return ""
	end

	local function getBody()
		local result = {}
		for i, sub in ipairs(node.children) do
			if type(sub) == "string" then
				local lines = {}
				for line in ext.lines(sub) do
					line = ext.trimLeft(line)
					table.insert(lines, line)
				end
				for j, line in ipairs(lines) do
					if j > 1 and i > 1 and j < #lines then
						line = " " .. line
					end
					if preserveLineBreaks then
						line = line .. "\n"
					end
					table.insert(result, line)
				end
			elseif not sub then
				table.insert(result, "")
			else
				table.insert(result, toMarkdown(sub, node, level, i, preserveLineBreaks))
			end
		end
		return table.concat(result, "")
	end

	if node.tag == "" then
		return getBody()
	elseif node.tag == "br" then
		return ""
	elseif node.tag == "div" then
		return getBody() .. "\n\n"
	elseif node.tag == "p" then
		return getBody() .. "\n\n"
	elseif node.tag == "h1" then
		return "# " .. getBody() .. "\n\n"
	elseif node.tag == "h2" then
		return "## " .. getBody() .. "\n\n"
	elseif node.tag == "h3" then
		return "### " .. getBody() .. "\n\n"
	elseif node.tag == "h4" then
		return "#### " .. getBody() .. "\n\n"
	elseif node.tag == "h5" then
		return "##### " .. getBody() .. "\n\n"
	elseif node.tag == "h6" then
		return "######" .. getBody() .. "\n\n"
	elseif node.tag == "strong" or node.tag == "b" then
		return "**" .. getBody() .. "**" .. " "
	elseif node.tag == "em" or node.tag == "i" then
		return "*" .. getBody() .. "*" .. " "
	elseif node.tag == "blockquote" then
		return "> " .. getBody() .. "\n\n"
	elseif node.tag == "li" then
		local result = ext.map(node.children, function(sub, i)
			if sub.tag == "ul" or sub.tag == "ol" then
				return "\n" .. ext.indent(toMarkdown(sub, node, level, i, preserveLineBreaks), (level + 1) * 3)
			else
				if i > 1 and node.children[i - 1].tag == "ul" then
					return ext.indent(toMarkdown(sub, node, level, preserveLineBreaks), (level + 1) * 3, i)
				else
					return toMarkdown(sub, node, level, i, preserveLineBreaks) .. " "
				end
			end
		end)
		return table.concat(result, "")
	elseif node.tag == "ul" or node.tag == "ol" then
		local ordered = node.tag == "ol"
		local result = ext.map(node.children, function(sub, i)
			if sub.tag == "li" then
				local prefix = ""
				if i > 1 and node.children[i - 1].tag ~= "li" then
					prefix = "\n"
				end
				local dash = ordered and "1. " or "-  "
				return prefix .. dash .. toMarkdown(sub, node, level, i, preserveLineBreaks) .. "\n"
			else
				return toMarkdown(sub, node, level + 1, i, preserveLineBreaks) .. (i < #node.children and " " or "")
			end
		end)
		return prevNewline() .. table.concat(result, "")
	elseif node.tag == "img" then
		return "![" .. node.attrs.alt .. "](" .. node.attrs.src .. ")" .. " "
	elseif node.tag == "code" then
		if parent and parent.tag == "pre" then
			return "```\n" .. ext.detab(getBody(), "| ") .. "\n````"
		end
		return "`" .. getBody() .. "`"
	elseif node.tag == "pre" then
		return getBody()
	elseif node.tag == "hr" then
		return prevNewline() .. "***" .. "\n\n"
	elseif node.tag == "a" then
		local title = node.attrs.title and ' "' .. node.attrs.title .. '"' or ""
		return "[" .. getBody() .. "](" .. node.attrs.href .. title .. ")" .. " "
	end

	return tostring(node)
end

return {
	Node = Node,
	toMarkdown = toMarkdown,
	importGlobals = importGlobals,
}
