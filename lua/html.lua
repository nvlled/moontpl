local html = {}
local ext = require "ext"
local strict = require "strict"

---@type Node { tag: string, children: (Node|string)[], attrs: { [string]: string } }
--- Node is an object created from the functions DIV, P, H1 and so on.

---@type { [string]: Node }
html.common = {} ---
--- This contains all the predefined, common nodes such
--- as DIV, H1, P, A, and so on.
--- Use html.importGlobals() to import all.
---
--- Example:
---     local DIV = require("html").common.DIV
---     DIV {
---         P "hello"
---     }


local ctorMeta
local nodeMeta

local truncateRest = {}

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
                table.insert(declarations, underscore2Dash(key) .. ": "
                                 .. tostring(value) .. "px")
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
                table.insert(entries, underscore2Dash(k) .. "=" .. '"'
                                 .. attrEscape(styleToString(v)) .. '"')
            elseif type(v) == "boolean" then
                table.insert(entries, underscore2Dash(k))
            else
                table.insert(entries, underscore2Dash(k) .. "=" .. '"'
                                 .. attrEscape(tostring(v)) .. '"')
            end
        end
    end
    return " " .. table.concat(entries, " ")
end

---@return string
function html.textContent(node)
    --- Converts node into an HTML string.
    --- html.textContent(node) is the same as node:tostring()
    if not node then
        return ""
    end

    if getmetatable(node) ~= nodeMeta then
        node = FRAGMENT(node)
    end

    if node.options.tostring then
        return node.options.tostring(node)
    end

    if node.options.selfClosing then
        if not node.children or #node.children == 0 then
            return ""
        end
    end

    return table.concat(ext.map(node.children, function(sub)
        if type(sub) == "string" then
            return node.options.noHTMLEscape and sub or htmlEscape(sub)
        elseif not sub then
            return ""
        end
        return html.textContent(sub)
    end), "")
end

local function nodeToString(node, level)
    if not node or type(node) ~= "table" or not node.tag then
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
            return prefix .. "<" .. tag .. attrsToString(node.attrs) .. "/>"
                       .. suffix
        end
    end

    if not level then
        level = 1
    end

    -- TODO: handle overflow, limited to 3000~ items
    local body = table.concat(ext.map(node.children, function(sub)
        if type(sub) == "string" then
            return node.options.noHTMLEscape and sub or htmlEscape(sub)
        elseif not sub then
            return ""
        end
        return nodeToString(sub, level)
    end), "")

    if node.tag == "" then
        return body
    end

    return prefix .. "<" .. node.tag .. attrsToString(node.attrs) .. ">" .. body
               .. "</" .. node.tag .. ">" .. suffix
end

local appendChild = function(a, b)
    if type(a) == "function" then
        a = a()
    end
    if type(a) == "string" then
        a = DIV(a)
    end
    table.insert(a.children, type(b) == "function" and b() or b)
    return a
end

local appendSibling = function(a, b)
    if not a then
        return b
    end
    if getmetatable(a) ~= nodeMeta then
        a = FRAGMENT(a)
    end
    a.nextSibling = b
    return a
end

nodeMeta = {
    __textContent=html.textContent;
    __tostring=nodeToString;
    __div=appendChild;
    __pow=appendChild;
    __call=function(_, a, b)
        return appendChild(a, b)
    end;
}

local NODESYM = {}

