package core

import (
	"github.com/mmcdole/runes/internal/config"
	"github.com/mmcdole/runes/internal/proxy"
	"github.com/mmcdole/runes/internal/proxy/mock"
	"github.com/mmcdole/runes/internal/types"
	"github.com/mmcdole/runes/internal/util"
)

const defaultSessionName = "default"

type SessionManager struct {
	connected         chan types.Connection
	disconnected      chan types.Connection
	sessionsMap       map[string]*Session
	clientSessionsMap map[string]*Session
	logger            util.Logger
	config            *config.Config
}

func NewSessionManager(logger util.Logger, config *config.Config) *SessionManager {
	sm := &SessionManager{
		sessionsMap:       map[string]*Session{},
		clientSessionsMap: map[string]*Session{},
		connected:         make(chan types.Connection),
		disconnected:      make(chan types.Connection),
		logger:            logger,
		config:            config,
	}

	return sm
}

func (sm *SessionManager) ConnectedChan() chan types.Connection {
	return sm.connected
}

func (sm *SessionManager) DisconnectedChan() chan types.Connection {
	return sm.disconnected
}

func (sm *SessionManager) Start() {
	sm.logger.Debug("[SessionManager]: Started")

	// Setup the initial default session clients connect to
	sm.setupDefaultSession()

	// Handle client connect/disconnects
	go func() {
		for {
			select {
			case client := <-sm.connected:
				sm.handleConnected(client)
			case client := <-sm.disconnected:
				sm.handleDisconnected(client)
			}
		}
	}()
}

func (sm *SessionManager) GetClientSession(client types.Connection) *Session {
	if session, ok := sm.clientSessionsMap[client.ID()]; ok {
		return session
	}
	return nil
}

func (sm *SessionManager) AttachClientSession(client types.Connection, sessionName string) {
	target := sm.GetSession(sessionName)
	if target == nil {
		return
	}

	// Detach from existing session
	if existing := sm.GetClientSession(client); existing != nil {
		existing.DetachClient(client)
	}

	// Attach to target session
	target.AttachClient(client)

	// Update client > session map
	sm.clientSessionsMap[client.ID()] = target
}

func (sm *SessionManager) DetachClientSession(client types.Connection) {
	if existing := sm.GetClientSession(client); existing != nil {
		existing.DetachClient(client)
		delete(sm.clientSessionsMap, client.ID())
	}
}

func (sm *SessionManager) CreateSession(name string, server proxy.ProxyConnection) *Session {
	if _, ok := sm.sessionsMap[name]; ok {
		// session with the same name already exists
		return nil
	}

	session := NewSession(sm.logger, sm.config, name, server, sm)

	sm.sessionsMap[name] = session
	sm.logger.Debug("[SessionManager]: Created '%s' Session", name)
	return session
}

func (sm *SessionManager) GetSession(name string) *Session {
	if session, ok := sm.sessionsMap[name]; ok {
		return session
	}
	return nil
}

func (sm *SessionManager) GetSessions() []*Session {
	sessions := []*Session{}
	for _, session := range sm.sessionsMap {
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
	session, ok := sm.sessionsMap[name]
	if !ok {
		// session with the given name does not exist
		return nil
	}

	delete(sm.sessionsMap, name)

	//	delete(sm.SessionClients, session)
	// TODO: Detach any clients and attach them to default session
	return session
}

func (sm *SessionManager) handleConnected(client types.Connection) {
	sm.AttachClientSession(client, defaultSessionName)
}

func (sm *SessionManager) handleDisconnected(client types.Connection) {
	sm.DetachClientSession(client)
}

func (sm *SessionManager) setupDefaultSession() {
	// Setup default session for when clients initially connect
	defaultServer := mock.NewDefaultServer(sm.logger)
	defaultSession := sm.CreateSession(defaultSessionName, defaultServer)

	defaultSession.Start()
	defaultServer.Connect()
}
