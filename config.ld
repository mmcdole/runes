-- LDoc configuration file for Runes MUD Client API
project = 'Runes MUD Client'
title = 'Runes MUD Client API Documentation'
description = 'Documentation for the Runes MUD Client Lua API'
format = 'markdown'
dir = 'docs'

-- Define the sections we want
kind_names = {
    module = "API Reference"
}

-- Specify API modules
file = {
    'pkg/client/lua/core/runes.lua',
    'pkg/client/lua/core/trigger.lua',
    'pkg/client/lua/core/alias.lua',
    'pkg/client/lua/core/events.lua',
    'pkg/client/lua/core/timer.lua',
    'pkg/client/lua/core/input.lua'
}

-- Custom styles
style = "!pale"  -- Use a predefined style (options: default, pale, elegant)

tparam_alias('string', 'string')
tparam_alias('func', 'function')
tparam_alias('tab', 'table')
tparam_alias('bool', 'boolean')
tparam_alias('num', 'number')

-- Custom see handler for cross-references
custom_see_handler('^([^:]+):(.+)$', function(name, link)
    return name, link
end)
