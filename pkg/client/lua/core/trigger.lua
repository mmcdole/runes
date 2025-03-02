--- Trigger System
-- This module provides functionality for creating and managing triggers
-- that respond to patterns in the MUD output.
-- @module trigger

-- Initialize the trigger namespace
runes.trigger = runes.trigger or {}

-- Private state
local triggers = {}  -- Private state
local next_id = 1    -- For unique trigger IDs

---
-- Adds a new trigger that responds to a pattern in the MUD output.
-- @param name string Name of the trigger for reference
-- @param pattern string Lua pattern to match against incoming text
-- @param options table|function Optional table of settings or the callback function
-- @param callback function Function to call when pattern matches (if options provided)
-- @return number Unique ID of the created trigger
-- @usage runes.trigger.add("hp_watch", "HP: (%d+)/(%d+)", function(matches, line)
--    if tonumber(matches[1]) < 100 then
--        runes.output("Low health warning!")
--    end
-- end)
function runes.trigger.add(name, pattern, options, callback)
    -- Handle case where options is omitted and callback is passed as third argument
    if type(options) == "function" and callback == nil then
        callback = options
        options = {}
    end
    
    if type(callback) ~= "function" then
        return nil
    end
    
    local id = next_id
    next_id = next_id + 1
    
    local new_trigger = {
        id = id,
        name = name,
        pattern = pattern,
        callback = callback,
        enabled = options.enabled ~= false, -- Default to true unless explicitly set to false
        gag = options.gag or false,         -- Whether to gag the line when matched
        raw = options.raw or false          -- Whether to match against raw line with ANSI codes
    }
    
    triggers[id] = new_trigger
    return id
end

---
-- Removes a trigger by its ID or name.
-- @param id_or_name number|string The ID or name of the trigger to remove
-- @return boolean True if the trigger was found and removed, false otherwise
-- @usage runes.trigger.remove(5)  -- Remove by ID
-- @usage runes.trigger.remove("hp_watch")  -- Remove by name
function runes.trigger.remove(id_or_name)
    if type(id_or_name) == "number" then
        -- Remove by ID
        if triggers[id_or_name] then
            triggers[id_or_name] = nil
            return true
        end
    else
        -- Remove by name
        for id, t in pairs(triggers) do
            if t.name == id_or_name then
                triggers[id] = nil
                return true
            end
        end
    end
    return false
end

---
-- Enables a trigger by its ID or name.
-- @param id_or_name number|string The ID or name of the trigger to enable
-- @return boolean True if the trigger was found and enabled, false otherwise
-- @usage runes.trigger.enable(5)
-- @usage runes.trigger.enable("hp_watch")
function runes.trigger.enable(id_or_name)
    local t = runes.trigger.get(id_or_name)
    if t then
        t.enabled = true
        return true
    end
    return false
end

---
-- Disables a trigger by its ID or name.
-- @param id_or_name number|string The ID or name of the trigger to disable
-- @return boolean True if the trigger was found and disabled, false otherwise
-- @usage runes.trigger.disable(5)
-- @usage runes.trigger.disable("hp_watch")
function runes.trigger.disable(id_or_name)
    local t = runes.trigger.get(id_or_name)
    if t then
        t.enabled = false
        return true
    end
    return false
end

---
-- Gets a trigger by its ID or name.
-- @param id_or_name number|string The ID or name of the trigger to retrieve
-- @return table|nil The trigger table if found, nil otherwise
-- @usage local trigger = runes.trigger.get(5)
-- @usage local trigger = runes.trigger.get("hp_watch")
function runes.trigger.get(id_or_name)
    if type(id_or_name) == "number" then
        return triggers[id_or_name]
    else
        for _, t in pairs(triggers) do
            if t.name == id_or_name then
                return t
            end
        end
    end
    return nil
end

---
-- Lists all triggers.
-- @return table Array of trigger information tables
-- @usage local all_triggers = runes.trigger.list()
-- for _, t in ipairs(all_triggers) do
--     runes.output(string.format("Trigger: %s (ID: %d)", t.name, t.id))
-- end
function runes.trigger.list()
    local result = {}
    for id, t in pairs(triggers) do
        table.insert(result, {
            id = id,
            name = t.name,
            pattern = t.pattern,
            enabled = t.enabled,
            gag = t.gag,
            raw = t.raw
        })
    end
    return result
end

---
-- Clears all triggers.
-- @usage runes.trigger.clear()
function runes.trigger.clear()
    triggers = {}
end

-- Check if a trigger matches a line (private function)
local function check_trigger(t, line)
    if not t.enabled then
        return false
    end
    
    -- Get the appropriate content based on raw flag
    local content = t.raw and line:raw() or line:line()
    
    -- Try to match the pattern
    local matches = {string.match(content, t.pattern)}
    if matches[1] then
        runes.debug(string.format("Trigger %q (ID: %d) matched: %s", t.name, t.id, content))
        
        -- Apply gagging if specified
        if t.gag then
            line:gag(true)
        end
        
        -- Mark the line as matched
        line:matched(true)
        
        -- Call the trigger callback with matches and the line
        t.callback(matches, line)
        return true
    end
    return false
end

-- Process output against all triggers (private function)
local function process_output(line)
    for _, t in pairs(triggers) do
        check_trigger(t, line)
    end
    
    -- Return the possibly modified line for chaining
    return line
end

-- Register with the output processor system
runes.add_output_processor(process_output)

-- For backward compatibility (optional)
if _G.trigger == nil then
    _G.trigger = runes.trigger
end
