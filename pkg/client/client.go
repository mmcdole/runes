package client

import (
	"fmt"
	"io"
	"os"

	"github.com/mmcdole/runes/pkg/client/buffer"
	"github.com/mmcdole/runes/pkg/client/tui"
	"github.com/mmcdole/runes/pkg/events"
	"github.com/mmcdole/runes/pkg/luaengine"
	"github.com/mmcdole/runes/pkg/protocol/telnet"
)

type Client struct {
	conn          *telnet.TelnetConnection
	tui           *tui.TUI
	engine        *luaengine.LuaEngine
	events        *events.EventProcessor
	bufferMgr     *buffer.BufferManager
	lineProcessor *buffer.LineProcessor
	connected     bool
	debug         bool
}

func NewClient(userScriptDir string, eventProcessor *events.EventProcessor, debug bool) (*Client, error) {
	config := tui.Config{
		BufferSize: 1000,
	}

	client := &Client{
		events:        eventProcessor,
		bufferMgr:     buffer.NewBufferManager(),
		lineProcessor: buffer.NewLineProcessor(),
		debug:         debug,
	}

	client.tui = tui.New(config, client)
	
	engine := luaengine.New(userScriptDir, eventProcessor)
	if err := engine.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize lua engine: %v", err)
	}
	client.engine = engine

	// Set up event handlers
	client.setupEventHandlers()

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
	c.bufferMgr.AddLine(data.Text)
	c.tui.AddLine(data.Text)
}

func (c *Client) handleQuit(e events.Event) {
	c.Disconnect()
	os.Exit(0)
}

func (c *Client) Connect(host string, port int) error {
	if c.connected {
		return fmt.Errorf("already connected")
	}

	c.bufferMgr.AddLine(fmt.Sprintf("Connecting to %s:%d...", host, port))

	conn, err := telnet.NewTelnetConnection(host, port, c.debug)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}

	c.conn = conn
	c.connected = true

	// Start processing incoming data
	go c.processIncoming()

	return nil
}

func (c *Client) processIncoming() {
	buf := make([]byte, 4096)
	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "Error reading from connection: %v\n", err)
			}
			c.connected = false
			c.events.Emit(events.Event{
				Type: events.EventDisconnected,
			})
			break
		}

		if n > 0 {
			// Process telnet protocol
			c.lineProcessor.Write(buf[:n])
			
			// Send raw output to lua engine for processing
			c.events.Emit(events.Event{
				Type: events.EventRawOutput,
				Data: string(buf[:n]),
			})
		}
	}
}

func (c *Client) HandleInput(input string) {
	// All input goes through the event system
	c.events.Emit(events.Event{
		Type: events.EventRawInput,
		Data: input,
	})
}

func (c *Client) SendCommand(cmd string) error {
	if !c.connected {
		return fmt.Errorf("not connected")
	}
	_, err := c.conn.Write([]byte(cmd + "\n"))
	return err
}

func (c *Client) Run() error {
	return c.tui.Run()
}

func (c *Client) ProcessServerOutput(line string) {
	c.tui.AddLine(line)
}

func (c *Client) Disconnect() error {
	if !c.connected {
		return nil
	}

	err := c.conn.Close()
	c.connected = false
	c.conn = nil
	return err
}

func (c *Client) IsConnected() bool {
	return c.connected
}

func (c *Client) HandleQuit() {
	c.events.Emit(events.Event{
		Type: events.EventQuit,
	})
}

// Close performs cleanup and shuts down the client
func (c *Client) Close() {
	if c.connected {
		c.Disconnect()
	}
	if c.engine != nil {
		c.engine.Close()
	}
}
