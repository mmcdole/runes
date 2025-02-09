# Runes MUD Client Specification

## Overview
Runes is a modern MUD client built in Go with Lua scripting support. It follows a direct processing model where Lua is the central processor for all I/O operations, inspired by Blightmud's architecture.

## Package Structure

```
pkg/
├── client/
│   ├── client.go         # Main client orchestrator
│   ├── events/           # System event handling
│   ├── lua/             # Lua scripting engine
│   │   ├── bindings.go  # Lua API bindings
│   │   ├── line.go      # Line processing
│   │   └── core/        # Core Lua scripts
│   ├── connection/      # Telnet handling
│   ├── terminal/        # Terminal UI
│   └── types/          # Shared types
└── protocol/
    └── telnet/         # Telnet protocol
```

## Core Components

### 1. Client (pkg/client/client.go)
- Top-level coordinator
- Owns all components
- Manages lifecycle
- Routes system events

### 2. Lua Engine (pkg/client/lua/)
- Central processor for all I/O
- Manages user scripts
- Provides API bindings
- Handles triggers/aliases

### 3. Connection (pkg/client/connection/)
- Handles telnet connection
- Raw I/O processing
- ANSI handling
- Protocol negotiation

### 4. Terminal UI (pkg/client/terminal/)
- User interface
- Input handling
- Output display
- Status bar

## Data Flow

### Input Flow (User → MUD)
```
[User Input] → [Terminal]
    → [Client.handleInput]
    → [LuaEngine.ProcessInput] (synchronous)
        → Process through input listeners
        → Modify/gag if needed
    → [Connection.Write] → [MUD Server]
```

### Output Flow (MUD → User)
```
[MUD Server] → [Connection.Read]
    → [OutputProcessor]
    → [LuaEngine.ProcessOutput] (synchronous)
        → Process through output listeners
        → Modify/gag if needed
    → [Terminal.Display]
```

## Line Processing

### Line Object
```go
type Line struct {
    Content     string     // Processed content
    Raw         string     // Raw ANSI content
    Flags       LineFlags  // Processing flags
    Replacement *string    // Optional replacement
}

type LineFlags struct {
    Gag         bool    // Don't display
    Matched     bool    // Matched by trigger
    IsPrompt    bool    // Is prompt line
    SkipLog     bool    // Don't log
    Source      string  // Line source
}
```

### Lua Processing Chain
```lua
-- Output processing
mud.add_output_listener(function(line)
    if line:match("Health") then
        line:gag(true)  -- Don't display
        return line
    end
    return line
end)

-- Input processing
mud.add_input_listener(function(line)
    if line:match("^gg$") then
        line:replace("say good game!")
    end
    return line
end)
```

## Event System

The event system is used only for system operations, not for I/O processing:

### System Events
```go
const (
    // Connection
    EventConnect      // Request connection
    EventConnected    // Connection established
    EventDisconnect   // Request disconnect
    EventDisconnected // Connection closed
    
    // UI
    EventRedraw       // Request UI redraw
    EventResize       // Terminal resize
    EventScroll       // Scroll viewport
    
    // Other
    EventLog          // Log message
    EventDebug        // Debug message
    EventQuit         // Quit application
)
```

## Scripting API

### Core API
```lua
-- Connection
mud.connect(host, port)
mud.disconnect()
mud.reconnect()

-- I/O
mud.send(text)
mud.output(text)

-- Triggers
trigger.add("^Health: (\\d+)$", {gag = true}, function(matches)
    -- Process health update
end)

-- Aliases
alias.add("^gg$", function()
    mud.send("say good game!")
end)
```

### Line API
```lua
line:raw()      -- Get raw ANSI text
line:display()  -- Get clean text
line:gag()      -- Get/set gag flag
line:prompt()   -- Get/set prompt flag
line:matched()  -- Get/set matched flag
line:replace()  -- Replace content
```

## Configuration

### Directory Structure
```
~/.runes/
├── config/
│   └── config.lua    # User config
├── scripts/
│   └── user/         # User scripts
└── logs/            # Session logs
```

### Core Scripts
```
pkg/client/lua/core/
├── alias.lua       # Alias system
├── trigger.lua     # Trigger system
├── timers.lua      # Timer system
└── events.lua      # Event handling
```

## Error Handling

1. **Lua Errors**
- Caught and logged
- Don't crash the client
- Reported to user
- Processing continues

2. **Connection Errors**
- Auto-reconnect support
- Error events emitted
- User notification
- Clean disconnect

3. **Processing Errors**
- Timeouts for long-running scripts
- Error logging
- Skip problematic processors
- Continue chain execution

## Performance Considerations

1. **Direct Processing**
- Synchronous I/O processing
- No event overhead for core operations
- Predictable execution order

2. **Event System**
- Used only for system operations
- Async where appropriate
- Buffered channels
- Non-blocking handlers

3. **Memory Management**
- Line object pooling
- Buffer reuse
- Lua garbage collection tuning
- Resource cleanup
