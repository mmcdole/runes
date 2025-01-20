-- core/trigger.lua

trigger = {}  -- Declare global trigger table
local triggers = {}  -- Private state

-- Add a new trigger
function trigger.add(name, pattern, callback)
    if type(callback) ~= "function" then
        return
    end

    table.insert(triggers, {
        name = name,
        pattern = pattern,
        callback = callback,
        enabled = true
    })
end

-- Remove a trigger by name
function trigger.remove(name)
    for i, trigger in ipairs(triggers) do
        if trigger.name == name then
            table.remove(triggers, i)
            return
        end
    end
end

-- Enable/disable triggers
function trigger.enable(name)
    for _, t in ipairs(triggers) do
        if t.name == name then
            t.enabled = true
            return
        end
    end
end

function trigger.disable(name)
    for _, t in ipairs(triggers) do
        if t.name == name then
            t.enabled = false
            return
        end
    end
end

-- List all triggers
function trigger.list()
    local result = {}
    for _, t in ipairs(triggers) do
        table.insert(result, {
            name = t.name,
            pattern = t.pattern,
            enabled = t.enabled
        })
    end
    return result
end

-- Process output against triggers
local function process(output)
    for _, trigger in ipairs(triggers) do
        if trigger.enabled then
            local matches = {string.match(output, trigger.pattern)}
            if matches[1] then
                runes.debug(string.format("Trigger %q matched: %s", trigger.name, output))
                trigger.callback(matches)
            end
        end
    end
end

-- Subscribe to the 'output' event
events.add("output", process)
