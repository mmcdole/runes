-- core/init.lua

-- Handle output
runes.add_output_listener(function(line)
    -- Default output handler - just return line if not gagged
    if line:gag() then
        return nil  -- Don't display gagged lines
    end
    return line
end)

-- Handle input
runes.add_input_listener(function(line)
    -- Default input handler - just return line
    return line
end)

-- Handle connect/disconnect
runes.on_connect(function(host, port)
    local line = line.new("Connected to " .. host .. ":" .. port)
    line:matched(true)  -- Mark as system message
    return line
end)

runes.on_disconnect(function()
    local line = line.new("Disconnected from server")
    line:matched(true)  -- Mark as system message
    return line
end)

-- Handle script resets
runes.on_reset(function()
    -- Re-initialize any state needed after reset
    print("Script reset")
end)

-- Initialize message
local welcome = line.new(C_GREEN .. "Welcome to Runes, the MUD client!" .. C_RESET)
welcome:matched(true)  -- Mark as system message
runes.output(welcome)

local help = line.new("Type /help for a list of available commands")
help:matched(true)  -- Mark as system message
runes.output(help)