package telnet

import (
	"fmt"
	"net"

	"github.com/google/uuid"
	"github.com/mmcdole/runes/internal/types"
	"github.com/mmcdole/runes/internal/util"
)

type TelnetConnection struct {
	logger         util.Logger
	inputChan      chan *types.ClientCommand
	outputChan     chan string
	disconnectChan chan types.Connection
	conn           net.Conn
	id             string
}

func NewTelnetConnection(log util.Logger, conn net.Conn, disconnectChan chan types.Connection) *TelnetConnection {
	return &TelnetConnection{
		logger:         log,
		outputChan:     make(chan string),
		disconnectChan: disconnectChan,
		conn:           conn,
		id:             uuid.New().String(),
	}
}

func (tc *TelnetConnection) ID() string {
	return tc.id
}

func (tc *TelnetConnection) Name() string {
	return fmt.Sprintf("telnet:%s", tc.conn.RemoteAddr())
}

func (tc *TelnetConnection) InputChan() chan *types.ClientCommand {
	return tc.inputChan
}

func (tc *TelnetConnection) SetInputChan(ic chan *types.ClientCommand) {
	tc.inputChan = ic
}

func (tc *TelnetConnection) OutputChan() chan string {
	return tc.outputChan
}

func (tc *TelnetConnection) DisconnectChan() chan types.Connection {
	return tc.disconnectChan
}

func (tc *TelnetConnection) Close() error {
	// TODO: Nate, should I send some 'done' channel a signal to end readinput/sendoutput ?
	err := tc.conn.Close()
	tc.handleDisconnect()
	return err
}

func (tc *TelnetConnection) HandleConnection() {
	go tc.readInput()
	go tc.sendOutput()
}

func (tc *TelnetConnection) readInput() {
	// Read input commands from telnet connection and write to inputChan
	buf := make([]byte, 4096)
	for {
		n, err := tc.conn.Read(buf)
		if err != nil {
			tc.conn.Close()
			tc.handleDisconnect()
			break
		}
		// TODO: Nate! mutex on inputChan access?
		if tc.inputChan != nil {
			ci := types.ClientCommand{
				Text:   string(buf[:n]),
				Client: tc,
			}

			tc.inputChan <- &ci
		}
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
				tc.handleDisconnect()
				break
			}
		}
	}
}

func (tc *TelnetConnection) handleDisconnect() {
	tc.logger.Debug("[TelnetServer]: Client Disconnected: '%s'", tc.conn.RemoteAddr().String())
	tc.disconnectChan <- tc
}
