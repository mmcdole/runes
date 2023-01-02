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

	sessionManager *core.SessionManager
	logger         util.Logger
)

func main() {
	logger = util.Logger{}
	logger.SetLogLevel(util.TraceLogLevel)

	logger.Info("Runes Launched!")

	conf := config.LoadOrCreateConfig()

	// Connection chan for when any client source start producing connections
	connectionChan = make(chan client.ClientConnection)

	// Setup telnet client server if configured
	if conf.Client.Telnet != nil {
		setupTelnetServer(conf.Client.Telnet)
	}

	// Setup websocket client server if configured
	if conf.Client.Websocket != nil {
		setupWebsocketServer(conf.Client.Websocket)
	}

	// Setup session manager, default session and start attaching clients
	sessionManager = core.NewSessionManager(logger)
	sessionManager.Start()

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
	telnetServer = telnet.NewTelnetServer(logger, address, connectionChan)
	telnetServer.Run()
}

func setupWebsocketServer(conf *config.WebsocketClientConfig) {
	// TODO: Setup websocket client server
	websocketServer = &websocket.WebsocketServer{}
	websocketServer.Run()
}
