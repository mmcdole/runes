-- core/init.lua

-- Handle output
runes.add_output_listener(function(line)
    -- Default input handler - just return line
    return line
end)

-- Handle input
runes.add_input_listener(function(line)
    -- Default input handler - just return line
    return line
end)

-- Register connect event handler using the events system
runes.events.add("connected", function(data)
    local message = "Connected to " .. data.host .. ":" .. data.port
    runes.output(message)
end)

-- Register disconnect event handler using the events system
runes.events.add("disconnected", function()
    runes.output("Disconnected from server")
end)

-- Register script reset event handler
runes.events.add("reset", function()
    -- Re-initialize any state needed after reset
    runes.output("Script reset")
end)

-- Initialize message
runes.output(C_GREEN .. "Welcome to Runes, the MUD client!" .. C_RESET)
runes.output("Type /help for a list of available commands")