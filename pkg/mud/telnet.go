package mud

import (
	"fmt"
	"net"
)

type TelnetConnection struct {
	conn net.Conn
	host string
	port int
}

func NewTelnetConnection() *TelnetConnection {
	return &TelnetConnection{}
}

func (t *TelnetConnection) Connect(host string, port int) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("telnet connection failed: %w", err)
	}
	t.conn = conn
	t.host = host
	t.port = port
	return nil
}

func (t *TelnetConnection) Close() error {
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}

func (t *TelnetConnection) Read(p []byte) (n int, err error) {
	if t.conn == nil {
		return 0, fmt.Errorf("not connected")
	}
	return t.conn.Read(p)
}

func (t *TelnetConnection) Write(p []byte) (n int, err error) {
	if t.conn == nil {
		return 0, fmt.Errorf("not connected")
	}
	return t.conn.Write(p)
}
