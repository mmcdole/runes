-- core/trigger.lua

trigger = {}  -- Declare global alias table

-- Trigger system
local triggers = {}

-- Add a new trigger
function trigger.add(name, pattern, callback)
    if type(callback) ~= "function" then
        runes.debug(string.format("Error: Invalid callback type for trigger %q", name))
        return
    end

    table.insert(triggers, {
        name = name,
        pattern = pattern,
        callback = callback
    })
    runes.debug(string.format("Trigger added: %q", name))
end

-- Remove a trigger by name
function trigger.remove(name)
    for i, trigger in ipairs(triggers) do
        if trigger.name == name then
            table.remove(triggers, i)
            runes.debug(string.format("Trigger removed: %q", name))
            return
        end
    end
end

-- Process output against triggers
local function process(output)
    for _, trigger in ipairs(triggers) do
        local matches = {string.match(output, trigger.pattern)}
        if matches[1] then
            runes.debug(string.format("Trigger %q matched: %s", trigger.name, output))
            trigger.callback(matches)
        end
    end
end

-- Subscribe to the 'output' event
events.add("output", process)
