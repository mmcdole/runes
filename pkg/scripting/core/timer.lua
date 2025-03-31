--- Timer System
-- This module provides functionality for creating and managing timers
-- that execute functions at specified intervals.
-- @module timer

-- Initialize the timer namespace
timer = timer or {}

-- Wrapper functions for the native timer implementation

---
-- Adds a new timer that executes a function at specified intervals.
-- @param interval number Time in seconds between executions
-- @param count number Number of times to execute (0 for infinite)
-- @param callback function Function to execute when the timer fires
-- @return number Timer ID that can be used to remove the timer
-- @usage local timerId = timer.add(1, 0, function()
--     runes.output("One second has passed!")
-- end)
function timer.add(interval, count, callback)
    -- Call the native timer.add function
    return timer.add(interval, count, callback)
end

---
-- Removes a timer by its ID.
-- @param id number The ID of the timer to remove
-- @usage timer.remove(timerId)
function timer.remove(id)
    -- Call the native timer.remove function
    timer.remove(id)
end

---
-- Clears all timers.
-- @usage timer.clear()
function timer.clear()
    -- Call the native timer.clear function
    timer.clear()
end

---
-- Gets all timer IDs.
-- @return table A list of timer IDs
-- @usage local ids = timer.get_ids()
function timer.get_ids()
    -- Call the native timer.get_ids function
    return timer.get_ids()
end

---
-- Registers a function to be called on every tick.
-- @param callback function Function to call on every tick
-- @usage timer.on_tick(function(millis)
--     runes.output("Tick: " .. millis .. "ms")
-- end)
function timer.on_tick(callback)
    -- Call the native timer.on_tick function
    timer.on_tick(callback)
end

-- Backward compatibility functions

---
-- Adds a timer with millisecond precision (for backward compatibility).
-- @param interval number Time in milliseconds between executions
-- @param callback function Function to execute when the timer fires
-- @param repeating boolean If true, timer will continue executing until removed
-- @return number Timer ID that can be used to remove or modify the timer
-- @usage local timerId = timer.add_ms(1000, function()
--     runes.output("One second has passed!")
-- end, true)
function timer.add_ms(interval, callback, repeating)
    local count = 0
    if not repeating then
        count = 1
    end
    return timer.add(interval / 1000, count, callback)
end

