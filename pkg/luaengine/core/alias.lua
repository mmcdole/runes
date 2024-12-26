-- core/alias.lua
alias = {}  -- Declare global alias table
local aliases = {}  -- Private state

function alias.add(pattern, callback)
    runes.debug(string.format("[ALIAS] Adding alias for pattern: %q", pattern))
    
    if type(callback) == "string" then
        local command = callback
        runes.debug(string.format("[ALIAS] Creating function for string command: %q", command))
        callback = function()
            runes.debug(string.format("[ALIAS] Executing string alias: %q", command))
            runes.send(command)
        end
    elseif type(callback) ~= "function" then
        runes.debug(string.format("[ALIAS] Invalid callback type for pattern: %q", pattern))
        return
    end

    aliases[pattern] = callback
    runes.debug(string.format("[ALIAS] Successfully added alias for: %q", pattern))
end

function alias.resolve(input)
    runes.debug(string.format("[ALIAS] Resolving alias for: %q", input))
    local result = aliases[input]
    if result then
        runes.debug(string.format("[ALIAS] Found alias for: %q", input))
    else
        runes.debug(string.format("[ALIAS] No alias found for: %q", input))
    end
    return result
end

function alias.remove(pattern)
    if aliases[pattern] then
        aliases[pattern] = nil
        runes.debug(string.format("[ALIAS] Removed alias: %q", pattern))
    end
end
