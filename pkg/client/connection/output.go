package connection

import (
    "bytes"

    "github.com/mmcdole/runes/pkg/client/ansi"
    "github.com/mmcdole/runes/pkg/client/events"
    "github.com/mmcdole/runes/pkg/client/lua"
    "github.com/mmcdole/runes/pkg/client/types"
    "github.com/mmcdole/runes/pkg/protocol/telnet"
)

// OutputProcessor handles the processing of telnet output into lines
type OutputProcessor struct {
    conn        *telnet.TelnetConnection
    buffer      []byte
    events      events.EventSystem
    luaEngine   *lua.LuaEngine
    ansiProc    *ansi.Processor
}

// NewOutputProcessor creates a new output processor
func NewOutputProcessor(conn *telnet.TelnetConnection, events events.EventSystem, luaEngine *lua.LuaEngine) *OutputProcessor {
    return &OutputProcessor{
        conn:      conn,
        events:    events,
        luaEngine: luaEngine,
        buffer:    make([]byte, 0, 4096),
        ansiProc:  ansi.NewProcessor(),
    }
}

// Start begins processing output from the telnet connection
func (p *OutputProcessor) Start() {
    go p.processOutput()
}

// processOutput reads from the telnet connection and processes the output
func (p *OutputProcessor) processOutput() {
    buf := make([]byte, 4096)
    for {
        n, err := p.conn.Read(buf)
        if err != nil {
            // Connection closed or error
            p.events.Emit(events.Event{
                Type: events.EventDisconnected,
            })
            return
        }

        // Process the received data
        p.Process(buf[:n])
    }
}

// Process processes a chunk of output from the MUD
func (p *OutputProcessor) Process(data []byte) {
    // Append new data to existing buffer
    p.buffer = append(p.buffer, data...)

    // Process complete lines
    for {
        i := bytes.IndexByte(p.buffer, '\n')
        if i == -1 {
            break
        }

        // Extract line content
        content := string(p.buffer[:i])
        p.buffer = p.buffer[i+1:]

        // Create line
        line := types.NewLine(content)
        line.Display = p.ansiProc.Process(content)
        line.Flags.Complete = true

        // Process through Lua
        if processed := p.luaEngine.ProcessOutput(line); processed != nil {
            // Line wasn't gagged, emit redraw event
            p.events.Emit(events.Event{
                Type: events.EventRedraw,
                Data: processed,
            })
        }
    }

    // Check remaining buffer for prompt
    if len(p.buffer) > 0 {
        content := string(p.buffer)
        if bytes.HasSuffix(p.buffer, []byte{telnet.GA}) || bytes.HasSuffix(p.buffer, []byte{telnet.EOR}) {
            // Create prompt line
            line := types.NewPrompt(content)
            line.Display = p.ansiProc.Process(content)

            // Process through Lua
            if processed := p.luaEngine.ProcessOutput(line); processed != nil {
                // Line wasn't gagged, emit redraw event
                p.events.Emit(events.Event{
                    Type: events.EventRedraw,
                    Data: processed,
                })
            }

            // Clear buffer after prompt
            p.buffer = p.buffer[:0]
        }
    }
}

// Close cleans up the output processor
func (p *OutputProcessor) Close() {
    // Process any remaining data
    if len(p.buffer) > 0 {
        content := string(p.buffer)
        line := types.NewLine(content)
        line.Display = p.ansiProc.Process(content)

        // Process through Lua
        if processed := p.luaEngine.ProcessOutput(line); processed != nil {
            p.events.Emit(events.Event{
                Type: events.EventRedraw,
                Data: processed,
            })
        }
    }

    p.buffer = nil
}
