-- core/mud.lua

-- Connection management
alias.add("/connect", function(args)
    local host, port = string.match(args, "(%S+)%s+(%d+)")
    if not host or not port then
        runes.output("Usage: /connect <host> <port>")
        return
    end
    runes.connect(host, tonumber(port))
end)

alias.add("/disconnect", function()
    runes.disconnect()
end)

-- Buffer management
alias.add("/buffer", function(args)
    local cmd, name = string.match(args, "(%S+)%s*(%S*)")
    
    if cmd == "list" then
        local buffers = mud.list_buffers()
        runes.output("=== Buffers ===")
        for _, buf in ipairs(buffers) do
            runes.output("- " .. buf)
        end
    elseif cmd == "switch" and name ~= "" then
        mud.switch_buffer(name)
    else
        runes.output("Usage: /buffer list|switch <name>")
    end
end)

-- Help command
alias.add("/help", function()
    runes.output([[
Available Commands:
  /connect <host> <port>  - Connect to a MUD server
  /disconnect            - Disconnect from current server
  /buffer list          - List available buffers
  /buffer switch <name> - Switch to specified buffer
  /help                 - Show this help message
]])
end)

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

-- Initialize message
runes.output("Runes MUD Client initialized", "system")
