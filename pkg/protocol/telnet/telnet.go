package telnet

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"time"
)

// Telnet commands (RFC 854)
const (
	cmdIAC  = 255 // Interpret As Command
	cmdDONT = 254
	cmdDO   = 253
	cmdWONT = 252
	cmdWILL = 251
	cmdSB   = 250 // Subnegotiation Begin
	cmdGA   = 249 // Go Ahead
	cmdEL   = 248 // Erase Line
	cmdEC   = 247 // Erase Character
	cmdAYT  = 246 // Are You There
	cmdAO   = 245 // Abort Output
	cmdIP   = 244 // Interrupt Process
	cmdBRK  = 243 // Break
	cmdDM   = 242 // Data Mark
	cmdNOP  = 241 // No Operation
	cmdSE   = 240 // Subnegotiation End
	cmdEOR  = 239 // End of Record
)

// Text formatting characters
const (
	charNUL = 0x00 // Null
	charBEL = 0x07 // Bell
	charBS  = 0x08 // Backspace
	charHT  = 0x09 // Horizontal Tab
	charLF  = 0x0A // Line Feed
	charVT  = 0x0B // Vertical Tab
	charFF  = 0x0C // Form Feed
	charCR  = 0x0D // Carriage Return
)

// Telnet options
const (
	optECHO          = 1
	optSUPPRESS_GA   = 3
	optSTATUS        = 5
	optTIMING_MARK   = 6
	optTERMINAL_TYPE = 24
	optWINDOW_SIZE   = 31
	optTERM_SPEED    = 32
	optLINEMODE      = 34
	optNEW_ENVIRON   = 39
	optMSDP          = 69  // MUD Server Data Protocol
	optMSSP          = 70  // MUD Server Status Protocol
	optMCCP2         = 86  // MUD Client Compression Protocol v2
	optMCCP3         = 87  // MUD Client Compression Protocol v3
	optMSP           = 90  // MUD Sound Protocol
	optMXP           = 91  // MUD Extension Protocol
	optGMCP          = 201 // Generic MUD Communication Protocol
)

// shouldFilter returns true if the byte should be filtered from output
func shouldFilter(b byte) bool {
	// Only filter out truly unwanted control characters
	// Keep: Bell, BS, Tab, LF, VT, FF, CR, ESC
	return b < 0x20 && b != charBEL && b != charBS && b != charHT &&
		b != charLF && b != charVT && b != charFF && b != charCR &&
		b != 0x1B // ESC for ANSI
}

// OptionState represents the state of a telnet option
type OptionState struct {
	Supported     bool
	LocalEnabled  bool
	RemoteEnabled bool
}

// TelnetConnection implements the Connection interface for telnet connections
type TelnetConnection struct {
	host  string
	port  int
	conn  net.Conn
	debug bool

	// Buffers for processing telnet protocol
	cmdBuffer []byte
	inCommand bool
	inSubneg  bool
	options   map[byte]OptionState
}

// NewTelnetConnection creates a new telnet connection
func NewTelnetConnection(host string, port int, debug bool) (*TelnetConnection, error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)))
	if err != nil {
		return nil, err
	}

	t := &TelnetConnection{
		host:      host,
		port:      port,
		conn:      conn,
		debug:     debug,
		cmdBuffer: make([]byte, 0, 3),
		options:   make(map[byte]OptionState),
	}

	// Set up supported options
	t.options[optSUPPRESS_GA] = OptionState{Supported: true}
	t.options[optMCCP2] = OptionState{Supported: true}
	t.options[optGMCP] = OptionState{Supported: true}

	return t, nil
}

func (t *TelnetConnection) Read(p []byte) (int, error) {
	n, err := t.conn.Read(p)
	if err != nil {
		return n, err
	}

	// Log raw bytes to file
	if f, err := os.OpenFile("/tmp/telnet.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		defer f.Close()
		fmt.Fprintf(f, "\n=== Read %d bytes at %s ===\n", n, time.Now().Format(time.RFC3339))
		fmt.Fprintf(f, "Raw bytes: %v\n", p[:n])
		fmt.Fprintf(f, "Hex dump:\n%s\n", hexDump(p[:n]))
	}

	output := make([]byte, len(p)*2)
	outIndex := 0
	dataStart := 0

	for i := 0; i < n; i++ {
		b := p[i]

		// Handle telnet commands
		if b == cmdIAC {
			// Copy any pending data before handling the command
			if dataStart < i {
				copy(output[outIndex:], p[dataStart:i])
				outIndex += i - dataStart
				if t.debug {
					fmt.Printf("Data received: \n%s", hexDump(p[dataStart:i]))
				}
			}
			dataStart = i + 1

			t.inCommand = true
			t.cmdBuffer = t.cmdBuffer[:0]
			t.cmdBuffer = append(t.cmdBuffer, b)
			continue
		}

		if t.inCommand {
			t.cmdBuffer = append(t.cmdBuffer, b)

			// Handle different command types
			switch len(t.cmdBuffer) {
			case 2:
				// Two-byte commands
				switch t.cmdBuffer[1] {
				case cmdSB:
					t.inSubneg = true
					continue
				case cmdIAC:
					// Escaped IAC byte
					output[outIndex] = cmdIAC
					outIndex++
					t.inCommand = false
					dataStart = i + 1
				case cmdGA, cmdEL, cmdEC, cmdAYT, cmdAO, cmdIP, cmdBRK, cmdDM, cmdNOP:
					// Simple commands
					t.inCommand = false
					dataStart = i + 1
				}
			case 3:
				// Three-byte negotiation commands
				events := t.handleCommand(t.cmdBuffer)
				if t.debug && len(events) > 0 {
					fmt.Printf("Command events: %v\n", events)
				}
				t.inCommand = false
				dataStart = i + 1
			}
			continue
		}

		if t.inSubneg {
			t.cmdBuffer = append(t.cmdBuffer, b)
			if b == cmdSE && len(t.cmdBuffer) >= 2 && t.cmdBuffer[len(t.cmdBuffer)-2] == cmdIAC {
				events := t.handleSubnegotiation(t.cmdBuffer)
				if t.debug && len(events) > 0 {
					fmt.Printf("Subnegotiation events: %v\n", events)
				}
				t.cmdBuffer = t.cmdBuffer[:0]
				t.inSubneg = false
				dataStart = i + 1
			}
			continue
		}

		// Filter unwanted control characters from regular data
		if shouldFilter(b) {
			if dataStart == i {
				dataStart = i + 1
			} else {
				// Copy data up to this point
				copy(output[outIndex:], p[dataStart:i])
				outIndex += i - dataStart
				dataStart = i + 1
			}
			continue
		}
	}

	// Copy any remaining data
	if dataStart < n {
		copy(output[outIndex:], p[dataStart:n])
		outIndex += n - dataStart
		if t.debug {
			fmt.Printf("Data received: \n%s", hexDump(p[dataStart:n]))
		}
	}

	copy(p, output[:outIndex])
	return outIndex, nil
}

