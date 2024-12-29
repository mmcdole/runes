-- core/init.lua

-- Set up event handlers
events.add("connect", function(data)
    runes.output("Connected to server")
end)

events.add("disconnect", function(data)
    runes.output("Disconnected from server")
end)

events.add("error", function(data)
    runes.output("Error: " .. data, "system")
end)

-- Handle output
events.add("output", function(data)
    -- For now, just display raw output
    -- Later we can add triggers and other processing here
    runes.output(data)
    return false
end)

-- Handle input
events.add("raw_input", function(data)
    -- Process input through alias system
    local aliasFunc = alias.resolve(data)
    if aliasFunc then
        aliasFunc()
        return true
    end
    -- If no alias found, send raw command
    runes.sendRaw(data)
    return false
end)

-- Initialize message
runes.output("Runes MUD Client initialized", "system") 