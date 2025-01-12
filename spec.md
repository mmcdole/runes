# Runes Specification

## Overview
Runes is a modern MUD client built in Go with an embedded Lua scripting engine. It features a terminal user interface (TUI) and a dual event-driven architecture where events flow between Go and Lua systems for comprehensive input/output processing.

## Core Architecture

### Package Structure

```
pkg/
├── client/           # Client implementation
│   ├── buffer/       # Line buffer management
│   ├── viewport/     # Display viewport
│   └── tui/         # Terminal UI (Bubble Tea)
├── events/          # Go event system
├── luaengine/       # Lua scripting engine
│   ├── core/        # Core Lua scripts
│   └── bindings.go  # Go-Lua bridge
└── protocol/        # Network protocols
    └── telnet/      # Telnet implementation
```

## Event Systems and Flow

### Core Event Types

#### Input-Related Events
- `EventRawInput`: Unprocessed user input from TUI
- `EventCommand`: Processed user input after Lua engine handling
  - Aliases expanded
  - Commands interpreted

#### Output-Related Events
- `EventRawOutput`: Raw server output from telnet
- `EventOutput`: Processed server output after Lua engine handling
  - Triggers processed
  - Text highlighted
  - Lines gagged/filtered

#### System Events
- `EventConnect`/`EventDisconnect`: Connection management
- `EventQuit`: Client shutdown

### Event Processing Details

#### Input Processing
1. User types in TUI
2. TUI emits `EventRawInput`
3. Lua engine processes input:
   - Checks for aliases
   - Expands variables
   - Processes script commands
4. Lua engine emits `EventCommand` with:
   - Processed command text
   - Any additional actions
5. Client executes the command

#### Output Processing
1. Server sends data through telnet
2. Client emits `EventRawOutput`
3. Lua engine processes output:
   - Checks trigger patterns
   - Applies highlighting rules
   - Handles gags/filters
   - Processes ANSI codes
4. For each trigger match:
   - Executes trigger actions
   - May generate additional output
5. Lua engine emits `EventOutput` with:
   - Processed text
   - Highlighting information
   - Display attributes
6. Viewport renders the processed output

### Lua Event System (`pkg/luaengine/core/events.lua`)
- Manages script-level events and callbacks
- Provides event registration and emission within Lua
- Allows scripts to:
  - Register trigger patterns
  - Define aliases
  - Process ANSI colors
  - Filter output
  - Generate new output

### Event Flow Bridge
The two event systems are bridged through the LuaEngine:
1. Go → Lua:
   - LuaEngine subscribes to Go events
   - Transforms events into Lua events via `emitLuaEvent`
   - Core Lua scripts process events
   
2. Lua → Go:
   - Lua scripts call bound functions (e.g., `output`, `connect`)
   - Bindings emit appropriate Go events
   - Go components handle the events

## Component Details

### Client (`pkg/client`)
- Manages application lifecycle
- Coordinates between components
- Handles network connections
- Routes events between systems

### Buffer System (`pkg/client/buffer`)
- Thread-safe line storage
- Fixed-size circular buffer
- Efficient line management
- Supports partial retrieval for viewport

### Viewport System (`pkg/client/viewport`)
- Manages visible content display
- Efficient partial updates
- Scroll position tracking
- Window resize handling

### TUI System (`pkg/client/tui`)
- Built on Bubble Tea
- Input handling
- Screen rendering
- Status line management

### Lua Engine (`pkg/luaengine`)
Core responsibilities:
- Initializes Lua state
- Loads core and user scripts
- Provides Go function bindings
- Manages event bridging

Key components:
1. Core Modules:
   - `events.lua`: Event system
   - Other core functionality modules

2. Bindings:
   - Connection management
   - Output handling
   - Buffer management
   - Client control

3. Script Management:
   - Core script embedding
   - User script loading
   - Error handling
   - Sandbox security

## Data Flow

### Input Processing
1. User input → TUI
2. TUI emits Go `EventRawInput`
3. LuaEngine bridges to Lua event
4. Lua scripts process input
5. Scripts may:
   - Emit output events
   - Trigger commands
   - Modify input
   - Handle locally

### Output Processing
1. Server data → Telnet
2. Client emits Go `EventRawOutput`
3. LuaEngine bridges to Lua event
4. Lua scripts process output
5. Scripts emit processed output
6. Go `EventOutput` triggered
7. Client renders in viewport

## Extension Points

### Script API
- Event handlers
- Custom commands
- Output processing
- Buffer management
- Connection handling

### Protocol Support
- Telnet extensions
- New protocol implementations
- Custom negotiation

### UI Customization
- Status line content
- Color schemes
- Layout options
- Custom widgets

## Security

### Script Sandboxing
- Limited system access
- Resource constraints
- Error isolation
- Safe defaults

### Network Security
- Protocol validation
- Input sanitization
- Connection safety

## Performance

### Buffer Management
- Fixed size buffers
- Efficient line storage
- Minimal copying
- Thread safety

### Viewport Optimization
- Partial updates
- Visible-only rendering
- Scroll position caching
- Resize efficiency

### Event Processing
- Async event handling
- Event batching
- Efficient bridging
- Memory management
