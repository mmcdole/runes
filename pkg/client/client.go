package client

import (
	"fmt"
	"io"
	"bytes"

	"github.com/mmcdole/runes/pkg/client/buffer"
	"github.com/mmcdole/runes/pkg/client/ui"
	"github.com/mmcdole/runes/pkg/events"
	"github.com/mmcdole/runes/pkg/luaengine"
	"github.com/mmcdole/runes/pkg/protocol/telnet"
)

// Client represents a MUD client
type Client struct {
	host          string
	port          int
	conn          *telnet.TelnetConnection
	ui            *ui.UI
	events        *events.EventProcessor
	bufferMgr     *buffer.BufferManager
	lineProcessor *buffer.LineProcessor
	debug         bool
	engine        *luaengine.LuaEngine
	done          chan struct{}
}

// Config holds the client configuration
type Config struct {
	Host string
	Port int
}

// NewClient creates a new client instance
func NewClient(userScriptDir string, eventProcessor *events.EventProcessor, config Config, debug bool) (*Client, error) {
	client := &Client{
		host:          config.Host,
		port:          config.Port,
		events:        eventProcessor,
		bufferMgr:     buffer.NewBufferManager(),
		lineProcessor: buffer.NewLineProcessor(),
		debug:         debug,
		done:          make(chan struct{}),
	}

	// Create UI
	ui, err := ui.New(client, client.bufferMgr)
	if err != nil {
		return nil, fmt.Errorf("failed to create UI: %w", err)
	}
	client.ui = ui

	engine := luaengine.New(userScriptDir, eventProcessor)
	if err := engine.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize lua engine: %w", err)
	}
	client.engine = engine

	// Set up event handlers
	client.setupEventHandlers()

	return client, nil
}

// Run starts the client and UI
func (c *Client) Run() error {
	// Start UI
	if err := c.ui.Run(); err != nil {
		return fmt.Errorf("failed to start UI: %w", err)
	}

	// Set status to connecting
	c.ui.SetStatus("Connecting...")

	// Connect to server
	var err error
	c.conn, err = telnet.NewTelnetConnection(c.host, c.port, c.debug)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Set status to connected
	c.ui.SetStatus("Connected")

	// Start reading from telnet
	go func() {
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
						c.ui.PrintOutput("Connection closed by server")
						c.Close()
						return
					}
					c.ui.PrintOutput(fmt.Sprintf("Error reading from server: %v", err))
					continue
				}
				line += string(buf[:n])
				if i := bytes.IndexByte(buf[:n], '\n'); i >= 0 {
					c.ProcessServerOutput(line[:i])
					line = line[i+1:]
				}
			}
		}
	}()

	return nil
}

// HandleInput handles user input
func (c *Client) HandleInput(input string) {
	if _, err := c.conn.Write([]byte(input + "\n")); err != nil {
		c.ui.PrintOutput(fmt.Sprintf("Error sending input: %v", err))
	}
}

// HandleQuit handles quit request
func (c *Client) HandleQuit() {
	c.Close()
}

// ProcessServerOutput processes output from the server
func (c *Client) ProcessServerOutput(line string) {
	c.ui.PrintOutput(line)
}

// Close closes the client
func (c *Client) Close() {
	close(c.done)
	if c.conn != nil {
		c.conn.Close()
	}
	if c.ui != nil {
		c.ui.Close()
	}
}

// Wait waits for the client to finish
func (c *Client) Wait() {
	<-c.done
}

func (c *Client) setupEventHandlers() {
	// Add event handlers here
}
