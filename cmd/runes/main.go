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
	telnetServer    *telnet.TelnetServer
	websocketServer *websocket.WebsocketServer
)

func main() {
	logger := util.Logger{}
	logger.SetLogLevel(util.TraceLogLevel)

	logger.Info("Runes Launched!")

	conf := config.LoadOrCreateConfig()

	// Connection chan for when any client source start producing connections
	onConnect := make(chan client.ClientConnection)

	// Setup telnet client server if configured
	if conf.Client.Telnet != nil {
		setupTelnetServer(logger, conf.Client.Telnet, onConnect)
	}

	// Setup websocket client server if configured
	if conf.Client.Websocket != nil {
		setupWebsocketServer(logger, conf.Client.Websocket, onConnect)
	}

	// Setup session manager, default session and start attaching clients
	sessionManager := core.NewSessionManager(logger, conf)
	sessionManager.Start()

	// Receive client connections from server(s) and pass to session manager
	for {
		select {
		case conn := <-onConnect:
			sessionManager.ConnectedChan() <- conn
		}
	}
}

func setupTelnetServer(logger util.Logger, conf *config.TelnetClientConfig, onConnect chan client.ClientConnection) {
	address := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	telnetServer = telnet.NewTelnetServer(logger, address, onConnect)
	telnetServer.Run()
}

func setupWebsocketServer(logger util.Logger, conf *config.WebsocketClientConfig, onConnect chan client.ClientConnection) {
	// TODO: Setup websocket client server
	websocketServer = &websocket.WebsocketServer{}
	websocketServer.Run()
}