local function _node(tagName, args, options)
    options = options or {}

    if type(args) == "string" then
        local result = {
            tag=tagName;
            attrs={};
            children={args};
            options=options;
            data={};
            parent=nil;
            [NODESYM]=true;
        }
        setmetatable(result, nodeMeta)
        return result
    end

    if getmetatable(args) ~= nil then
        args = {args}
    end

    local attrs = {}
    local data = {}
    local children = {}

    if args["data"] then
        data = args["data"]
        args["data"] = nil
    end

    local result = {
        tag=tagName;
        attrs=attrs;
        children=children;
        options=options;
        data=data;
        parent=nil;
        [NODESYM]=true;
    }
    setmetatable(result, nodeMeta)

    for k, v in pairs(args) do
        if v == truncateRest then
            break
        end

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

                if mt == nodeMeta or v[NODESYM] then
                    v.parent = result
                    table.insert(children, v)
                elseif mt and mt == ctorMeta then
                    local child = v()
                    if getmetatable(child) == nodeMeta or v[NODESYM] then
                        child.parent = result
                    end
                    table.insert(children, child)
                elseif mt and mt.__tostring then
                    table.insert(children, tostring(v))
                elseif type(v) == "table" then
                    for _, c in ipairs(v) do
                        table.insert(children, c)
                    end
                end

                local x = v.nextSibling
                while x do
                    table.insert(children, x)
                    x = x.nextSibling
                end
            elseif v then
                table.insert(children, tostring(v))
            end
        end
    end

    return result
end

ctorMeta = {
    __call=function(self, args)
        return self.ctor(args)
    end;
    __pow=function(self, args)
        return self.ctor(args)
    end;
    __div=function(self, args)
        return self.ctor(args)
    end;
    __idiv=function(self, args)
        return self.ctor(args)
    end;
}

---@param tagName string
---@param options? { selfClosing: string, noHTMLEscape: string }
---@return Node
function html.CreateNode(tagName, options) ---
    --- Defines a node constructor. Common
    --- html elements are already pre-defined,
    --- and can be accessed with require("html").importGlobals()
    --- 
    --- `tagName` is the element name, such as body, or h1.
    --- `options` is an optional table argument that
    --- has the default fields:
    --- {
    ---     selfClosing = false;
    ---     noHTMLEscape = false;
    --- }
    --- 
    --- Example:
    ---     local Node = require("html").CreateNode
    ---     local H1 = Node "h1"
    ---     local EM = Node "em"
    ---     local DIV = Node "div"
    ---     local WIDGET = Node "widget"
    ---     
    ---     print(DIV {
    ---         EM "blah";
    ---         H1 {
    ---             EM "foo";
    ---             "bar";
    ---         };
    ---         WIDGET {}
    ---     })
    ---     -- Outputs:
    ---     <div>
    ---         <em>blah</em>
    ---         <h1>
    ---             <em>foo</em> bar
    ---         </h1>
    ---         <widget></widget>
    ---     </div>
    local ctor = function(args)
        args = args or {}
        if getmetatable(args) == ctorMeta then
            args = args {}
        end
        local result = _node(tagName, args, options)
        return result
    end
    return setmetatable({ctor=ctor}, ctorMeta)
end

function html.Component(ctor, tagName)
    local Comp = html.CreateNode(tagName or "div")
    return setmetatable({
        ctor=function(args)
            return ctor(args)
        end;
    }, ctorMeta)
end

local pp = html.Component(function(args)
    local node = FRAGMENT(args)
    local attrs, children = node.attrs, node.children
    local p = P(attrs)
    local paras = {p}
    
    table.insert(p.children, "\n")

    local function addElem(c)
        local t = type(c)
        if t == "string" then
            local lastLineBlank = false
            for line in ext.lines(c) do
                if ext.trim(line) == "" then
                    if #p.children > 0 and not lastLineBlank then
                        p = P(attrs)
                        table.insert(paras, p)
                        table.insert(p.children, "\n")
                    end
                    lastLineBlank = true
                else
                    if #p.children > 0 and type(p.children[#p.children])
                        == "string" then
                        table.insert(p.children, "\n")
                    end
                    table.insert(p.children, line)
                    lastLineBlank = false
                end
            end
        else
            table.insert(p.children, c)
        end
    end

    for _, c in ipairs(children) do
        addElem(c)
    end

    node.children = paras

    return setmetatable(node, {
        __div=function(a, c)
            addElem(c)
            return a
        end;
    })
end)

