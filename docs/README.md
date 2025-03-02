# Runes MUD Client API Documentation

This directory contains the API documentation for the Runes MUD Client.

## Generating Documentation

To generate the documentation, run the following command from the project root:

```bash
./generate_docs.sh
```

This will generate HTML documentation in this directory. Open `index.html` to view the documentation.

## Documentation Format

The documentation is generated from LuaDoc comments in the source code. Each function is documented with:

- Description
- Parameters
- Return values
- Usage examples

## Example Usage

Here's how to use the Runes API in your scripts:

```lua
-- Method 1: Use the full namespace
runes.trigger.add("hp_watch", "HP: (%d+)/(%d+)", function(matches, line)
    local current = tonumber(matches[1])
    local max = tonumber(matches[2])
    if current < max * 0.3 then
        runes.output("LOW HEALTH WARNING!")
    end
end)

-- Method 2: Using local variables
local trigger = runes.trigger
local alias = runes.alias

trigger.add("mana_watch", "MP: (%d+)/(%d+)", function(matches, line)
    -- Your code here
end)

-- Method 3: Using the import helper
runes.import("events", "trigger", "alias")
-- Now events, trigger, and alias are available globally
```

For more examples, see the `examples` directory.
