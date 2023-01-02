package core

import (
	"fmt"

	"github.com/mmcdole/runes/internal/plugin"
)

func NewSession(name string, server *ServerConnection, sm *SessionManager) *Session {
	pe := &plugin.PluginEngine{
		InCommandChan:   make(chan string),
		InTextLineChan:  make(chan string),
		OutSendChan:     make(chan string),
		OutCommandChan:  make(chan string),
		OutTextLineChan: make(chan plugin.TextLineOutput),
	}

	return &Session{
		Name:          name,
		manager:       sm,
		pluginEngine:  pe,
		clients:       []ClientConnection{},
		clientBuffers: map[*ClientConnection]string{},
		buffers:       map[string][]string{},
	}
}

type Session struct {
	Name          string
	manager       *SessionManager
	pluginEngine  *plugin.PluginEngine
	server        ServerConnection
	clients       []ClientConnection
	clientBuffers map[*ClientConnection]string
	buffers       map[string][]string
}

func (s *Session) AttachClient(client ClientConnection) {
	s.clients = append(s.clients, client)
}

func (s *Session) DetachClient(client ClientConnection) {
	// Remove the connection from the connections slice
	for i, c := range s.clients {
		if c == client {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			break
		}
	}
}

func (s *Session) SwitchToSession(client ClientConnection, sessionName string) (*Session, error) {
	// Detach the client connection from the current session
	s.DetachClient(client)

	// Get the new session and attach the client connection to it
	session := s.manager.GetSession(sessionName)
	if session == nil {
		return nil, fmt.Errorf("Switch session failed: target session '%s' not found.", sessionName)
	}
	session.AttachClient(client)
	return session, nil
}

func (s *Session) Start() {
	// // Receive input from client connections
	// go func() {
	// 	for {
	// 		select {
	// 		case input := <-s.ClientConnections[i].InputChan:
	// 			s.handleInput(input)
	// 		}
	// 	}
	// }()

	// // Receive input from the plugin engine and process it, or forward it
	// go func() {
	// 	for {
	// 		select {
	// 		case input := <-s.PluginEngine.CommandProcessedChan:
	// 			s.forwardInput(input)
	// 		case input := <-s.PluginEngine.SendChan:
	// 			s.handleInput(input)
	// 		}
	// 	}
	// }()

	// // Receive output from the server and send it to the plugin engine
	// go func() {
	// 	for {
	// 		select {
	// 		case output := <-s.ServerConnection.OutputChan:
	// 			s.processOutputFromServer(output)
	// 		}
	// 	}
	// }()
}

// func (s *Session) processOutputFromServer(output string) {
// 	s.PluginEngine.OutputPendingChan <- output
// }

// func (s *Session) forwardInputToServer(input string) {
// 	s.ServerConnection.InputChan <- input
// }

// func (s *Session) handleInput(input string) {
// 	// Check if input is a command, otherwise send to plugin engine
// 	if ok := s.handleCommand(input); !ok {
// 		s.PluginEngine.CommandPendingChan <- input
// 	}
// }

// func (s *Session) handleCommand(input string) bool {
// 	// TODO: Implement command handling
// 	return false
// }
