-- core/alias.lua
alias = {}  -- Declare global alias table
local aliases = {}  -- Private state

--- Adds a new alias with a regex pattern
-- @param pattern The Lua pattern to match against input
-- @param callback Function or string to execute when matched.
--                 If string: sends the string as a command
--                 If function: called with (matches, input) arguments
function alias.add(pattern, callback)    
    if type(callback) == "string" then
        local command = callback
        callback = function(matches, input)
            runes.send(command)
        end
    elseif type(callback) ~= "function" then
        return
    end
    aliases[pattern] = callback
end

--- Creates an alias that matches text exactly
-- @param text The text to match exactly (will be escaped for regex)
-- @param callback Function or string to execute when matched
function alias.exact(text, callback)
    -- Escape any special pattern characters in the text
    local escaped = text:gsub("[%(%)%.%%%+%-%*%?%[%]%^%$]", "%%%1")
    return alias.add("^" .. escaped .. "$", callback)
end

--- Creates an alias that matches any of the provided texts exactly
-- @param texts Array of text strings to match (each will be escaped)
-- @param callback Function or string to execute when matched
function alias.any(texts, callback)
    -- Escape special characters in each option and join with |
    local patterns = {}
    for _, text in ipairs(texts) do
        local escaped = text:gsub("[%(%)%.%%%+%-%*%?%[%]%^%$]", "%%%1")
        table.insert(patterns, escaped)
    end
    return alias.add("^(" .. table.concat(patterns, "|") .. ")$", callback)
end

--- Creates an alias that matches text starting with the given prefix
-- @param prefix The prefix to match at the start (will be escaped)
-- @param callback Function or string to execute when matched.
function alias.starts(prefix, callback)
    -- Escape special characters in prefix and capture the rest
    local escaped = prefix:gsub("[%(%)%.%%%+%-%*%?%[%]%^%$]", "%%%1")
    return alias.add("^" .. escaped .. "(.+)$", callback)
end

--- Attempts to match input against registered aliases
-- @param input The input string to check against aliases
-- @return function|nil Returns a wrapper function if matched, nil otherwise
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

--- Removes an alias by its pattern
-- @param pattern The exact pattern string that was used to create the alias
function alias.remove(pattern)
    if aliases[pattern] then
        aliases[pattern] = nil
    end
end