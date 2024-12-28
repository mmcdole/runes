-- core/alias.lua
alias = {}  -- Declare global alias table
local aliases = {}  -- Private state

function alias.add(pattern, callback)    
    if type(callback) == "string" then
        local command = callback
        callback = function()
            runes.send(command)
        end
    elseif type(callback) ~= "function" then
        return
    end

    aliases[pattern] = callback
end

function alias.resolve(input)
    return aliases[input]
end

function alias.remove(pattern)
    if aliases[pattern] then
        aliases[pattern] = nil
    end
end
