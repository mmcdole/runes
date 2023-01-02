package client

import (
	"fmt"
	"net"

	"github.com/mmcdole/runes/internal/core"
)

type TelnetConnection struct {
	inputChan      chan string
	outputChan     chan string
	disconnectChan chan bool
	conn           net.Conn
}

func NewTelnetConnection(conn net.Conn) *TelnetConnection {
	return &TelnetConnection{
		inputChan:      make(chan string),
		outputChan:     make(chan string),
		disconnectChan: make(chan bool),
		conn:           conn,
	}
}

func (tc *TelnetConnection) Name() string {
	return fmt.Sprintf("telnet:%s", tc.conn.RemoteAddr())
}

func (tc *TelnetConnection) InputChan() chan string {
	return tc.inputChan
}

func (tc *TelnetConnection) OutputChan() chan string {
	return tc.outputChan
}

func (tc *TelnetConnection) DisconnectChan() chan bool {
	return tc.disconnectChan
}

func (tc *TelnetConnection) Close() error {
	return tc.conn.Close()
}

func (tc *TelnetConnection) HandleConnection() {
	go tc.readInput()
	go tc.sendOutput()
}

func (tc *TelnetConnection) readInput() {
	// Read input commands from telnet connection and write to inputChan
	buf := make([]byte, 1024)
	for {
		n, err := tc.conn.Read(buf)
		if err != nil {
			tc.conn.Close()
			tc.disconnectChan <- true
			break
		}
		tc.inputChan <- string(buf[:n])
	}
}

func (tc *TelnetConnection) sendOutput() {
	// Read text from outputChan and write to the telnet connection
	for {
		select {
		case output := <-tc.outputChan:
			_, err := tc.conn.Write([]byte(output))
			if err != nil {
				tc.conn.Close()
				tc.disconnectChan <- true
				break
			}
		}
	}
}

type TelnetServer struct {
	Address   string
	connected chan core.ClientConnection
}

func NewTelnetServer(address string, connected chan core.ClientConnection) *TelnetServer {
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
