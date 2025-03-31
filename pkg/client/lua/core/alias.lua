--- Alias System
-- This module provides functionality for creating and managing aliases
-- that transform user input before sending it to the server.
-- @module alias

-- Initialize the alias namespace
alias = alias or {}

-- Private state
local aliases = {}  -- Private state

---
-- Adds a new alias with a regex pattern.
-- @param pattern string The Lua pattern to match against input
-- @param callback function|string Function or string to execute when matched.
--                 If string: sends the string as a command
--                 If function: called with (matches, input) arguments
-- @usage alias.add("^hp$", "look at health")
-- @usage alias.add("^cast (.+) at (.+)$", function(matches, input)
--     runes.send("cast " .. matches[1] .. " " .. matches[2])
-- end)
function alias.add(pattern, callback)    
    if type(callback) == "string" then
        local command = callback
        callback = function(matches, input)
            -- Split the command by semicolons and send each part
            for cmd in (command..";"):gmatch("(.-);") do
                cmd = cmd:match("^%s*(.-)%s*$") -- Trim whitespace
                if cmd ~= "" then
                    runes.send(cmd)
                end
            end
        end
    elseif type(callback) ~= "function" then
        return
    end
    aliases[pattern] = callback
end

---
-- Creates an alias that matches text exactly.
-- @param text string The text to match exactly (will be escaped for regex)
-- @param callback function|string Function or string to execute when matched
-- @usage alias.exact("hp", "look at health")
function alias.exact(text, callback)
    -- Escape any special pattern characters in the text
    local escaped = text:gsub("[%(%)%.%%%+%-%*%?%[%]%^%$]", "%%%1")
    return runes.alias.add("^" .. escaped .. "$", callback)
end

---
-- Creates an alias that matches any of the provided texts exactly.
-- @param texts table Array of text strings to match (each will be escaped)
-- @param callback function|string Function or string to execute when matched
-- @usage alias.any({"hp", "health", "hitpoints"}, "look at health")
function alias.any(texts, callback)
    -- Escape special characters in each option and join with |
    local patterns = {}
    for _, text in ipairs(texts) do
        local escaped = text:gsub("[%(%)%.%%%+%-%*%?%[%]%^%$]", "%%%1")
        table.insert(patterns, escaped)
    end
    return runes.alias.add("^(" .. table.concat(patterns, "|") .. ")$", callback)
end

---
-- Creates an alias that matches text starting with the given prefix.
-- @param prefix string The prefix to match at the start (will be escaped)
-- @param callback function|string Function or string to execute when matched
-- @usage alias.starts("cast ", function(matches, input)
--     runes.send("cast_spell " .. matches[1])
-- end)
function alias.starts(prefix, callback)
    -- Escape special characters in prefix and capture the rest
    local escaped = prefix:gsub("[%(%)%.%%%+%-%*%?%[%]%^%$]", "%%%1")
    return alias.add("^" .. escaped .. "(.+)$", callback)
end

---
-- Attempts to match input against registered aliases.
-- @param input string The input string to check against aliases
-- @return function|nil Returns a wrapper function if matched, nil otherwise
-- @usage local handler = alias.resolve("hp")
-- if handler then handler() end
function alias.resolve(input)
    -- Try each alias pattern
    for pattern, callback in pairs(aliases) do
        -- Collect all matches from the pattern
        local matches = {input:match(pattern)}
        if #matches > 0 then
            -- Return a wrapper that calls the callback with matches and original line
            return function()
                callback(matches, input)
            end
        end
    end
    return nil
end

---
-- Removes an alias by its pattern.
-- @param pattern string The exact pattern string that was used to create the alias
-- @usage alias.remove("^hp$")
function alias.remove(pattern)
    if aliases[pattern] then
        aliases[pattern] = nil
    end
end

---
-- Returns a list of all defined aliases.
-- @return table A list of alias patterns
-- @usage local all_aliases = alias.list()
-- for _, pattern in ipairs(all_aliases) do
--     runes.output("Alias pattern: " .. pattern)
-- end
function alias.list()
    local result = {}
    for pattern, _ in pairs(aliases) do
        table.insert(result, pattern)
    end
    return result
end