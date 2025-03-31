package client

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "time"

    "github.com/mmcdole/runes/pkg/buffer"
    "github.com/mmcdole/runes/pkg/events"
    "github.com/mmcdole/runes/pkg/network/telnet"
    "github.com/mmcdole/runes/pkg/scripting"
    "github.com/mmcdole/runes/pkg/types"
    "github.com/mmcdole/runes/pkg/client/ui/components"
    "github.com/mmcdole/runes/pkg/client/ui/layout"
)

// Client represents the MUD client
type Client struct {
    // Core components
    eventSystem events.EventSystem
    luaEngine   *scripting.LuaEngine
    conn        *telnet.TelnetConnection
    
    // UI components
    layout    *layout.Layout
    viewport  *components.Viewport
    input     *components.Input
    status    *components.Status
    
    // State
    buffer     *buffer.Buffer
    scriptPath string
    quit       bool
}

// New creates a new client
func New(scriptPath string) (*Client, error) {
    // Create client
    client := &Client{
        eventSystem: events.New(),
        scriptPath:  scriptPath,
        buffer:      buffer.New(10000), // 10k line buffer
    }

    // Create UI layout
    client.layout = layout.New()

    // Create viewport
    client.viewport = components.NewViewport(client.buffer)
    if err := client.layout.RegisterComponent("viewport", client.viewport); err != nil {
        return nil, fmt.Errorf("failed to register viewport: %w", err)
    }

    // Create input
    client.input = components.NewInput()
    if err := client.layout.RegisterComponent("input", client.input); err != nil {
        return nil, fmt.Errorf("failed to register input: %w", err)
    }

    // Create status
    client.status = components.NewStatus()
    if err := client.layout.RegisterComponent("status", client.status); err != nil {
        return nil, fmt.Errorf("failed to register status: %w", err)
    }

    // Create Lua engine
    var err error
    client.luaEngine, err = scripting.NewLuaEngine(client.eventSystem, client.scriptPath)
    if err != nil {
        return nil, fmt.Errorf("failed to create Lua engine: %w", err)
    }

    // Subscribe to events
    client.subscribeEvents()

    return client, nil
}

// Run starts the client
func (c *Client) Run() error {
    // Load core scripts
    if err := c.loadCoreScripts(); err != nil {
        return fmt.Errorf("failed to load core scripts: %w", err)
    }

    // Load user scripts
    if c.scriptPath != "" {
        if err := c.loadUserScripts(); err != nil {
            return fmt.Errorf("failed to load user scripts: %w", err)
        }
    }

    // Run main loop
    for !c.quit {
        // Update UI
        if err := c.layout.Draw(); err != nil {
            return fmt.Errorf("failed to draw UI: %w", err)
        }

        // Sleep to prevent CPU spinning
        time.Sleep(16 * time.Millisecond)
    }

    return nil
}

// Close cleans up the client
func (c *Client) Close() {
    if c.conn != nil {
        c.conn.Close()
    }
    if c.luaEngine != nil {
        c.luaEngine.Close()
    }
}

// Event handlers
func (c *Client) handleInput(event events.Event) {
    text := event.Data.(string)

    // Create line
    line := types.NewClientLine(text)

    // Process through Lua
    c.luaEngine.ProcessInput(line)

    // Send to server
    if c.conn != nil && !line.Flags.Gag {
        c.conn.Write([]byte(line.Raw))
    }

    // Add to buffer
    c.buffer.Add(line)
}

func (c *Client) handleConnect(event events.Event) {
    data := event.Data.(struct {
        Host string
        Port int
    })

    // Create connection
    var err error
    c.conn, err = telnet.NewTelnetConnection(data.Host, data.Port, false)
    if err != nil {
        c.status.SetError(fmt.Sprintf("Failed to connect: %v", err))
        return
    }

    // Connection is ready to use after creation

    c.status.SetMessage(fmt.Sprintf("Connected to %s:%d", data.Host, data.Port))
}

func (c *Client) handleDisconnect(event events.Event) {
    if c.conn != nil {
        c.conn.Close()
        c.conn = nil
    }
    c.status.SetMessage("Disconnected")
}

func (c *Client) handleQuit(event events.Event) {
    c.quit = true
}

func (c *Client) handleResize(event events.Event) {
    if err := c.layout.Resize(); err != nil {
        c.status.SetError(fmt.Sprintf("Failed to resize: %v", err))
    }
}

// Helpers
func (c *Client) loadCoreScripts() error {
    // Get core script path
    dir, err := os.Executable()
    if err != nil {
        return fmt.Errorf("failed to get executable path: %w", err)
    }
    corePath := filepath.Join(filepath.Dir(dir), "core")

    // Load core scripts
    entries, err := os.ReadDir(corePath)
    if err != nil {
        return fmt.Errorf("failed to read core scripts: %w", err)
    }

    for _, entry := range entries {
        if !entry.IsDir() && filepath.Ext(entry.Name()) == ".lua" {
            path := filepath.Join(corePath, entry.Name())
            if err := c.luaEngine.LoadUserScript(path); err != nil {
                return fmt.Errorf("failed to load core script %s: %w", entry.Name(), err)
            }
            log.Printf("[Client] Loaded core script: %s", entry.Name())
        }
    }

    return nil
}

func (c *Client) loadUserScripts() error {
    // Load user scripts
    entries, err := os.ReadDir(c.scriptPath)
    if err != nil {
        return fmt.Errorf("failed to read user scripts: %w", err)
    }

    for _, entry := range entries {
        if !entry.IsDir() && filepath.Ext(entry.Name()) == ".lua" {
            path := filepath.Join(c.scriptPath, entry.Name())
            if err := c.luaEngine.LoadUserScript(path); err != nil {
                return fmt.Errorf("failed to load user script %s: %w", entry.Name(), err)
            }
            log.Printf("[Client] Loaded user script: %s", entry.Name())
        }
    }

    return nil
}

func (c *Client) subscribeEvents() {
    // Input events
    c.eventSystem.Subscribe(events.EventInput, c.handleInput)
    
    // Connection events
    c.eventSystem.Subscribe(events.EventConnect, c.handleConnect)
    c.eventSystem.Subscribe(events.EventDisconnect, c.handleDisconnect)
    
    // System events
    c.eventSystem.Subscribe(events.EventQuit, c.handleQuit)
    c.eventSystem.Subscribe(events.EventResize, c.handleResize)
}
