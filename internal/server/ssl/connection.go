package ssl

import (
	"fmt"
	"net"
	"sync"

	"github.com/google/uuid"
	"github.com/mmcdole/runes/internal/types"
	"github.com/mmcdole/runes/internal/util"
)

type SSLConnection struct {
	logger         util.Logger
	inputChan      chan *types.ClientCommand
	inputChanMu    sync.Mutex
	outputChan     chan string
	disconnectChan chan types.Connection
	conn           net.Conn
	id             string
}

func NewSSLConnection(log util.Logger, conn net.Conn, disconnectChan chan types.Connection) *SSLConnection {
	return &SSLConnection{
		logger:         log,
		outputChan:     make(chan string),
		disconnectChan: disconnectChan,
		conn:           conn,
		id:             uuid.New().String(),
		inputChanMu:    sync.Mutex{},
	}
}

func (tc *SSLConnection) ID() string {
	return tc.id
}

func (tc *SSLConnection) Name() string {
	return fmt.Sprintf("ssl:%s", tc.conn.RemoteAddr())
}

func (tc *SSLConnection) InputChan() chan *types.ClientCommand {
	return tc.inputChan
}

func (tc *SSLConnection) SetInputChan(ic chan *types.ClientCommand) {
	tc.inputChan = ic
}

func (tc *SSLConnection) OutputChan() chan string {
	return tc.outputChan
}

func (tc *SSLConnection) DisconnectChan() chan types.Connection {
	return tc.disconnectChan
}

func (tc *SSLConnection) Close() error {
	// TODO: Nate, should I send some 'done' channel a signal to end readinput/sendoutput ?
	err := tc.conn.Close()
	tc.handleDisconnect()
	return err
}

func (tc *SSLConnection) HandleConnection() {
	go tc.readInput()
	go tc.sendOutput()
}

func (tc *SSLConnection) readInput() {
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

func (tc *SSLConnection) sendOutput() {
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

func (tc *SSLConnection) handleDisconnect() {
	tc.logger.Debug("[SSLServer]: Client Disconnected: '%s'", tc.conn.RemoteAddr().String())
	tc.disconnectChan <- tc
}
