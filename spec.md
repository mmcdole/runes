# Runes Specification

## Overview
Runes is a modern MUD client built in Go with an embedded Lua scripting engine. It features a terminal user interface (TUI) and a sophisticated event-driven architecture that manages bidirectional data flow between the MUD server, Go core, Lua scripting engine, and client interface.

## Core Architecture

### Package Structure

```
pkg/
├── client/           # Core client implementation
│   ├── buffer/       # Thread-safe line buffer management
│   ├── viewport/     # Display and window management
│   └── tui/         # Terminal UI using Bubble Tea
├── events/          # Event system for inter-component communication
├── luaengine/       # Lua scripting and event processing
│   ├── core/        # Built-in core Lua scripts
│   │   ├── init.lua       # Core initialization
│   │   ├── input.lua      # Input processing and command splitting
│   │   ├── alias.lua      # Alias expansion system
│   │   ├── commands.lua   # Built-in command handlers
│   │   ├── trigger.lua    # Trigger system
│   │   ├── timer.lua      # Timer management
│   │   ├── events.lua     # Event handling
│   │   └── defaults.lua   # Default settings
│   └── bindings.go  # Go-Lua bridge and API definitions
└── protocol/        # Network protocol implementations
    └── telnet/      # Telnet protocol handler
```

## Data Flow Architecture

### Input Flow (User → MUD Server)
1. User Input Capture
   - TUI captures raw keyboard input
   - Input is wrapped in `EventRawInput`
   - Event dispatched to input processing pipeline

2. Lua Input Processing
   - Raw input passed to Lua engine
   - Scripts process for:
     * Command aliases
     * Input transformation
   - Results in `EventCommand`

3. Command Execution
   - Processed command sent to protocol handler (Telnet)
   - Data transmitted to MUD server

### Output Flow (MUD Server → User)
1. Server Data Reception
   - Protocol handler receives raw server data
   - Data wrapped in `EventRawOutput`
   - Passed to output processing pipeline

2. Lua Output Processing
   - Raw output processed by Lua scripts for:
     * Trigger matching
     * Text highlighting
     * Line filtering/gagging
   - Results in `EventOutput`

3. Display Rendering
   - Processed output sent to viewport
   - Buffer system manages line storage
   - Viewport handles display and scrolling
   - TUI renders final output

## Component Details

### Client (`pkg/client`)
Central coordinator managing:
- Application lifecycle
- Component initialization
- Event routing
- Network connections
- Buffer management

### Buffer System (`pkg/client/buffer`)
Efficient line management featuring:
- Thread-safe operations
- Circular buffer implementation
- Partial retrieval support
- History management
- Line attribute storage

### Viewport (`pkg/client/viewport`)
Display management providing:
- Window content rendering
- Scroll position tracking
- Partial screen updates
- Dynamic resizing
- View attribute handling

### Lua Engine (`pkg/luaengine`)
Script processing engine offering:
- Event system bridge
- Go function bindings
- Script loading/management
- Sandboxed execution
- Error handling

Core Script System:
- Built-in scripts loaded at startup:
  * Command parsing and splitting (default separator ";")
  * Alias system for command expansion
  * Trigger system for output processing
  * Timer management for scheduled actions
  * Event handling for system integration
  * Default settings and configurations
- Builds the Runes Lua API:
  * Core event system (add/emit handlers)
  * Input processing functions
  * Alias management
  * Trigger registration
  * Timer control
  * Command system
  * Default configurations

User Script System:
- Scripts loaded via command or from configured directory
- Supports runtime loading and reloading
- Utilizes Runes Lua API for:
  * Custom command definitions
  * Personal aliases and triggers
  * User-defined timers and events
  * Configuration overrides
- Persistent between sessions

### Event System
Manages bi-directional communication:
1. Go Events:
   - `EventRawInput`: User input
   - `EventCommand`: Processed commands
   - `EventRawOutput`: Server data
   - `EventOutput`: Processed output
   - System events (connect, disconnect, quit)

2. Lua Events:
   - Input processing
   - Output handling

## Extension Points

### Scripting API
Comprehensive Lua API for:
- Event handling
- Input/Output processing
- Buffer management
- Connection control
- Timer operations
- Variable management

### Protocol Support
Extensible protocol system:
- Telnet implementation
- Custom protocol support
- Negotiation handling
- Data transformation
