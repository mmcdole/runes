-- core/command.lua
command = {}  -- Declare global command table

local commandQueue = {}  -- Private state
local commandSeparator = commandSeparator or ";"

-- Redefine mud.send to use the command system
function runes.send(cmd)  -- Rename parameter to avoid shadowing
    if cmd then
        command.enqueue(cmd)  -- Use global command table
    end
end

-- Split a command string based on the separator
local function splitCommand(command)
    local commands = {}
    for cmd in (command..commandSeparator):gmatch("(.-)"..commandSeparator) do
        cmd = cmd:match("^%s*(.-)%s*$")
        table.insert(commands, cmd)
    end
    return commands
end

-- Process a single command
local function processCommand(command)
    if command ~= "" then
        local aliasFunc = alias.resolve(command)
        if aliasFunc then
            runes.debug(string.format("Found alias for command: %q", command))
            aliasFunc()
            return {}
        else
            runes.debug(string.format("No alias found for command: %q", command))
        end
    end
    
    runes.sendRaw(command)
    return {}
end

-- Process the command queue
local function processCommandQueue()
    while #commandQueue > 0 do
        local command = table.remove(commandQueue, 1)
        local commands = splitCommand(command)
        for _, cmd in ipairs(commands) do
            local newCommands = processCommand(cmd)
            for i = #newCommands, 1, -1 do
                table.insert(commandQueue, 1, newCommands[i])
            end
        end
    end
end

function command.enqueue(command)
    table.insert(commandQueue, command)
    processCommandQueue()
end

-- Subscribe to the 'input' event
events.add("input", function(input)
    runes.debug("[INPUT] " .. input)
    command.enqueue(input)
end)
