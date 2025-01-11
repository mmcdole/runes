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
    },
    load = {
        syntax = "/load <path>",
        description = "Load a script file",
        help = "Examples:\n  /load myscript.lua\n  /load /absolute/path/script.lua"
    },
    aliases = {
        syntax = "/aliases",
        description = "List all defined aliases"
    },
    triggers = {
        syntax = "/triggers",
        description = "List all defined triggers"
    }
}

-- Helper function to show command help
local function show_command_help(cmd_name)
    local cmd = commands[cmd_name]
    if not cmd then return end
    
    runes.output(C_GREEN .. string.format([[
Command: %s
Syntax: %s
%s
%s]], 
        cmd_name,
        cmd.syntax,
        cmd.description,
        cmd.help or ""
    ) .. C_RESET)
end

-- Helper function for state labels (enabled/disabled)
local function state_label(state, label)
    local len = label:len() + 1
    if not state then
        label = "-" .. label
        return C_RED .. string.format("%" .. len .. "s", label) .. C_RESET
    else
        label = "+" .. label
        return C_GREEN .. string.format("%" .. len .. "s", label) .. C_RESET
    end
end

-- Helper function to show just the syntax
local function show_syntax(cmd_name)
    local cmd = commands[cmd_name]
    if not cmd then return end
    runes.output(C_GREEN .. "Syntax: " .. cmd.syntax .. C_RESET)
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
        runes.output(C_RED .. "Error: Port must be between 1 and 65535" .. C_RESET)
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
        runes.output(C_GREEN .. "=== Buffers ===" .. C_RESET)
        for _, buf in ipairs(buffers) do
            runes.output(C_GREEN .. "- " .. buf .. C_RESET)
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
            runes.output(C_GREEN .. "Unknown command: " .. cmd_name .. C_RESET)
        end
        return
    end
    
    runes.output(C_GREEN .. [[
Available commands:
  /help [command] - Show help for all commands or a specific command
  /connect        - Connect to a MUD server: /connect <host> <port>
  /disconnect     - Disconnect from server
  /buffer list    - List all buffers
  /buffer switch  - Switch to a different buffer
  /load           - Load a script file: /load <path>
  /aliases        - List all defined aliases
  /triggers       - List all defined triggers
  /quit           - Quit the client

Type /help <command> for detailed help on a specific command.
]] .. C_RESET)
end)

-- Load command
alias.add("^/load%s*(.*)$", function(matches, line)
    local path = matches[1]
    if not path then
        show_syntax("load")
        return
    end
    
    -- Load the script using Go binding
    local ok, err = runes.load_script(path)
    if not ok then
        runes.output(C_RED .. "Error: " .. err .. C_RESET)
        return
    end
    
    runes.output(C_GREEN .. "Successfully loaded script: " .. path .. C_RESET)
end)

-- List aliases command
alias.add("^/aliases$", function(matches, line)
    runes.output(C_GREEN .. "=== Aliases ===" .. C_RESET)
    local alist = alias.list()
    table.sort(alist)  -- Sort alphabetically
    for _, pattern in ipairs(alist) do
        runes.output(string.format("%s%s%s %s", 
            C_YELLOW,
            pattern,
            C_RESET,
            state_label(true, "enabled")
        ))
    end
end)

-- List triggers command
alias.add("^/triggers$", function(matches, line)
    runes.output(C_GREEN .. "=== Triggers ===" .. C_RESET)
    local tlist = trigger.list()
    -- Sort by name
    table.sort(tlist, function(a, b) return a.name < b.name end)
    for _, t in ipairs(tlist) do
        runes.output(string.format("%-20s : %s%s%s %s",
            t.name,
            C_YELLOW,
            t.pattern,
            C_RESET,
            state_label(t.enabled, "enabled")
        ))
    end
end)

-- Quit command
alias.add("^/quit$", function(matches, line)
    runes.quit()
end)
