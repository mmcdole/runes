package ssl

import (
	"net"

	"github.com/mmcdole/runes/internal/types"
	"github.com/mmcdole/runes/internal/util"
)

type SSLServer struct {
	Address      string
	connected    chan types.Connection
	disconnected chan types.Connection
	logger       util.Logger
}

func NewSSLServer(logger util.Logger, address string, connected chan types.Connection, disconnected chan types.Connection) *SSLServer {
	return &SSLServer{
		Address:      address,
		connected:    connected,
		disconnected: disconnected,
		logger:       logger,
	}
}

func (ts *SSLServer) Run() error {
	ln, err := net.Listen("tcp", ts.Address)
	if err != nil {
		return err
	}

	go ts.acceptConnections(ln)

	ts.logger.Debug("[SSLServer]: Started server at '%s'", ts.Address)
	return nil
}

func (ts *SSLServer) acceptConnections(ln net.Listener) {
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}

		ts.logger.Debug("[SSLServer]: Client Connected: '%s'", conn.RemoteAddr().String())

		// Create telnet connection wrapper struct
		tc := NewSSLConnection(ts.logger, conn, ts.disconnected)
		// Event externally that a new connection has been produced
		ts.connected <- tc
		// Begin receiving and sending input/output from the telnet connection
		go tc.HandleConnection()
	}
}
