package telnet

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/mmcdole/runes/internal/util"
)

func NewTelnetProxy(log *util.Logger, host string, port string) *TelnetProxy {
	return &TelnetProxy{
		log:        log,
		host:       host,
		port:       port,
		inputChan:  make(chan string),
		outputChan: make(chan string),
	}
}

// TelnetProxy creates a connectable entity that connects to a single
// external telnet server that a runes session can interact with.
// Typically this would be an external game server that supports telnet
// connections.
type TelnetProxy struct {
	inputChan  chan string
	outputChan chan string
	conn       net.Conn
	host       string
	port       string
	log        *util.Logger
}

func (p *TelnetProxy) Connect() error {
	if p.conn != nil {
		p.conn.Close()
	}

	address := fmt.Sprintf("%s:%s", p.host, p.port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	p.conn = conn
	// p.ConnectChan <- true

	go p.readOutput()
	go p.sendInput()

	return nil
}

func (p *TelnetProxy) Close() error {
	return p.conn.Close()
}

func (p *TelnetProxy) Output() chan string {
	return p.outputChan
}

func (p *TelnetProxy) Input() chan string {
	return p.inputChan
}

func (p *TelnetProxy) readOutput() {
	for {
		err := p.readFromConn()
		if err != nil {
			p.conn.Close()
			// p.DisconnectChan <- true
			break
		}
	}
}

func (p *TelnetProxy) readFromConn() error {
	buf := make([]byte, 0)
	for {
		// Read a chunk of data from the connection
		chunk := make([]byte, 1024)
		n, err := p.conn.Read(chunk)
		if err != nil {
			if err == io.EOF {
				// Process the last chunk of data
				p.processLines(buf, chunk[:n])
				break
			} else {
				// Handle the error
				log.Printf("Error reading from connection: %v", err)
				return err
			}
		}
		// Process the lines in the chunk
		p.processLines(buf, chunk[:n])
	}
	return nil
}

func (p *TelnetProxy) processLines(buf []byte, data []byte) {
	// Append the data to the buffer
	buf = append(buf, data...)
	// Find the first newline in the buffer
	i := bytes.IndexByte(buf, '\n')
	for i >= 0 {
		// Output the line
		p.outputChan <- string(buf[:i+1])
		// Remove the processed line from the buffer
		buf = buf[i+1:]
		// Find the next newline in the buffer
		i = bytes.IndexByte(buf, '\n')
	}
	// Check if there is a partial line with no newline
	if len(buf) > 0 {
		p.outputChan <- string(buf)
	}
}

func (p *TelnetProxy) sendInput() {
	for {
		input := <-p.inputChan
		_, err := p.conn.Write([]byte(input))
		if err != nil {
			p.conn.Close()
			// p.DisconnectChan <- true
			break
		}
	}
}
