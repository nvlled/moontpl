local P = {}

function P.trim(s)
    if not s then
        return ""
    end
    return s:match "^%s*(.-)%s*$"
end

function P.trimRight(s)
    if not s then
        return ""
    end
    return s:match "^(.-)%s*$"
end
function P.trimLeft(s)
    if not s then
        return ""
    end
    return s:match "^%s*(.-)$"
end

function P.dirPath(path)
    local i = 1
    while true do
        local j = path:find("/", i)
        if not j then
            break
        end
        i = j + 1
    end
    return path:sub(1, i - 2), path:sub(i)
end

function P.isEmptyString(str)
    return not str or str == ""
end

function P.rep(str, n)
    if type(n) ~= "number" then
        error "repeat second parameter must be a number"
    end

    local result = {}
    for i = 1, n do
        table.insert(result, str)
    end
    return table.concat(result, "")
end

function P.relativePath(targetPath, srcPath)
    if targetPath:sub(1, 1) ~= "/" or not srcPath or srcPath == "" then
        return targetPath
    end

    local slashCount = 0
    local i = 1
    while true do
        local j = srcPath:find("/", i)
        if not j then
            break
        end
        slashCount = slashCount + 1
        i = j + 1
    end

    return string.rep("../", slashCount - 1) .. targetPath:sub(2)
end

function P.filter(t, pred)
    local result = {}
    for i, v in pairs(t) do
        if pred(v, i) then
            table.insert(result, v)
        end
    end
    return result
end

function P.each(xs, fn)
    for i, v in pairs(t) do
        fn(v, i)
    end
end

function P.sortedIter(m, fn)
    local keys = {}
    for k in pairs(m) do
        table.insert(keys, k)
    end
    table.sort(keys)
    for _, k in ipairs(keys) do
        fn(m[k], k)
    end
end

function P.map(t, fn)
    local result = {}
    for i, v in pairs(t) do
        table.insert(result, fn(v, i))
    end
    return result
end

function P.mapSlice(si, sj, t, fn)
    local result = {}
    for i, v in pairs(t) do
        if i >= si then
            table.insert(result, fn(v, i))
        end
        if i >= sj then
            goto finish
        end
    end
    ::finish::
    return result
end

function P.slice(t, from, to)
    local result = {}
    for i = from, to, 1 do
        table.insert(result, t[i])
    end
    return result
end

function P.len(t)
    local count = 0
    for _ in pairs(t) do
        count = count + 1
    end
    return count
end

function P.split(inputstr, sep)
    local i = 1
    return function()
        local a, b = inputstr:find(sep, i)
        if i then
            if not a then
                local s = inputstr:sub(i, -1)
                i = nil
                return s
            else
                local s = inputstr:sub(i, a - 1)
                i = b + 1
                return s
            end
        end
    end
end

function P.endsWith(s, suffix)
    return s:sub(-#suffix) == suffix
end

function P.contains(t, elem)
    for _, x in ipairs(t) do
        if x == elem then
            return true
        end
    end
    return false
end

function P.getFileExt(s)
    local lastDotIndex = -1
    local i = #s
    while i > 1 do
        if s:sub(i, i) == "." then
            lastDotIndex = i
        end
        i = i - 1
    end
    if lastDotIndex > 0 then
        return s:sub(lastDotIndex, -1)
    end

    return ""
end

function P.reverse(xs)
    local result = {}
    for i = #xs, 1, -1 do
        result[#xs - i + 1] = xs[i]
    end
    return result
end

function P.alt(x, y)
    if not x or x == "" then
        return y
    end
    return x
end

function P.parseDateTime(dateTimeStr)
    if not dateTimeStr then
        return nil
    end

    local i = dateTimeStr:find " " or #dateTimeStr + 1
    local dateStr = dateTimeStr:sub(1, i - 1)
    local timeStr = dateTimeStr:sub(i + 1)
    local year, month, day = string.match(dateStr, "(%d+)-(%d+)-(%d+)")
    local hour, min, sec = string.match(timeStr, "(%d+):(%d+):?(%d+)")

    if not year then
        return nil
    end

    return os.time {
        year=year;
        month=month;
        day=day;
        hour=hour;
        min=min;
        sec=sec;
    }
end

function P.indent(str, numSpaces)
    if not numSpaces then
        numSpaces = 1
    end
    local result = {}
    for line in str:gmatch "[^\n]+" do
        for i = 1, numSpaces do
            table.insert(result, " ")
        end
        table.insert(result, line)
        table.insert(result, "\n")
    end
    return table.concat(result, "")
end

function P.lines(str)
    return P.split(str, "\n")
end

function P.joinLines(str)
    local lines = {}
    for line in P.lines(str) do
        line = P.trim(line)
        table.insert(lines, line)
    end
    return table.concat(lines, " ")
end

function P.detab(str, fence)
    local lines = {}
    for line in P.lines(str) do
        table.insert(lines, line)
    end

    for i, line in ipairs(lines) do
        line = P.trimLeft(line)
        if fence and line:sub(1, #fence) == fence then
            line = line:sub(#fence + 1)
        end
        lines[i] = line
    end

    return P.trimRight(table.concat(lines, "\n"))
end

function P.curry2(fn)
    return function(x)
        return function(y)
            return fn(x, y)
        end
    end
end

function P.curry3(fn)
    return P.curry2(function(x, y)
        return function(z)
            return fn(x, y, z)
        end
    end)
end

function P.curry4(fn)
    return P.curry3(function(x, y, z)
        return function(w)
            return fn(x, y, z, w)
        end
    end)
end

function P.partial(fn, ...)
    local args = {...}
    return function(...)
        for _, x in ipairs {...} do
            table.insert(args, x)
        end
        return fn(unpack(args))
    end
end

function P.find(t, item)
    local result = nil
    for i, x in ipairs(t) do
        if x == item then
            return x, i
        end
    end
    return nil
end

function P.findBy(t, pred)
    local result = nil
    for i, x in ipairs(t) do
        if pred(x, i) then
            return x, i
        end
    end
    return nil
end

return P
