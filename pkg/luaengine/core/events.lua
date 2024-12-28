-- core/events.lua
events = {}

local handlers = {}

function events.add(eventName, handler)
    if not handlers[eventName] then
        handlers[eventName] = {}
    end
    table.insert(handlers[eventName], handler)
end

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