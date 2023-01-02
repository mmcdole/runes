package telnet

import (
	"net"

	"github.com/mmcdole/runes/internal/client"
)

type TelnetServer struct {
	Address   string
	connected chan client.ClientConnection
}

func NewTelnetServer(address string, connected chan client.ClientConnection) *TelnetServer {
	return &TelnetServer{
		Address:   address,
		connected: connected,
	}
}

func (ts *TelnetServer) Run() error {
	ln, err := net.Listen("tcp", ts.Address)
	if err != nil {
		return err
	}

	go ts.acceptConnections(ln)
	return nil
}

func (ts *TelnetServer) acceptConnections(ln net.Listener) {
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}

		// Create telnet connection wrapper struct
		tc := NewTelnetConnection(conn)
		// Event externally that a new connection has been produced
		ts.connected <- tc
		// Begin receiving and sending input/output from the telnet connection
		go tc.HandleConnection()
	}
}
