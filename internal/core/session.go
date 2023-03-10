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

	"github.com/fatih/color"

	"github.com/mmcdole/runes/internal/config"
	"github.com/mmcdole/runes/internal/plugin"
	"github.com/mmcdole/runes/internal/proxy"
	"github.com/mmcdole/runes/internal/types"
	"github.com/mmcdole/runes/internal/util"
)

func NewSession(logger util.Logger, conf *config.Config, name string, proxy proxy.ProxyConnection, sm *SessionManager) *Session {
	pe := plugin.NewPluginEngine(logger, conf)
	bm := NewBufferManager(conf.Core.BufferSize)

	session := &Session{
		Name:              name,
		config:            conf,
		proxyConnection:   proxy,
		clientConnections: map[string]types.Connection{},
		sessionManager:    sm,
		bufferManager:     bm,
		pluginEngine:      pe,
		inputChan:         make(chan *types.ClientCommand),
		log:               logger,
	}

	session.commandHandlers = session.buildCommandHandlers()

	return session
}

type Session struct {
	Name              string
	config            *config.Config
	inputChan         chan *types.ClientCommand
	sessionManager    *SessionManager
	bufferManager     *BufferManager
	pluginEngine      *plugin.PluginEngine
	proxyConnection   proxy.ProxyConnection
	clientConnections map[string]types.Connection
	commandHandlers   map[string]Command
	log               util.Logger
}

func (s *Session) AttachClient(client types.Connection) {
	s.log.Debug("[Session]: Client: '%s' Attached", client.Name())
	s.clientConnections[client.ID()] = client

	s.SwitchClientToBuffer(client, primaryBufferName)

	client.SetInputChan(s.inputChan)
}

func (s *Session) DetachClient(client types.Connection) {
	s.log.Debug("[Session]: Client: '%s' Detached", client.Name())
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
	s.pluginEngine.Start()

	// Handle input from Clients
	go func() {
		for {
			select {
			case input := <-s.inputChan:
				s.handleClientInput(input)
			}
		}
	}()

	// Handle output from PluginEngine
	go func() {
		for {
			select {
			case input := <-s.pluginEngine.ReceiveProcessedCommands():
				s.handlePluginProcessedCmd(input)
			case input := <-s.pluginEngine.ReceiveCommands():
				s.handlePluginCmd(input)
			case output := <-s.pluginEngine.ReceiveTextLines():
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

	s.log.Debug("[Session]: Started")
}

func (s *Session) handlePluginProcessedCmd(command string) {
	s.log.Trace("[Session]: [Plugin->Session] Command: %s", strings.TrimSpace(command))

	// Foward processed commands from the PluginEngine to the Proxy

	s.log.Trace("[Session]: [Session->Proxy] Command: %s", strings.TrimSpace(command))
	s.proxyConnection.Input() <- command
}

func (s *Session) handlePluginOutput(output types.BufferOutput) {
	s.log.Trace("[Session]: [Plugin->Session] Text: %s", strings.TrimSpace(output.Line))

	s.writeBufferLine(output.BufferName, output.Line)

	s.log.Trace("[Session]: [Session->Client] Text: %s", strings.TrimSpace(output.Line))
}

func (s *Session) handleProxyOutput(output string) {
	s.log.Trace("[Session]: [Proxy->Session]: Text: %s", strings.TrimSpace(output))

	s.log.Trace("[Session]: [Session->Plugin] Text: %s", strings.TrimSpace(output))
	s.pluginEngine.EnqueueText(output)
}

func (s *Session) handleClientInput(input *types.ClientCommand) {
	s.log.Trace("[Session]: [Client->Session] Command: %s", strings.TrimSpace(input.Text))
	s.handleInput(input)
}

func (s *Session) handlePluginCmd(input string) {
	// Plugin send() calls have generated new commands to be processed
	s.log.Trace("[Session]: [Plugin->Session] Command (Send): %s", strings.TrimSpace(input))

	// TODO: refactor this to maybe not use ClientInput? Plugin commands
	// aren't truly the same as client input.
	s.handleInput(&types.ClientCommand{Text: input})
}

func (s *Session) handleInput(input *types.ClientCommand) {
	// check if input command has newline
	if strings.HasSuffix(input.Text, "\n") {
		// Split commands that have multiple commands separated by separator
		commands := strings.Split(input.Text, s.config.Core.CommandSeparator)
		for _, command := range commands {
			// Trim leading and trailing whitespace from the command
			cmd := strings.TrimSpace(command)
			// add newline to the end of command
			cmd += "\n"
			// Wrap the command and use the same client as the parent command
			cc := &types.ClientCommand{
				Text:   cmd,
				Client: input.Client,
			}
			// Check if the command is a runes command, otherwise send to plugin engine
			if ok := s.handleSessionCmd(cc); !ok {
				s.log.Trace("[Session]: [Session->Plugin] Command: %s", strings.TrimSpace(cmd))
				s.pluginEngine.EnqueueCommand(cmd)
			}
		}
	} else {
		// Input was not a command ending in newline, pass through input to server
		s.log.Trace("[Session]: [Session->Proxy] Command (Passthrough): %s", input.Text)
		s.proxyConnection.Input() <- input.Text
	}
}

func (s *Session) writeClientText(client types.Connection, text string) {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		s.writeClientLine(client, line+"\n")
	}
}

func (s *Session) writeClientLine(client types.Connection, line string) {
	white := color.New(color.FgWhite, color.Bold).SprintFunc()
	green := color.New(color.FgHiGreen, color.Bold).SprintFunc()
	text := fmt.Sprintf("%s%s%s %s", white("["), green("r"), white("]"), line)
	client.OutputChan() <- text
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
func (s *Session) handleSessionCmd(input *types.ClientCommand) bool {
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
		"plugin":  &PluginCommand{},
	}
}
