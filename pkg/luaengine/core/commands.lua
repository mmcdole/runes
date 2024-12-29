-- core/commands.lua

-- Command syntax definitions
local commands = {
    connect = {
        syntax = "/connect <host> <port>",
        description = "Connect to a MUD server",
        help = "Examples:\n  /connect example.com 4000\n  /connect localhost 4000"
    },
    disconnect = {
        syntax = "/disconnect",
        description = "Disconnect from the current server"
    },
    quit = {
        syntax = "/quit",
        description = "Quit the client"
    },
    buffer = {
        syntax = "/buffer <list|switch <name>>",
        description = "Buffer management commands",
        help = "Examples:\n  /buffer list\n  /buffer switch system"
    },
    help = {
        syntax = "/help [command]",
        description = "Show help for all commands or a specific command"
    }
}

-- Helper function to show command help
local function show_command_help(cmd_name)
    local cmd = commands[cmd_name]
    if not cmd then return end
    
    runes.output(string.format([[
Command: %s
Syntax: %s
%s
%s]], 
        cmd_name,
        cmd.syntax,
        cmd.description,
        cmd.help or ""
    ))
end

-- Helper function to show just the syntax
local function show_syntax(cmd_name)
    local cmd = commands[cmd_name]
    if not cmd then return end
    runes.output("Syntax: " .. cmd.syntax)
end

-- Connection management
alias.add("^/connect%s*(.*)$", function(matches, line)
    local args = matches[1]
    local host, port = string.match(args, "^(%S+)%s+(%d+)$")
    
    if not host or not port then
        show_syntax("connect")
        return
    end
    
    port = tonumber(port)
    if port < 1 or port > 65535 then
        runes.output("Error: Port must be between 1 and 65535")
        show_syntax("connect")
        return
    end
    
    runes.connect(host, port)
end)

alias.add("^/disconnect$", function(matches, line)
    runes.disconnect()
end)

-- Buffer management
alias.add("^/buffer%s*(.*)$", function(matches, line)
    local args = matches[1]
    
    if args == "list" then
        local buffers = mud.list_buffers()
        runes.output("=== Buffers ===")
        for _, buf in ipairs(buffers) do
            runes.output("- " .. buf)
        end
        return
    end
    
    local cmd, name = string.match(args, "^(%S+)%s+(%S+)$")
    if cmd == "switch" and name then
        mud.switch_buffer(name)
        return
    end
    
    show_syntax("buffer")
end)

-- Help command
alias.add("^/help%s*(.*)$", function(matches, line)
    local cmd_name = matches[1]
    if cmd_name ~= "" then
        if commands[cmd_name] then
            show_command_help(cmd_name)
        else
            runes.output("Unknown command: " .. cmd_name)
        end
        return
    end
    
    runes.output([[
Available commands:
  /help [command] - Show help for all commands or a specific command
  /connect        - Connect to a MUD server: /connect <host> <port>
  /disconnect     - Disconnect from server
  /buffer list    - List all buffers
  /buffer switch  - Switch to a different buffer
  /quit           - Quit the client

Type /help <command> for detailed help on a specific command.
]])
end)

-- Quit command
alias.add("^/quit$", function(matches, line)
    runes.quit()
end)