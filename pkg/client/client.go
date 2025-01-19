package client

import (
	"fmt"
	"io"
	"strings"

	"github.com/mmcdole/runes/pkg/client/buffer"
	"github.com/mmcdole/runes/pkg/client/history"
	"github.com/mmcdole/runes/pkg/client/terminal"
	"github.com/mmcdole/runes/pkg/client/ui/components"
	"github.com/mmcdole/runes/pkg/client/ui/layout"
	"github.com/mmcdole/runes/pkg/events"
	"github.com/mmcdole/runes/pkg/luaengine"
	"github.com/mmcdole/runes/pkg/protocol/telnet"
)

// Client represents a MUD client
type Client struct {
	// Terminal and UI components
	term      *terminal.Terminal
	layout    *layout.Manager
	buffer    *buffer.Buffer
	statusBar *components.StatusBar
	viewport  *components.Viewport
	inputBar  *components.InputBar
	history   *history.History

	// Connection handling
	host string
	port int
	conn *telnet.TelnetConnection

	// Event and script handling
	events *events.EventProcessor
	engine *luaengine.LuaEngine

	// State
	running bool
	debug   bool
	done    chan struct{}
}

// Config holds the client configuration
type Config struct {
	Host          string
	Port          int
	UserScriptDir string
	Debug         bool
}

// NewClient creates a new client instance
func NewClient(userScriptDir string, eventProcessor *events.EventProcessor, config Config, debug bool) (*Client, error) {
	term := terminal.New()
	buf := buffer.New()

	client := &Client{
		// Terminal and UI
		term:   term,
		buffer: buf,

		// Connection
		host: config.Host,
		port: config.Port,

		// Event handling
		events: eventProcessor,
		debug:  debug,
		done:   make(chan struct{}),
	}

	// Create layout manager with standard policy
	client.layout = layout.NewManager(layout.NewStandardPolicy())

	// Create components
	client.statusBar = components.NewStatusBar(term)
	client.viewport = components.NewViewport(term, buf)
	client.history = history.New() // Create history instance
	client.inputBar = components.NewInputBar(term, client.history)

	// Register components with layout manager
	client.layout.RegisterComponent(layout.StatusBarType, client.statusBar)
	client.layout.RegisterComponent(layout.ViewportType, client.viewport)
	client.layout.RegisterComponent(layout.InputBarType, client.inputBar)

	// Initialize Lua engine
	engine := luaengine.New(userScriptDir, eventProcessor)
	if err := engine.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize lua engine: %w", err)
	}
	client.engine = engine

	// Set up event handlers
	client.setupEventHandlers()

	return client, nil
}

