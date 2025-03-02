--- Events System
-- This module provides an event system for registering and handling events
-- throughout the Runes MUD client.
-- @module events

-- Initialize the events namespace
runes.events = runes.events or {}

-- Private state
local handlers = {}

---
-- Registers a handler function for a specific event.
-- @param eventName string The name of the event to listen for
-- @param handler function The function to call when the event occurs
-- @usage runes.events.add("connect", function(data)
--     runes.output("Connected to server: " .. data.host)
-- end)
function runes.events.add(eventName, handler)
    if not handlers[eventName] then
        handlers[eventName] = {}
    end
    table.insert(handlers[eventName], handler)
end

---
-- Emits an event, triggering all registered handlers.
-- @param eventName string The name of the event to emit
-- @param eventData any Data to pass to the event handlers
-- @usage runes.events.emit("custom_event", { value = 123 })
function runes.events.emit(eventName, eventData)
    if not handlers[eventName] then
        return
    end
    
    for _, handler in ipairs(handlers[eventName]) do
        local status, err = pcall(function()
            handler(eventData)
        end)
        if not status then
            runes.log(string.format("Error in event handler: %s", err))
        end
    end
end

---
-- Removes all handlers for a specific event.
-- @param eventName string The name of the event to clear handlers for
-- @usage runes.events.clear("custom_event")
function runes.events.clear(eventName)
    handlers[eventName] = {}
end

---
-- Removes a specific handler function from an event.
-- @param eventName string The name of the event
-- @param handler function The handler function to remove
-- @return boolean True if the handler was found and removed, false otherwise
-- @usage runes.events.remove("connect", myConnectHandler)
function runes.events.remove(eventName, handler)
    if not handlers[eventName] then
        return false
    end
    
    for i, h in ipairs(handlers[eventName]) do
        if h == handler then
            table.remove(handlers[eventName], i)
            return true
        end
    end
    
    return false
end

---
-- Lists all events that have registered handlers.
-- @return table A list of event names
-- @usage local events = runes.events.list()
-- for _, name in ipairs(events) do
--     runes.output("Event: " .. name)
-- end
function runes.events.list()
    local result = {}
    for eventName, _ in pairs(handlers) do
        table.insert(result, eventName)
    end
    return result
end

-- Register system event handlers
runes.events.add("connect", function(data)
    -- Handle connect event
end)

runes.events.add("disconnect", function(data)
    -- Handle disconnect event
end)

runes.events.add("reset", function(data)
    -- Handle reset event
end)

-- For backward compatibility (optional)
if _G.events == nil then
    _G.events = runes.events
end

return runes.events