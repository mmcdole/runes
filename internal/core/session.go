// End-to-end data flow(s)
//
// Input Steps: Server>Client>Session>PluginEngine>Session>Proxy
//
// 1.  Server: Client Connnection Created
// 2.  Client: Attached to a Session
// 3.  Client: Reads input commands from net.conn (in Go-Routine)
// 4.  Client: Writes these commands to its "InputChan"
// 5.  Session: Reads from all Client Connection InputChan's
// 6.  Session: Sends input to PluginEngine InCommandChan
// 7.  PluginEngine: Read from InCommandChan
// 8.  PluginEngine checks command against aliased lua commands
//        8b. If command is not an alias, forward command to PluginEngine OutCommandChan
//        8a. If command is an alias, execute aliased lua code
// 9.  Session: Read from OutCommandChan and foward to Proxy InputChan
// 10. Proxy: Read from Proxy InputChan and write to net.conn
//
// Output Steps: Proxy>Session>PluginEngine>Session>Server
//
// 1.  Proxy: Connection Created
// 2.  Proxy: Read from net.conn, send lines of output to Proxy OutputChan
// 3.  Session: Read from Proxy "OutputChan" and write to PluginEngine InTextLineChan
// 4.  PluginEngine: Read InTextLineChan for new text lines to process
// 5.  PluginEngine: Checks for Actions/Triggers/Subs/Highlights against the line of text
// 6.  PluginEngine: Send text line to OutTextLineChan, with a buffer set as "default"
// 7.  Session: Read from Plugin OutTextLineChan
// 8.  Session: Write the text line to the appropriate buffer/window
// 9.  Session: Send the text line to any Client's OutputChan for the given buffer/window
// 10. Client: Read from Client Connection's OutputChan and write to client net.conn

package core

import (
	"fmt"
	"strings"

	"github.com/mmcdole/runes/internal/config"
	"github.com/mmcdole/runes/internal/plugin"
	"github.com/mmcdole/runes/internal/proxy"
	"github.com/mmcdole/runes/internal/types"
	"github.com/mmcdole/runes/internal/util"
)

func NewSession(logger util.Logger, conf *config.Config, name string, proxy proxy.ProxyConnection, sm *SessionManager) *Session {
	pe := plugin.NewPluginEngine(logger)
	bm := NewBufferManager(conf.Core.BufferSize)

	session := &Session{
		Name:              name,
		config:            conf,
		proxyConnection:   proxy,
		clientConnections: map[string]types.Connection{},
		sessionManager:    sm,
		bufferManager:     bm,
		pluginEngine:      pe,
		inputChan:         make(chan *types.ConnectionInput),
		log:               logger,
	}

	session.commandHandlers = session.buildCommandHandlers()

	return session
}

type Session struct {
	Name              string
	config            *config.Config
	inputChan         chan *types.ConnectionInput
	sessionManager    *SessionManager
	bufferManager     *BufferManager
	pluginEngine      *plugin.PluginEngine
	proxyConnection   proxy.ProxyConnection
	clientConnections map[string]types.Connection
	commandHandlers   map[string]Command
	log               util.Logger
}

func (s *Session) AttachClient(client types.Connection) {
	s.log.Debug("[Session@%s]: Client: '%s' Attached", s.Name, client.Name())
	s.clientConnections[client.ID()] = client

	s.SwitchClientToBuffer(client, primaryBufferName)

	client.SetInputChan(s.inputChan)
}

func (s *Session) DetachClient(client types.Connection) {
	s.log.Debug("[Session@%s]: Client: '%s' Detached", s.Name, client.Name())
	delete(s.clientConnections, client.ID())
	client.SetInputChan(nil)
}

func (s *Session) SwitchClientToBuffer(client types.Connection, bufferName string) {
	// Assign this client to the provided buffer in manager
	s.bufferManager.SwitchClientToBuffer(client.ID(), bufferName)

	// Send a configured number of "replay history" history lines to the client
	histLines := s.bufferManager.GetLastLines(bufferName, s.config.Core.BufferReplaySize)
	for _, line := range histLines {
		client.OutputChan() <- line
	}
}

