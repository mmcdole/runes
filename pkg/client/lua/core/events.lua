--- Events System
-- This module provides an event system for registering and handling events
-- throughout the Runes MUD client.
-- @module events

-- Initialize the events namespace
events = events or {}

-- Private state
local handlers = {}

---
-- Registers a handler function for a specific event.
-- @param eventName string The name of the event to listen for
-- @param handler function The function to call when the event occurs
-- @usage events.add("connect", function(data)
--     runes.output("Connected to server: " .. data.host)
-- end)
function events.add(eventName, handler)
    if not handlers[eventName] then
        handlers[eventName] = {}
    end
    table.insert(handlers[eventName], handler)
end

---
-- Emits an event, triggering all registered handlers.
-- @param eventName string The name of the event to emit
-- @param eventData any Data to pass to the event handlers
-- @usage events.emit("custom_event", { value = 123 })
function events.emit(eventName, eventData)
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
-- @usage events.clear("custom_event")
function events.clear(eventName)
    handlers[eventName] = {}
end

---
-- Removes a specific handler function from an event.
-- @param eventName string The name of the event
-- @param handler function The handler function to remove
-- @return boolean True if the handler was found and removed, false otherwise
-- @usage events.remove("connect", myConnectHandler)
function events.remove(eventName, handler)
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
-- @usage local events = events.list()
-- for _, name in ipairs(events) do
--     runes.output("Event: " .. name)
-- end
function events.list()
    local result = {}
    for eventName, _ in pairs(handlers) do
        table.insert(result, eventName)
    end
    return result
end