package main

import (
	"fmt"

	"github.com/mmcdole/runes/internal/client"
	"github.com/mmcdole/runes/internal/client/telnet"
	"github.com/mmcdole/runes/internal/client/websocket"
	"github.com/mmcdole/runes/internal/config"
	"github.com/mmcdole/runes/internal/core"
	"github.com/mmcdole/runes/internal/util"
)

var (
	// Client Connection Sources
	telnetServer    *telnet.TelnetServer
	websocketServer *websocket.WebsocketServer

	// Client Connection Events
	connectionChan chan client.ClientConnection

	// Session Manager which attaches clients to sessions
	sessionManager *core.SessionManager
)

func main() {
	log := util.Logger{}
	log.SetLogLevel(util.TraceLogLevel)

	conf := config.LoadOrCreateConfig()

	// Setup session manager, default session and start attaching clients
	sessionManager = core.NewSessionManager()
	sessionManager.Run()

	// Connection chan for when any client source start producing connections
	connectionChan := make(chan client.ClientConnection)

	// Setup telnet client server if configured
	if conf.Client.Telnet != nil {
		setupTelnetServer(conf.Client.Telnet)
	}

	// Setup websocket client server if configured
	if conf.Client.Websocket != nil {
		setupWebsocketServer(conf.Client.Websocket)
	}

	// Receive client connections from servers and pass to session manager
	for {
		select {
		case conn := <-connectionChan:
			sessionManager.ConnectedChan() <- conn
		}
	}
}

func setupTelnetServer(conf *config.TelnetClientConfig) {
	address := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	telnetServer = telnet.NewTelnetServer(address, connectionChan)
	telnetServer.Run()
}

func setupWebsocketServer(conf *config.WebsocketClientConfig) {
	// TODO: Setup websocket client server
	websocketServer = &websocket.WebsocketServer{}
	websocketServer.Run()
}
