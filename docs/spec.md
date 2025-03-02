# Runes Specification

## Overview
Runes is a modern MUD client built in Go with an embedded Lua scripting engine. It features a terminal user interface (TUI) and a sophisticated event-driven architecture that manages bidirectional data flow between the MUD server, Go core, Lua scripting engine, and client interface.

## Core Architecture

### Package Structure

```
pkg/
├── protocol/          # Network protocol implementations
│   └── telnet/           # Telnet protocol implementation with IAC command handling
│                       
└── client/           # MUD client implementation
    ├── ansi/             # ANSI escape sequence parsing and color handling
    ├── buffer/           # Circular buffer for managing output history
    ├── connection/       # Network connection and line-based output processing
    ├── events/           # Event dispatcher and handler registration
    ├── history/          # Command history with search and persistence
    ├── lua/              # Lua scripting and API integration
    │   └── core/             # Core Lua scripts and default functionality
    │       ├── init.lua          # Core initialization and environment setup
    │       ├── input.lua         # Input processing and command splitting
    │       ├── alias.lua         # Command alias expansion and management
    │       ├── commands.lua      # Built-in command implementations
    │       ├── trigger.lua       # Pattern matching and response system
    │       ├── timer.lua         # Scheduled event management
    │       ├── events.lua        # Event registration and dispatch
    │       └── defaults.lua      # Default client settings
    │                       
    ├── terminal/         # Terminal input handling and raw mode management
    └── ui/               # Terminal user interface
        ├── components/       # Individual UI elements (input bar, status line, etc)
        └── layout/           # Screen layout and window management
```

### Design Rationale
1. Most functionality is MUD-client specific and lives under `pkg/client/`
2. Only telnet protocol handling is truly generic and lives in `pkg/protocol/`
3. Better organization acknowledges natural coupling between MUD-specific components
4. Follows Go's standard library patterns (e.g., net/http/httputil)

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
   - Data processed by OutputProcessor for line detection
   - Lines emitted as `EventRawOutput` with prompt detection
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

### Connection System (`pkg/client/connection`)
Connection and output processing providing:
- Telnet connection management
- Line-based output processing
- Prompt detection
- Event emission
- Buffer management

### UI System (`pkg/client/ui`)
Display management providing:
- Window content rendering
- Scroll position tracking
- Partial screen updates
- Dynamic resizing
- View attribute handling

### Lua Engine (`pkg/client/lua`)
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

### Event System (`pkg/client/events`)
Manages bi-directional communication:
1. Go Events:
   - `EventRawInput`: User input
   - `EventCommand`: Processed commands
   - `EventRawOutput`: Server data (with line/prompt detection)
   - `EventOutput`: Processed output
   - System events (connect, disconnect, quit)

2. Lua Events:
   - Input processing
   - Output handling
   - Buffer management
   - Connection control
   - Timer operations
   - Variable management

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
