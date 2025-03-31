-- This module provides the core functionality for the Runes MUD client,
-- including connection management, input/output handling, and script loading.
-- It serves as the main entry point for interacting with the Runes API.
-- @module runes

-- Initialize the runes namespace
runes = runes or {}

--
-- Core system functions
--

---
-- Prints a debug message to the console.
-- @param text The text to print as a debug message
-- @usage runes.debug("Debug information")
function runes.debug(text)
    return runes._debug(text)
end

---
-- Gets the current version of the Runes client.
-- @return string The version string
-- @usage local version = runes.version()
function runes.version()
    return runes._version()
end

---
-- Gets the current timestamp in milliseconds.
-- @return number Current time in milliseconds
-- @usage local now = runes.getTime()
function runes.getTime()
    -- Use os.time() (seconds) and convert to milliseconds since there's no binding
    return os.time() * 1000
end

---
-- Logs a message to the console.
-- @param text The text to log
-- @usage runes.log("Log information")
function runes.log(text)
    return runes._log(text)
end

---
-- Outputs text to the client.
-- @param text The text to output
-- @usage runes.output("Hello, world!")
function runes.output(text)
    return runes._output(text)
end

---
-- Sends text directly to the server, bypassing alias processing.
-- @param text The text to send
-- @usage runes.send_raw("look")
function runes.send_raw(text)
    -- Bypass the input system and send directly to the server
    return runes._send_raw(text)
end

---
-- Sends text to the server, with alias processing.
-- @param text The text to send
-- @usage runes.send("look")
function runes.send(text)
    -- Use the input system to process the command through aliases
    runes.input.send(text)
end

---
-- Connects to a MUD server.
-- @param host The hostname or IP address
-- @param port The port number
-- @usage runes.connect("example.com", 4000)
function runes.connect(host, port)
    return runes._connect(host, port)
end

---
-- Disconnects from the current server.
-- @usage runes.disconnect()
function runes.disconnect()
    return runes._disconnect()
end

---
-- Adds a processor for input lines.
-- @param processor Function that processes input lines
-- @usage runes.add_input_listener(function(line)
--     return line
-- end)
function runes.add_input_listener(processor)
    return runes._add_input_processor(processor)
end

---
-- Adds a processor for output lines.
-- @param processor Function that processes output lines
-- @usage runes.add_output_listener(function(line)
--     return line
-- end)
function runes.add_output_listener(processor)
    return runes._add_output_processor(processor)
end

---
-- Adds a handler for tick events.
-- @param handler Function to call on each tick
-- @usage runes.add_tick_handler(function()
--     -- Do something on each tick
-- end)
function runes.add_tick_handler(handler)
    -- Store handlers in a table since there's no binding
    if not runes._tick_handlers then
        runes._tick_handlers = {}
    end
    table.insert(runes._tick_handlers, handler)
end