local function initCommonNodes()
    local Node = html.CreateNode
    html.common = {
        HTML=Node("html", {prefix="<!DOCTYPE html>"});

        HEAD=Node "head";
        TITLE=Node "title";
        BODY=Node "body";
        SCRIPT=Node("script", {noHTMLEscape=true});
        NOSCRIPT=Node("noscript");
        LINK=Node("link", {selfClosing=true});
        STYLE=Node("style", {noHTMLEscape=true});
        META=Node("meta", {selfClosing=true});

        A=Node "a";
        BASE=Node("base", {selfClosing=true});

        P=Node "p";
        DIV=Node "div";
        SPAN=Node "span";
        PP=pp;

        DETAILS=Node "details";
        SUMMARY=Node "summary";

        B=Node "b";
        I=Node "i";
        EM=Node "em";
        STRONG=Node "strong";
        SMALL=Node "small";
        S=Node "s";
        PRE=Node "pre";
        CODE=Node "code";
        BLOCKQUOTE=Node "blockquote";

        OL=Node "ol";
        UL=Node "ul";
        LI=Node "li";

        FORM=Node "form";
        INPUT=Node("input", {selfClosing=true});
        TEXTAREA=Node "textarea";
        BUTTON=Node "button";
        LABEL=Node "label";
        SELECT=Node "select";
        OPTION=Node "option";

        TABLE=Node "table";
        THEAD=Node "thead";
        TBODY=Node "tbody";
        COL=Node("col", {selfClosing=true});
        TR=Node "tr";
        TD=Node "td";

        SVG=Node "svg";

        BR=Node("br", {selfClosing=true});
        HR=Node("hr", {selfClosing=true});
        NBSP=Node("", {noHTMLEscape=true})("&nbsp;");

        -- lol?
        __=HR;
        ___=HR;
        ____=HR;
        _____=HR;
        ______=HR;
        _______=HR;
        ________=HR;
        _________=HR;
        __________=HR;
        ___________=HR;
        ____________=HR;
        _____________=HR;
        ______________=HR;
        _______________=HR;

        H1=Node "h1";
        H2=Node "h2";
        H3=Node "h3";
        H4=Node "h4";
        H5=Node "h5";
        H6=Node "h6";

        IMG=Node("img", {selfClosing=true});
        AREA=Node("area", {selfClosing=true});

        VIDEO=Node "video";
        IFRAME=Node "iframe";
        EMBED=Node("embed", {selfClosing=true});
        TRACK=Node("track", {selfClosing=true});
        SOURCE=Node("source", {selfClosing=true});

        FRAGMENT=Node "";

        TRUNCATE=truncateRest;

        MARKDOWN=html.toMarkdown;
    }
end

---@return nil
function html.importGlobals()
    --- Adds all pre-defined HTML functions into the global scope.
    --- 
    --- Example:
    ---     -- in web.lua file
    ---     require("html").importGlobals()
    ---     
    ---     -- in example.lua file
    ---     require("web")
    ---     -- functions can now be used without explicity importing
    ---     print(DIV { "blah" })
    strict.disable()
    local Node = html.CreateNode
    for k, v in pairs(html.common) do
        _G[k] = v
    end

    strict.enable()
end
local inlineNodes = {
    a=true;
    b=true;
    i=true;
    s=true;
    span=true;
    small=true;
    em=true;
    strong=true;
    code=true;
    label=true;
}

