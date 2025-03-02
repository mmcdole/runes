--- Timer System
-- This module provides functionality for creating and managing timers
-- that execute functions at specified intervals.
-- @module timer

-- Initialize the timer namespace
runes.timer = runes.timer or {}
local timers = {}  -- Private state
local nextId = 1

---
-- Adds a new timer that executes a function at specified intervals.
-- @param interval number Time in milliseconds between executions
-- @param callback function Function to execute when the timer fires
-- @param repeating boolean If true, timer will continue executing until removed
-- @return number Timer ID that can be used to remove or modify the timer
-- @usage local timerId = runes.timer.add(1000, function()
--     runes.output("One second has passed!")
-- end, true)
function runes.timer.add(interval, callback, repeating)
    local id = nextId
    nextId = nextId + 1
    
    timers[id] = {
        interval = interval,
        callback = callback,
        repeating = repeating or false,
        lastRun = runes.getTime(),
        enabled = true
    }
    
    return id
end

---
-- Removes a timer by its ID.
-- @param id number The ID of the timer to remove
-- @return boolean True if the timer was found and removed, false otherwise
-- @usage runes.timer.remove(timerId)
function runes.timer.remove(id)
    if timers[id] then
        timers[id] = nil
        return true
    end
    return false
end

---
-- Enables a timer that was previously disabled.
-- @param id number The ID of the timer to enable
-- @return boolean True if the timer was found and enabled, false otherwise
-- @usage runes.timer.enable(timerId)
function runes.timer.enable(id)
    if timers[id] then
        timers[id].enabled = true
        return true
    end
    return false
end

---
-- Disables a timer without removing it.
-- @param id number The ID of the timer to disable
-- @return boolean True if the timer was found and disabled, false otherwise
-- @usage runes.timer.disable(timerId)
function runes.timer.disable(id)
    if timers[id] then
        timers[id].enabled = false
        return true
    end
    return false
end

---
-- Lists all active timers.
-- @return table A list of timer IDs and their properties
-- @usage local timerList = runes.timer.list()
function runes.timer.list()
    local result = {}
    for id, t in pairs(timers) do
        result[id] = {
            interval = t.interval,
            repeating = t.repeating,
            enabled = t.enabled
        }
    end
    return result
end

-- Process timers (internal function)
local function process_timers()
    local currentTime = runes.getTime()
    
    for id, t in pairs(timers) do
        if t.enabled and (currentTime - t.lastRun) >= t.interval then
            t.lastRun = currentTime
            
            local status, err = pcall(t.callback)
            if not status then
                runes.log(string.format("Error in timer callback: %s", err))
            end
            
            if not t.repeating then
                timers[id] = nil
            end
        end
    end
end

-- Register timer processor with the main loop
runes.add_tick_handler(process_timers)

-- For backward compatibility (optional)
if _G.timer == nil then
    _G.timer = runes.timer
end