// Run starts the client
func (c *Client) Run() error {
	if err := c.term.Init(); err != nil {
		return fmt.Errorf("failed to initialize terminal: %w", err)
	}
	defer c.term.Cleanup()

	c.running = true

	// Initialize UI
	width, height := c.term.Size()
	c.term.Clear()
	c.layout.RenderAll(width, height)
	c.updateStatus("Ready")

	// Input buffer
	buf := make([]byte, 1024)

	// Main event loop
	for c.running {
		select {
		case <-c.term.ResizeChan():
			width, height := c.term.Size()
			c.layout.HandleResize(width, height)
			c.updateStatus("Connected")
		case <-c.done:
			return nil
		default:
			n, err := c.term.Read(buf)
			if err != nil {
				if err != io.EOF {
					return fmt.Errorf("read error: %w", err)
				}
				continue
			}
			if n > 0 {
				if err := c.handleInput(buf[:n]); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// updateStatus updates the status bar text
func (c *Client) updateStatus(status string) {
	width, height := c.term.Size()
	mode := c.viewport.GetMode()
	c.statusBar.SetText(fmt.Sprintf("Runes - %dx%d - %s - %s", width, height, mode, status))
}

// handleInput processes all input
func (c *Client) handleInput(input []byte) error {
	// Check for special key combinations first
	if c.handleSpecialKeys(input) {
		return nil
	}

	// Let inputbar handle all input
	handled := c.inputBar.HandleInput(input)

	// If it was handled and it was Enter, send the command
	if handled && terminal.IsEnter(input) {
		content := c.inputBar.GetContent()
		if content != "" {
			// First emit raw input for Lua processing
			c.events.Emit(events.Event{
				Type: events.EventRawInput,
				Data: content,
			})
			c.history.Add(content) // Add command to history
			c.inputBar.Clear()
		}
	}
	return nil
}

// handleSpecialKeys handles special key combinations
func (c *Client) handleSpecialKeys(input []byte) bool {
	// Ctrl+C - quit immediately
	if terminal.IsCtrlC(input) {
		c.running = false
		c.Close()
		return true
	}

	// Up/Down arrows for history
	if terminal.IsUpArrow(input) {
		if c.inputBar.HandleInput(input) {
			return true
		}
	}
	if terminal.IsDownArrow(input) {
		if c.inputBar.HandleInput(input) {
			return true
		}
	}

	// Page Up/Down for viewport scrolling
	if terminal.IsPageUp(input) {
		c.viewport.ScrollUp()
		c.updateStatus("")
		return true
	}
	if terminal.IsPageDown(input) {
		c.viewport.ScrollDown()
		c.updateStatus("")
		return true
	}

	return false
}

// handleTelnetOutput processes output from the telnet connection
func (c *Client) handleTelnetOutput() {
	buf := make([]byte, 4096)
	line := ""
	for {
		select {
		case <-c.done:
			return
		default:
			n, err := c.conn.Read(buf)
			if err != nil {
				if err == io.EOF {
					c.events.Emit(events.Event{
						Type: events.EventRawOutput,
						Data: "Connection closed by server\n",
					})
					c.Close()
					return
				}
				c.events.Emit(events.Event{
					Type: events.EventRawOutput,
					Data: fmt.Sprintf("Error reading from server: %v\n", err),
				})
				continue
			}

			// Add new data to line buffer
			line += string(buf[:n])

			// Process complete lines
			for {
				i := strings.Index(line, "\n")
				if i < 0 {
					break
				}
				// Emit complete line without the newline
				c.events.Emit(events.Event{
					Type: events.EventRawOutput,
					Data: line[:i], // Don't include the newline
				})
				line = line[i+1:]
			}

			// If there's remaining data without a newline, emit it too
			if len(line) > 0 {
				c.events.Emit(events.Event{
					Type: events.EventRawOutput,
					Data: line,
				})
			}
		}
	}
}

// processServerOutput handles a line of server output
func (c *Client) processServerOutput(line string) {
	// Emit raw output event for Lua processing
	c.events.Emit(events.Event{
		Type: events.EventRawOutput,
		Data: line,
	})
}

// handleCommand handles processed command events
func (c *Client) handleCommand(e events.Event) {
	if cmd, ok := e.Data.(string); ok {
		// Send command to telnet connection
		if c.conn != nil {
			c.conn.Write([]byte(cmd + "\n"))
		}
	}
}

// handleProcessedOutput handles processed output events
func (c *Client) handleProcessedOutput(e events.Event) {
	if output, ok := e.Data.(struct {
		Text   string
		Buffer string
	}); ok {
		// Output text already has ANSI codes from Lua
		c.buffer.Write(output.Text + "\n")
		c.viewport.UpdateView()
		width, height := c.term.Size()
		c.layout.RenderAll(width, height)
	}
}

// handleLog handles log events
func (c *Client) handleLog(e events.Event) {
	if msg, ok := e.Data.(string); ok {
		// Add log prefix with ANSI color
		c.buffer.Write("\033[1;34mLog:\033[0m " + msg + "\n")
		c.viewport.UpdateView()
		width, height := c.term.Size()
		c.layout.RenderAll(width, height)
	}
}

// handleDebug handles debug events
func (c *Client) handleDebug(e events.Event) {
	if c.debug {
		if msg, ok := e.Data.(string); ok {
			// Add debug prefix with ANSI color
			c.buffer.Write("\033[1;35mDebug:\033[0m " + msg + "\n")
			c.viewport.UpdateView()
			width, height := c.term.Size()
			c.layout.RenderAll(width, height)
		}
	}
}

// Close closes the client
func (c *Client) Close() {
	if !c.running {
		return
	}
	c.running = false

	// Close connection first
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	// Close lua engine
	if c.engine != nil {
		c.engine.Close()
	}

	// Close event system
	if c.events != nil {
		// Signal quit before closing
		c.events.Emit(events.Event{
			Type: events.EventOutput,
			Data: struct {
				Text   string
				Buffer string
			}{"Goodbye!", ""},
		})
	}

	// Reset terminal
	if c.term != nil {
		c.term.Cleanup()
	}

	// Signal all goroutines to stop
	close(c.done)
}

// handleConnect handles connection requests
func (c *Client) handleConnect(e events.Event) {
	// Already connected
	if c.conn != nil {
		return
	}

	// Get connection parameters from event
	params, ok := e.Data.(map[string]interface{})
	if !ok {
		c.buffer.Write("Invalid connect parameters\n")
		c.viewport.UpdateView()
		width, height := c.term.Size()
		c.layout.RenderAll(width, height)
		return
	}

	host, ok := params["host"].(string)
	if !ok {
		c.buffer.Write("Invalid host parameter\n")
		c.viewport.UpdateView()
		width, height := c.term.Size()
		c.layout.RenderAll(width, height)
		return
	}

	portVal, ok := params["port"].(int)
	if !ok {
		c.buffer.Write("Invalid port parameter\n")
		c.viewport.UpdateView()
		width, height := c.term.Size()
		c.layout.RenderAll(width, height)
		return
	}

	c.updateStatus("Connecting...")
	var err error
	c.conn, err = telnet.NewTelnetConnection(host, portVal, c.debug)
	if err != nil {
		c.buffer.Write(fmt.Sprintf("Failed to connect: %v\n", err))
		c.viewport.UpdateView()
		width, height := c.term.Size()
		c.layout.RenderAll(width, height)
		return
	}

	// Store successful connection details
	c.host = host
	c.port = portVal

	// Start reading from telnet
	go c.handleTelnetOutput()

	c.updateStatus("Connected")
	c.events.Emit(events.Event{
		Type: events.EventConnected,
	})
}

// handleDisconnect handles disconnect requests
func (c *Client) handleDisconnect(e events.Event) {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
		c.updateStatus("Disconnected")
		c.events.Emit(events.Event{
			Type: events.EventDisconnected,
		})
	}
}

// handleRawInput handles raw input from the client
func (c *Client) handleRawInput(e events.Event) {
    // Raw input is already emitted in handleInput, nothing to do here
}

// handleQuit handles quit requests
func (c *Client) handleQuit(e events.Event) {
	c.running = false
	c.Close()
	// Exit program
	c.events.Emit(events.Event{
		Type: events.EventOutput,
		Data: struct {
			Text   string
			Buffer string
		}{"Goodbye!", ""},
	})
}

// setupEventHandlers sets up event subscriptions
func (c *Client) setupEventHandlers() {
	c.events.Subscribe(events.EventConnect, c.handleConnect)
	c.events.Subscribe(events.EventDisconnect, c.handleDisconnect)
	c.events.Subscribe(events.EventRawInput, c.handleRawInput)
	c.events.Subscribe(events.EventQuit, c.handleQuit)

	// Subscribe to processed events from LuaEngine
	c.events.Subscribe(events.EventCommand, c.handleCommand)
	c.events.Subscribe(events.EventOutput, c.handleProcessedOutput)
	c.events.Subscribe(events.EventLog, c.handleLog)
	c.events.Subscribe(events.EventDebug, c.handleDebug)
}
