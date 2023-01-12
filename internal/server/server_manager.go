package server

import (
	"fmt"

	"github.com/mmcdole/runes/internal/config"
	"github.com/mmcdole/runes/internal/server/ssl"
	"github.com/mmcdole/runes/internal/server/telnet"
	"github.com/mmcdole/runes/internal/server/websocket"
	"github.com/mmcdole/runes/internal/types"
	"github.com/mmcdole/runes/internal/util"
)

func NewServerManager(log util.Logger, config *config.Config, connected chan types.Connection, disconnected chan types.Connection) *ServerManager {
	return &ServerManager{
		log:          log,
		config:       config,
		connected:    connected,
		disconnected: disconnected,
	}
}

type ServerManager struct {
	config       *config.Config
	log          util.Logger
	connected    chan types.Connection
	disconnected chan types.Connection

	telnetServer    *telnet.TelnetServer
	sslServer       *ssl.SSLServer
	websocketServer *websocket.WebsocketServer
}

func (sm *ServerManager) Start() {
	// Setup Telnet Server if configured
	if sm.config.Server.Telnet != nil {
		conf := sm.config.Server.Telnet
		address := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
		sm.telnetServer = telnet.NewTelnetServer(sm.log, address, sm.connected, sm.disconnected)
		sm.telnetServer.Run()
	}

	// Setup SSL Server if configured
	if sm.config.Server.SSL != nil {
		conf := sm.config.Server.SSL
		address := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
		sm.sslServer = ssl.NewSSLServer(sm.log, address, sm.connected, sm.disconnected)
		sm.sslServer.Run()
	}

	// Setup Websocket Server if configured
	if sm.config.Server.Websocket != nil {
		// conf := sm.config.Server.Websocket
		sm.websocketServer = &websocket.WebsocketServer{}
		sm.websocketServer.Run()
	}
}
