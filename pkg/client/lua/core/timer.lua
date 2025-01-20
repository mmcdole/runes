-- core/timer.lua

timer = {}  -- Declare global timer table
local timers = {}  -- Private state
local nextId = 1

-- Add a new timer
-- interval: milliseconds between executions
-- callback: function to execute
-- repeating: if true, timer will continue executing until removed
function timer.add(interval, callback, repeating)
    if type(callback) ~= "function" then
        return nil
    end
    
    local id = nextId
    nextId = nextId + 1
    
    timers[id] = {
        interval = interval,
        callback = callback,
        repeating = repeating or false,
        next_run = os.time() * 1000 + interval,
        enabled = true
    }
    
    return id
end

-- Convenience function for one-time timers
function timer.once(interval, callback)
    return timer.add(interval, callback, false)
end

-- Remove a timer by id
function timer.remove(id)
    timers[id] = nil
end

-- Enable/disable timers
function timer.enable(id)
    if timers[id] then
        timers[id].enabled = true
        timers[id].next_run = os.time() * 1000 + timers[id].interval
    end
end

function timer.disable(id)
    if timers[id] then
        timers[id].enabled = false
    end
end

-- List all timers
function timer.list()
    local result = {}
    for id, t in pairs(timers) do
        table.insert(result, {
            id = id,
            interval = t.interval,
            repeating = t.repeating,
            enabled = t.enabled,
            next_run = t.next_run
        })
    end
    return result
end

-- Process timers
local function process()
    local now = os.time() * 1000  -- Current time in ms
    
    for id, timer in pairs(timers) do
        if timer.enabled and now >= timer.next_run then
            local status, err = pcall(timer.callback)
            if not status then
                runes.debug(string.format("Timer %d error: %s", id, err))
            end
            
            if timer.repeating then
                timer.next_run = now + timer.interval
            else
                timers[id] = nil
            end
        end
    end
end

-- Subscribe to a tick event (this should be emitted by the Go side regularly)
events.add("tick", process) 