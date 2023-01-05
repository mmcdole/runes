package telnet

import (
	"net"

	"github.com/mmcdole/runes/internal/types"
	"github.com/mmcdole/runes/internal/util"
)

type TelnetServer struct {
	Address      string
	connected    chan types.Connection
	disconnected chan types.Connection
	logger       util.Logger
}

func NewTelnetServer(logger util.Logger, address string, connected chan types.Connection, disconnected chan types.Connection) *TelnetServer {
	return &TelnetServer{
		Address:      address,
		connected:    connected,
		disconnected: disconnected,
		logger:       logger,
	}
}

func (ts *TelnetServer) Run() error {
	ln, err := net.Listen("tcp", ts.Address)
	if err != nil {
		return err
	}

	go ts.acceptConnections(ln)

	ts.logger.Debug("[TelnetServer]: Started server at '%s'", ts.Address)
	return nil
}

func (ts *TelnetServer) acceptConnections(ln net.Listener) {
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}

		ts.logger.Debug("[TelnetServer]: Client Connected: '%s'", conn.RemoteAddr().String())

		// Create telnet connection wrapper struct
		tc := NewTelnetConnection(ts.logger, conn, ts.disconnected)
		// Event externally that a new connection has been produced
		ts.connected <- tc
		// Begin receiving and sending input/output from the telnet connection
		go tc.HandleConnection()
	}
}
