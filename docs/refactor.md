# Output Processing Refactor

## Package Design

### Revised Package Structure
```
pkg/
├── protocol/           # Only truly protocol-level code
│   └── telnet/        # Generic telnet implementation
│
└── client/            # Everything MUD-client specific
    ├── ansi/          # ANSI processing
    ├── buffer/        # Output buffer
    ├── connection/    # Connection handling
    ├── events/        # Event system
    ├── history/       # Command history
    ├── lua/           # Lua scripting
    ├── terminal/      # Terminal handling
    └── ui/            # User interface
```

### Design Rationale
1. Most functionality is MUD-client specific
2. Only telnet protocol handling is truly generic
3. Better to acknowledge coupling than force false separation
4. Follows Go's standard library patterns (e.g., net/http/httputil)

## Package Name Updates

### Required Changes
All package declarations and imports need to be updated to reflect the new structure:

1. **Protocol Package**
```go
// Before
package telnet
import "github.com/mmcdole/runes/pkg/protocol/telnet"

// After - no change needed
package telnet
import "github.com/mmcdole/runes/pkg/protocol/telnet"
```

2. **Client Packages**
```go
// Before
package ansi
import "github.com/mmcdole/runes/pkg/ansi"

// After
package ansi
import "github.com/mmcdole/runes/pkg/client/ansi"
```

### Full Package Update List
1. `pkg/ansi` → `pkg/client/ansi`
   - Update in all files using ANSI processing
   - Check for any external tools/scripts importing this package

2. `pkg/events` → `pkg/client/events`
   - Update all event emitters and handlers
   - Check Lua bindings that might reference events

3. `pkg/luaengine` → `pkg/client/lua`
   - Update package name and directory
   - Update all imports in Lua-related code
   - Check any script files that might have hardcoded paths

4. `pkg/buffer` → `pkg/client/buffer`
   - Update buffer package references
   - Check viewport and UI components that use buffer

### Files to Check
- All .go files for import statements
- go.mod and go.sum
- Any build scripts or tools
- Documentation references
- Test files
- Example code
- Integration tests

### Verification Steps
1. Run `go mod tidy` after moves
2. Check for broken imports: `go build ./...`
3. Run all tests: `go test ./...`
4. Verify external tools still work
5. Check documentation is up to date

### Common Places for Package References
1. Import statements
2. Type assertions
3. Documentation comments
4. Test files
5. Example code
6. Build tags
7. Interface definitions
8. Mock implementations

### Tools to Help
```bash
# Find all Go files
find . -name "*.go" -type f

# Search for old import paths
grep -r "github.com/mmcdole/runes/pkg/" .

# Check for package declarations
grep -r "^package " .
```

Remember to:
1. Update one package at a time
2. Run tests after each package move
3. Commit changes in logical groups
4. Update documentation as you go
5. Verify external tools still work

## Connection Output Processing

### pkg/client/connection/output.go

```go
type Mode int

const (
    ModeNormal Mode = iota
    ModePrompt
)

type OutputProcessor struct {
    conn     *protocol.TelnetConnection
    buffer   []byte
    mode     Mode
    events   *events.EventProcessor
    maxPromptSize int  // Size threshold for prompt detection (default 500)
}

type Line struct {
    Content   string
    Timestamp time.Time
    IsPrompt  bool
}
```

### Line Processing Behavior

1. **Buffer Management**
```go
func (p *OutputProcessor) Write(data []byte) {
    p.buffer = append(p.buffer, data...)
    p.processBuffer()
}
```

2. **Line Detection**
```go
func (p *OutputProcessor) processBuffer() {
    // Look for line endings (\n, \r\n, \n\r)
    // Keep partial lines in buffer
    // Handle prompts for small buffers
}
```

3. **Prompt Detection** (based on Blightmud's approach)
```go
func (p *OutputProcessor) checkPrompt() bool {
    // If buffer is small (< maxPromptSize)
    // And contains no newline
    // Consider it a potential prompt
    if len(p.buffer) < p.maxPromptSize && !bytes.Contains(p.buffer, []byte{'\n'}) {
        p.emitLine(true)  // Emit as prompt
        return true
    }
    return false
}
```

4. **Line Emission**
```go
func (p *OutputProcessor) emitLine(isPrompt bool) {
    line := &Line{
        Content:   string(p.buffer),
        Timestamp: time.Now(),
        IsPrompt:  isPrompt,
    }
    
    // For prompts, don't clear buffer in case more data comes
    if !isPrompt {
        p.buffer = p.buffer[:0]
    }
    
    p.events.Emit(events.RawLineEvent{Line: line})
}
```

### Integration with Client

1. **Client Setup**
```go
type Client struct {
    outputProc *connection.OutputProcessor
    // ... other fields
}

func NewClient() *Client {
    c := &Client{
        outputProc: connection.NewOutputProcessor(
            telnetConn,
            events,
            connection.WithMaxPromptSize(500),
        ),
    }
    // ... other setup
}
```

2. **Event Flow**
```
OutputProcessor -> RawLineEvent -> Lua Processing -> ANSI Processing -> Display
```

### Key Features

1. **Prompt Handling**
- Small buffers (< 500 bytes) without newlines treated as prompts
- Prompts emitted but kept in buffer
- Handles cases like "Password:" properly

2. **Line Buffering**
- Accumulates partial lines until complete
- Only emits on newline or prompt condition
- Maintains buffer across reads

3. **Event Integration**
- Emits RawLineEvent for complete lines
- Includes prompt flag in events
- Preserves timing information

4. **Configuration**
- Configurable prompt size threshold
- Adjustable buffer sizes
- Debug logging options

## Implementation Steps

1. Create OutputProcessor
   - Basic buffer management
   - Line detection
   - Event emission

2. Add Prompt Detection
   - Size threshold checking
   - Newline scanning
   - Prompt event handling

3. Integrate with Client
   - Remove old buffering code
   - Wire up event handling
   - Update ANSI processing

4. Add Testing
   - Test partial line handling
   - Test prompt detection
   - Test various line endings

5. Add Monitoring/Debug
   - Buffer size tracking
   - Prompt detection logging
   - Performance metrics