func (s *Session) Start() {
	// Start processing plugin input/output
	go s.pluginEngine.Start()

	// Handle input from Clients
	go func() {
		for {
			select {
			// TODO: need to fan-in all the client input
			case input := <-s.inputChan:
				s.handleClientInput(input)
			}
		}
	}()

	// Handle output from PluginEngine
	go func() {
		for {
			select {
			case input := <-s.pluginEngine.OutCommandChan:
				s.handlePluginCommand(input)
			case input := <-s.pluginEngine.OutSendChan:
				// Plugin send() calls have generated new commands to be processed

				// TODO: refactor this to maybe not use ClientInput? Plugin commands
				// aren't truly the same as client input.
				s.handleClientInput(&types.ConnectionInput{Text: input})
			case output := <-s.pluginEngine.OutTextLineChan:
				s.handlePluginOutput(output)
			}
		}
	}()

	// Handle output from the ProxyConnection
	go func() {
		for {
			select {
			case output := <-s.proxyConnection.Output():
				s.handleProxyOutput(output)
			}
		}
	}()

	s.log.Debug("[Session@%s]: Started", s.Name)
}

func (s *Session) handlePluginCommand(command string) {
	s.log.Trace("[Session@%s]: Command In (Plugin): '%s'", s.Name, strings.TrimSpace(command))

	// Foward processed commands from the PluginEngine to the Proxy

	s.log.Trace("[Session@%s]: Command Out (Proxy): '%s'", s.Name, strings.TrimSpace(command))
	s.proxyConnection.Input() <- command
}

func (s *Session) handlePluginOutput(output plugin.BufferOutput) {
	s.log.Trace("[Session@%s]: Text In (Plugin): %s", s.Name, strings.TrimSpace(output.Line))

	s.writeBufferLine(output.BufferName, output.Line)

	s.log.Trace("[Session@%s]: Text Out (Client): %s", s.Name, strings.TrimSpace(output.Line))
}

func (s *Session) handleProxyOutput(output string) {
	s.log.Trace("[Session@%s]: Text In (Proxy): %s", s.Name, strings.TrimSpace(output))

	s.log.Trace("[Session@%s]: Text Out (Plugin): %s", s.Name, strings.TrimSpace(output))
	s.pluginEngine.InTextLineChan <- output
}

func (s *Session) handleClientInput(input *types.ConnectionInput) {
	s.log.Trace("[Session@%s]: Command In (Client): '%s'", s.Name, strings.TrimSpace(input.Text))

	// Check if input is a runes command, otherwise send to plugin engine
	if ok := s.handleCommand(input); !ok {

		s.log.Trace("[Session@%s]: Command Out (Plugin): '%s'", s.Name, strings.TrimSpace(input.Text))
		s.pluginEngine.InCommandChan <- input.Text
	}
}

func (s *Session) writeClientText(client types.Connection, text string) {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		s.writeClientLine(client, line+"\n")
	}
}

func (s *Session) writeClientLine(client types.Connection, line string) {
	client.OutputChan() <- fmt.Sprintf("[r] %s", line)
}

// Write text which may contain multiple lines to the named buffer
// and output to any clients with that buffer assigned.
func (s *Session) writeBufferText(bufferName string, text string) {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		s.writeBufferLine(bufferName, line+"\n")
	}
}

// Write a line to the the named buffer and output to any client with that
// buffer assigned.
func (s *Session) writeBufferLine(bufferName string, line string) {
	// Write new line(s) to the appropriate buffer
	s.bufferManager.AppendLine(bufferName, line)

	// Output new line to any clients with this buffer assigned
	clientIds := s.bufferManager.GetClientsForBuffer(bufferName)
	for _, clientId := range clientIds {
		if conn, ok := s.clientConnections[clientId]; ok {
			conn.OutputChan() <- line
		}
	}
}

// Handle built-in commands otherwise, return false
func (s *Session) handleCommand(input *types.ConnectionInput) bool {
	cmdPrefix := s.config.Core.CommandPrefix

	// Command has configured prefix?
	if !strings.HasPrefix(input.Text, cmdPrefix) {
		return false
	}

	// Trim trailing carriage return
	cmdText := strings.TrimSpace(input.Text)

	parts := strings.Split(cmdText[len(cmdPrefix):], " ")
	command := parts[0]
	args := parts[1:]

	if len(parts) == 0 {
		return false
	}

	params := &CommandParams{
		Command:     command,
		Args:        args,
		FullCommand: input.Text,
		Session:     s,
		Executor:    input.Client,
	}

	if handler, ok := s.commandHandlers[command]; ok {
		return handler.Execute(params)
	}

	return false
}

func (s *Session) buildCommandHandlers() map[string]Command {
	return map[string]Command{
		"session": &SessionCommand{},
		"ping":    &PingCommand{},
		"buffer":  &BufferCommand{},
	}
}