local function _toMarkdown(node, parent, level, index, preserveLineBreaks)
    -- TODO: shit kludgey code, please refactor later
    if not node then
        return nil
    end
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
        local prevSibling = parent and parent.children[index - 1] or {}
        local isPrevBlock = not (type(prevSibling) == "string"
                                or inlineNodes[prevSibling.tag])
        local isCurBlock = not (type(node) == "string" or inlineNodes[node.tag])

        if isPrevBlock or isCurBlock then
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
                    table.insert(lines, line)
                end
                for j, line in ipairs(lines) do
                    if preserveLineBreaks then
                        line = line .. "\n"
                    else
                        line = line:gsub("([ \t]+)", " ")

                        if node.tag == "p" and line == "" then
                            line = "\n\n" .. ext.trim(line)
                            if j < #lines then
                                lines[j + 1] = ext.trimLeft(lines[j + 1])
                            end
                        end
                    end
                    table.insert(result, line)
                end
            elseif not sub then
                table.insert(result, "")
            else
                table.insert(result, _toMarkdown(sub, node, level, i,
                                                 preserveLineBreaks))
            end
        end
        local v = (table.concat(result, ""))
        return v
    end

    if node.tag == "" then
        return getBody()
    elseif node.tag == "br" then
        return "\n"
    elseif node.tag == "div" then
        return prevNewline() .. getBody()
    elseif node.tag == "p" then
        return prevNewline() .. getBody()
    elseif node.tag == "h1" then
        return prevNewline() .. "# " .. getBody()
    elseif node.tag == "h2" then
        return prevNewline() .. "## " .. getBody()
    elseif node.tag == "h3" then
        return prevNewline() .. "### " .. getBody()
    elseif node.tag == "h4" then
        return prevNewline() .. "#### " .. getBody()
    elseif node.tag == "h5" then
        return prevNewline() .. "##### " .. getBody()
    elseif node.tag == "h6" then
        return prevNewline() .. "######" .. getBody()
    elseif node.tag == "strong" or node.tag == "b" then
        return prevNewline() .. "**" .. getBody() .. "**"
    elseif node.tag == "em" or node.tag == "i" then
        return prevNewline() .. "*" .. getBody() .. "*"
    elseif node.tag == "span" then
        return prevNewline() .. getBody()
    elseif node.tag == "blockquote" then
        return prevNewline() .. "> " .. getBody()
    elseif node.tag == "li" then
        local result = ext.map(node.children, function(sub, i)
            if sub.tag == "ul" or sub.tag == "ol" then
                return prevNewline()
                           .. ext.indent(
                               _toMarkdown(sub, node, level, i,
                                           preserveLineBreaks), (level + 1) * 3)
            else
                if i > 1 and node.children[i - 1].tag == "ul" then
                    return ext.indent(_toMarkdown(sub, node, level,
                                                  preserveLineBreaks),
                                      (level + 1) * 3, i) .. ""
                else
                    return _toMarkdown(sub, node, level, i, preserveLineBreaks)
                               .. " "
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
                return prefix .. dash
                           .. _toMarkdown(sub, node, level, i,
                                          preserveLineBreaks) .. "\n"
            else
                return _toMarkdown(sub, node, level + 1, i, preserveLineBreaks)
                           .. (i < #node.children and " " or "")
            end
        end)
        return ext.trimRight(prevNewline() .. table.concat(result, ""))
    elseif node.tag == "img" then
        return prevNewline() .. "![" .. (node.attrs.alt or "") .. "]("
                   .. node.attrs.src .. ")"
    elseif node.tag == "code" then
        if parent and parent.tag == "pre" then
            local lang = node.data.lang or ""
            return "```" .. lang .. "\n" .. ext.detab(getBody(), "|") .. "\n```"
        end
        return "`" .. getBody() .. "`"
    elseif node.tag == "pre" then
        return prevNewline() .. getBody()
    elseif node.tag == "hr" then
        return prevNewline() .. "-----------------"
    elseif node.tag == "a" then
        local title = node.attrs.title and ' "' .. node.attrs.title .. '"' or ""
        return "[" .. getBody() .. "](" .. node.attrs.href .. title .. ")"
    end

    return tostring(node)
end

---@type node table
---@return string
function html.toMarkdown(node)
    --- Converts a node into markdown string
    local result = ext.trim(_toMarkdown(FRAGMENT(node))):gsub("\n\n\n+", "\n\n")
    return result
end


initCommonNodes()

return html
