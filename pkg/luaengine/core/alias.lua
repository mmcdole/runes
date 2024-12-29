-- core/alias.lua
alias = {}  -- Declare global alias table
local aliases = {}  -- Private state

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

function alias.remove(pattern)
    if aliases[pattern] then
        aliases[pattern] = nil
    end
end
