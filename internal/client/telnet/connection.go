package telnet

import (
	"fmt"
	"net"

	"github.com/mmcdole/runes/internal/client"
)

type TelnetConnection struct {
	inputChan      chan client.ClientInput
	outputChan     chan string
	disconnectChan chan bool
	conn           net.Conn
}

func NewTelnetConnection(conn net.Conn) *TelnetConnection {
	return &TelnetConnection{
		outputChan:     make(chan string),
		disconnectChan: make(chan bool),
		conn:           conn,
	}
}

func (tc *TelnetConnection) Name() string {
	return fmt.Sprintf("telnet:%s", tc.conn.RemoteAddr())
}

func (tc *TelnetConnection) InputChan() chan client.ClientInput {
	return tc.inputChan
}

func (tc *TelnetConnection) SetInputChan(ic chan client.ClientInput) {
	tc.inputChan = ic
}

func (tc *TelnetConnection) OutputChan() chan string {
	return tc.outputChan
}

func (tc *TelnetConnection) DisconnectChan() chan bool {
	return tc.disconnectChan
}

func (tc *TelnetConnection) Close() error {
	// TODO: Nate, should I send some 'done' channel a signal to end readinput/sendoutput ?
	return tc.conn.Close()
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
			tc.disconnectChan <- true
			break
		}
		// TODO: Nate! mutex on inputChan access?
		if tc.inputChan != nil {
			tc.inputChan <- client.ClientInput{
				Text:   string(buf[:n]),
				Client: tc,
			}
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
				tc.disconnectChan <- true
				break
			}
		}
	}
}
