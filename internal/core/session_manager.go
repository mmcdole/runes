package core

import (
	"github.com/mmcdole/runes/internal/client"
	"github.com/mmcdole/runes/internal/config"
	"github.com/mmcdole/runes/internal/server"
	"github.com/mmcdole/runes/internal/server/runes"
	"github.com/mmcdole/runes/internal/util"
)

const defaultSessionName = "default"

type SessionManager struct {
	connected chan client.ClientConnection
	sessions  map[string]*Session
	logger    util.Logger
	config    *config.Config
}

func NewSessionManager(logger util.Logger, config *config.Config) *SessionManager {
	sm := &SessionManager{
		sessions:  map[string]*Session{},
		connected: make(chan client.ClientConnection),
		logger:    logger,
		config:    config,
	}

	return sm
}

func (sm *SessionManager) Start() {
	sm.logger.Debug("[SessionManager]: Started")

	// Setup default session for when clients initially connect
	defaultServer := runes.NewRunesServer(sm.logger)
	defaultSession := sm.CreateSession(defaultSessionName, defaultServer)

	defaultSession.Start()
	defaultServer.Connect()

	go func() {
		for {
			select {
			case client := <-sm.connected:
				sm.handleConnected(client)
			}
		}
	}()
}

func (sm *SessionManager) CreateSession(name string, server server.ServerConnection) *Session {
	if _, ok := sm.sessions[name]; ok {
		// session with the same name already exists
		return nil
	}

	session := NewSession(sm.logger, sm.config, name, server, sm)

	sm.sessions[name] = session
	sm.logger.Debug("[SessionManager]: Created '%s' Session", name)
	return session
}

func (sm *SessionManager) GetSession(name string) *Session {
	if session, ok := sm.sessions[name]; ok {
		return session
	}
	return nil
}

func (sm *SessionManager) GetSessions() []*Session {
	sessions := []*Session{}
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (sm *SessionManager) GetDefaultSession() *Session {
	return sm.GetSession(defaultSessionName)
}

func (sm *SessionManager) DeleteSession(name string) *Session {
	if name == defaultSessionName {
		// don't allow deleting the default session
		return nil
	}
	session, ok := sm.sessions[name]
	if !ok {
		// session with the given name does not exist
		return nil
	}

	delete(sm.sessions, name)
	//	delete(sm.SessionClients, session)
	// TODO: Detach any clients and attach them to default session
	return session
}

func (sm *SessionManager) ConnectedChan() chan client.ClientConnection {
	return sm.connected
}

func (sm *SessionManager) handleConnected(client client.ClientConnection) {
	ds := sm.GetDefaultSession()
	ds.AttachClient(client)
}
