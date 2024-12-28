-- core/command.lua
local commandQueue = {}
local commandSeparator = commandSeparator or ";"

local function splitCommand(command)
    local commands = {}
    for cmd in (command..commandSeparator):gmatch("(.-)"..commandSeparator) do
        cmd = cmd:match("^%s*(.-)%s*$")
        table.insert(commands, cmd)
    end
    return commands
end

local function processCommand(command)
    if command ~= "" then
        local aliasFunc = alias.resolve(command)
        if aliasFunc then
            aliasFunc()
            return {}
        end
    end
    
    runes.sendRaw(command)
    return {}
end

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

-- Public API
function runes.send(commandStr)
    if commandStr then
        table.insert(commandQueue, commandStr)
        processCommandQueue()
    end
end

-- Subscribe to input events directly
events.add("input", function(input)
    table.insert(commandQueue, input)
    processCommandQueue()
end)