func (t *TelnetConnection) Write(p []byte) (n int, err error) {
	// Escape any IAC bytes in the data
	var escaped []byte
	for _, b := range p {
		if b == cmdIAC {
			escaped = append(escaped, cmdIAC, cmdIAC)
		} else {
			escaped = append(escaped, b)
		}
	}
	return t.conn.Write(escaped)
}

func (t *TelnetConnection) Close() error {
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}

// hexDump returns a hex+ASCII dump of data for debugging
func hexDump(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	var buf bytes.Buffer
	for i := 0; i < len(data); i += 16 {
		// Print hex values
		fmt.Fprintf(&buf, "%04x: ", i)
		for j := 0; j < 16; j++ {
			if i+j < len(data) {
				fmt.Fprintf(&buf, "%02x ", data[i+j])
			} else {
				buf.WriteString("   ")
			}
		}
		buf.WriteString("|")

		// Print ASCII representation
		for j := 0; j < 16 && i+j < len(data); j++ {
			c := data[i+j]
			if c >= 32 && c <= 126 {
				buf.WriteByte(c)
			} else {
				buf.WriteByte('.')
			}
		}
		buf.WriteString("|\n")
	}
	return buf.String()
}

type TelnetEvent interface{}

type NegotiationEvent struct {
	Command byte
	Option  byte
}

type SubnegotiationEvent struct {
	Option byte
	Data   []byte
}

func (t *TelnetConnection) handleCommand(cmd []byte) []TelnetEvent {
	if len(cmd) != 3 {
		return nil
	}

	var events []TelnetEvent
	switch cmd[1] {
	case cmdWILL:
		events = append(events, NegotiationEvent{Command: cmdWILL, Option: cmd[2]})
		if state, ok := t.options[cmd[2]]; ok && state.Supported {
			t.conn.Write([]byte{cmdIAC, cmdDO, cmd[2]})
			state.RemoteEnabled = true
			t.options[cmd[2]] = state
		} else {
			t.conn.Write([]byte{cmdIAC, cmdDONT, cmd[2]})
		}
	case cmdWONT:
		events = append(events, NegotiationEvent{Command: cmdWONT, Option: cmd[2]})
		if state, ok := t.options[cmd[2]]; ok {
			state.RemoteEnabled = false
			t.options[cmd[2]] = state
		}
		t.conn.Write([]byte{cmdIAC, cmdDONT, cmd[2]})
	case cmdDO:
		events = append(events, NegotiationEvent{Command: cmdDO, Option: cmd[2]})
		if state, ok := t.options[cmd[2]]; ok && state.Supported {
			t.conn.Write([]byte{cmdIAC, cmdWILL, cmd[2]})
			state.LocalEnabled = true
			t.options[cmd[2]] = state
		} else {
			t.conn.Write([]byte{cmdIAC, cmdWONT, cmd[2]})
		}
	case cmdDONT:
		events = append(events, NegotiationEvent{Command: cmdDONT, Option: cmd[2]})
		if state, ok := t.options[cmd[2]]; ok {
			state.LocalEnabled = false
			t.options[cmd[2]] = state
		}
		t.conn.Write([]byte{cmdIAC, cmdWONT, cmd[2]})
	}
	return events
}

func (t *TelnetConnection) handleSubnegotiation(cmd []byte) []TelnetEvent {
	if len(cmd) < 4 {
		return nil
	}

	var events []TelnetEvent
	option := cmd[2]
	data := cmd[3 : len(cmd)-2] // Remove IAC SE at the end
	events = append(events, SubnegotiationEvent{Option: option, Data: data})
	return events
}
