package mud

import (
	"fmt"
	"io"
	"os"
)

type Client struct {
	conn      Connection
	engine    Engine
	display   *Display
	connected bool
	cmdQueue  chan string
}

type Engine interface {
	EmitEvent(name string, data interface{})
}

func NewClient(engine Engine) *Client {
	client := &Client{
		conn:    NewTelnetConnection(),
		engine:  engine,
		display: NewDisplay(os.Stdout),
	}

	return client
}

func (c *Client) SendCommand(cmd string) {
	if !c.connected {
		c.engine.EmitEvent("error", "Not connected")
		return
	}
	if _, err := c.conn.Write([]byte(cmd + "\n")); err != nil {
		c.engine.EmitEvent("error", fmt.Sprintf("failed to send command: %v", err))
	}
}

func (c *Client) Connect(host string, port int) error {
	if err := c.conn.Connect(host, port); err != nil {
		return err
	}
	c.connected = true
	go c.readLoop()
	c.engine.EmitEvent("connect", "")
	return nil
}

func (c *Client) Disconnect() error {
	if !c.connected {
		return nil
	}
	c.connected = false
	err := c.conn.Close()
	c.engine.EmitEvent("disconnect", "")
	return err
}

func (c *Client) readLoop() {
	buf := make([]byte, 4096)
	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				c.engine.EmitEvent("error", err.Error())
			}
			c.connected = false
			c.engine.EmitEvent("disconnect", "")
			return
		}
		if n > 0 {
			c.engine.EmitEvent("output", string(buf[:n]))
		}
	}
}

func (c *Client) HandleInput(input string) {
	c.engine.EmitEvent("input", input)
}
