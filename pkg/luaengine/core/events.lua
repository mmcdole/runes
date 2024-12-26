-- core/events.lua
events = {}  -- Declare global events table first!

local handlers = {}

-- Register an event handler
function events.add(eventName, handler)
    if not handlers[eventName] then
        handlers[eventName] = {}
    end
    table.insert(handlers[eventName], handler)
    runes.debug("Event handler added: " .. eventName)
end

-- Emit an event
function events.emit(eventName, eventData)
    if handlers[eventName] then
        for _, handler in ipairs(handlers[eventName]) do
            handler(eventData)
        end
    end
end
