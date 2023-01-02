// End-to-end data flow(s)
//
// Input Steps: Client>Session>PluginEngine>Session>Server

// 1.  ClientConnection: Connnection Created
// 4.  ClientConnection: Attached to a Session
// 2.  ClientConnection: Reads input commands from net.conn (in Go-Routine)
// 3.  ClientConnection: Writes these commands to its "InputChan"
// 5.  Session: Reads from all ClientConnection "InputChan"s
// 6.  Session: Sends input to Session's PluginEngine InCommandChan
// 7.  PluginEngine: Read from InCommandChan
// 8.  PluginEngine checks command against aliased lua commands
//        8b. If command is not an alias, forward command to PluginEngine OutCommandChan
//        8a. If command is an alias, execute aliased lua code
// 9.  Session: Read from OutCommandChan and foward to ServerConnection InputChan
// 10. ServerConnection: Read from ServerConnection InputChan and write to server net.conn

// Output Steps: Server>Session>PluginEngine>Session>Client

// 1. ServerConnection: Connection Created
// 2. ServerConnection: Read from net.conn/whatever, send lines of output to OutputChan
// 3. Session: Read from ServerConnection "OutputChan" and write to PluginEngine InTextLineChan
// 4. PluginEngine: Read InTextLineChan for new text lines to process
// 5. PluginEngine: Checks for Actions/Triggers/Subs/Highlights against the line of text
// 6. PluginEngine: Send text line to OutTextLineChan, with a buffer set as "default"
// 7. Session: Read from OutTextLineChan
// 8. Session: Write the text line to the appropriate buffer/window
// 9. Session: Send the text line to any ClientConnection OutputChan for the given buffer/window
// 10 ClientConnection: Read from ClientConnection OutputChan and write to client net.conn

package core

import (
	"fmt"
	"strings"

	"github.com/mmcdole/runes/internal/client"
	"github.com/mmcdole/runes/internal/plugin"
	"github.com/mmcdole/runes/internal/server"
	"github.com/mmcdole/runes/internal/util"
)

func NewSession(logger util.Logger, name string, server server.ServerConnection, sm *SessionManager) *Session {
	pe := plugin.NewPluginEngine(logger)

	return &Session{
		Name:           name,
		server:         server,
		sessionManager: sm,
		pluginEngine:   pe,
		inputChan:      make(chan client.ClientInput),
		clients:        []client.ClientConnection{},
		clientBuffers:  map[*client.ClientConnection]string{},
		buffers:        map[string][]string{},
		logger:         logger,
	}
}

type Session struct {
	Name           string
	inputChan      chan client.ClientInput
	sessionManager *SessionManager
	pluginEngine   *plugin.PluginEngine
	server         server.ServerConnection
	clients        []client.ClientConnection
	clientBuffers  map[*client.ClientConnection]string
	buffers        map[string][]string
	logger         util.Logger
}

func (s *Session) AttachClient(client client.ClientConnection) {
	s.logger.Debug("Session: '%s' Client: '%s':  Attached", s.Name, client.Name())
	s.clients = append(s.clients, client)
	client.SetInputChan(s.inputChan)
}

func (s *Session) DetachClient(client client.ClientConnection) {
	s.logger.Debug("Session: '%s' Client: '%s': Detached", s.Name, client.Name())
	// Remove the connection from the connections slice
	for i, c := range s.clients {
		if c == client {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			break
		}
	}
	client.SetInputChan(nil)
}

func (s *Session) SwitchToSession(client client.ClientConnection, sessionName string) (*Session, error) {
	// Detach the client connection from the current session
	s.DetachClient(client)

	// Get the new session and attach the client connection to it
	session := s.sessionManager.GetSession(sessionName)
	if session == nil {
		return nil, fmt.Errorf("Switch session failed: target session '%s' not found.", sessionName)
	}
	session.AttachClient(client)
	return session, nil
}

func (s *Session) Start() {
	// Start processing plugin input/output
	go s.pluginEngine.Start()

	// Handle input from ClientConnections
	go func() {
		for {
			select {
			// TODO: need to fan-in all the client input
			case input := <-s.inputChan:
				s.handleClientInput(input.Text)
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
				s.handleClientInput(input)
			case output := <-s.pluginEngine.OutTextLineChan:
				s.handlePluginOutput(output)
			}
		}
	}()

	// Handle output from the ServerConnection
	go func() {
		for {
			select {
			case output := <-s.server.Output():
				s.handleServerOutput(output)
			}
		}
	}()

	s.logger.Debug("Session: Started '%s'", s.Name)
}

func (s *Session) handlePluginCommand(command string) {
	s.logger.Trace("Session: '%s': Plugin command: '%s'", s.Name, strings.TrimSpace(command))
	// Foward processed commands from the PluginEngine to the Server
	s.server.Input() <- command
}

func (s *Session) handlePluginOutput(output plugin.BufferOutput) {
	s.logger.Trace("Session: '%s':  Plugin output: '%s'", s.Name, output)
	// TODO: Write to buffer & send to clients with that buffer as active
}

func (s *Session) handleServerOutput(output string) {
	s.logger.Trace("Session: '%s':  Server output: '%s'", s.Name, output)

	s.pluginEngine.InTextLineChan <- output
}

func (s *Session) handleClientInput(input string) {
	s.logger.Trace("Session: '%s': Client input: '%s'", s.Name, strings.TrimSpace(input))

	// Check if input is a runes command, otherwise send to plugin engine
	if ok := s.handleCommand(input); !ok {
		s.pluginEngine.InCommandChan <- input
	}
}

func (s *Session) handleCommand(input string) bool {
	// TODO: Implement mud client command handling
	return false
}
