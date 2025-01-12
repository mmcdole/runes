package client

import (
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmcdole/runes/pkg/events"
	"github.com/mmcdole/runes/pkg/luaengine"
	"github.com/mmcdole/runes/pkg/protocol/telnet"
)

// Client handles the core MUD client functionality
type Client struct {
	conn          Connection
	engine        *luaengine.LuaEngine
	events        *events.EventProcessor
	bufferMgr     *BufferManager
	lineProcessor *LineProcessor
	connected     bool
	debug         bool
	program       *tea.Program
	model         *model
}

// NewClient creates a new MUD client
func NewClient(eventProcessor *events.EventProcessor, userScriptDir string, debug bool) (*Client, error) {
	// Initialization order is critical:
	// Event handlers must be set up before Lua engine initialization
	// to capture all events emitted during core script loading

	client := &Client{
		events:        eventProcessor,
		bufferMgr:     NewBufferManager(),
		lineProcessor: NewLineProcessor(),
		debug:         debug,
	}

	// Initialize TUI first
	model := NewModel(client, client.bufferMgr)
	client.model = model
	client.program = tea.NewProgram(model, tea.WithAltScreen())

	client.setupEventHandlers()

	engine := luaengine.New(userScriptDir, eventProcessor)
	if err := engine.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize lua engine: %v", err)
	}
	client.engine = engine

	return client, nil
}

func (c *Client) setupEventHandlers() {
	c.events.Subscribe(events.EventConnect, c.handleConnect)
	c.events.Subscribe(events.EventDisconnect, c.handleDisconnect)
	c.events.Subscribe(events.EventCommand, c.handleCommand)
	c.events.Subscribe(events.EventOutput, c.handleOutput)
	c.events.Subscribe(events.EventQuit, c.handleQuit)
}

func (c *Client) handleConnect(e events.Event) {
	data, ok := e.Data.(struct {
		Host string
		Port int
	})
	if !ok {
		return
	}

	if err := c.Connect(data.Host, data.Port); err == nil {
		// Only emit Connected event on success
		c.events.Emit(events.Event{
			Type: events.EventConnected,
			Data: data,
		})
	}
}

func (c *Client) handleDisconnect(e events.Event) {
	if err := c.Disconnect(); err == nil {
		c.events.Emit(events.Event{
			Type: events.EventDisconnected,
		})
	}
}

func (c *Client) handleCommand(e events.Event) {
	if cmd, ok := e.Data.(string); ok {
		c.SendCommand(cmd)
	}
}

func (c *Client) handleOutput(e events.Event) {
	data, ok := e.Data.(struct {
		Text   string
		Buffer string
	})
	if !ok {
		return
	}
	c.bufferMgr.AddLine(data.Text, data.Buffer)
	c.model.UpdateContent()
}

func (c *Client) handleQuit(e events.Event) {
	c.Close()
	os.Exit(0)
}

// IsConnected returns true if the client is connected
func (c *Client) IsConnected() bool {
	return c.connected
}

// Connect connects to a MUD server
func (c *Client) Connect(host string, port int) error {
	if c.conn != nil {
		c.conn.Close()
	}

	telnetConn, err := telnet.NewTelnetConnection(host, port, c.debug)
	if err != nil {
		c.events.Emit(events.Event{
			Type: events.EventRawOutput,
			Data: fmt.Sprintf("Failed to connect: %v", err),
		})
		return err
	}

	c.conn = telnetConn
	c.connected = true

	// Start reading from connection
	go c.readLoop()

	return nil
}

// Disconnect closes the connection to the MUD server
func (c *Client) Disconnect() error {
	if !c.connected {
		return nil
	}
	c.connected = false
	return c.conn.Close()
}

// SendCommand sends a command to the MUD server
func (c *Client) SendCommand(cmd string) error {
	if !c.connected {
		return fmt.Errorf("not connected")
	}
	_, err := c.conn.Write([]byte(cmd + "\n"))
	return err
}

// Send sends data to the server
func (c *Client) Send(data string) {
	if !c.connected {
		return
	}
	c.conn.Write([]byte(data + "\n"))
}

func (c *Client) readLoop() {
	buf := make([]byte, 4096)
	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Read error: %v\n", err)
			}
			c.connected = false
			c.events.Emit(events.Event{
				Type: events.EventDisconnected,
			})
			return
		}

		if n > 0 {
			// Process the raw data into lines
			lines := c.lineProcessor.Write(buf[:n])
			for _, line := range lines {
				c.events.Emit(events.Event{
					Type: events.EventRawOutput,
					Data: line,
				})
			}
		}
	}
}

// New method to handle input from the TUI
func (c *Client) HandleInput(input string) {
	c.events.Emit(events.Event{
		Type: events.EventRawInput,
		Data: input,
	})
}

// Replace inputLoop with TUI program
func (c *Client) Run() error {
	if err := c.program.Start(); err != nil {
		return fmt.Errorf("error running program: %v", err)
	}
	return nil
}

// Close closes the client connection
func (c *Client) Close() {
	if c.connected {
		c.Disconnect()
	}
}
