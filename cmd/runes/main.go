package main

import (
	"fmt"

	"github.com/mmcdole/runes/internal/client"
	"github.com/mmcdole/runes/internal/config"
	"github.com/mmcdole/runes/internal/core"
	"github.com/mmcdole/runes/internal/util"
)

var (
	telnetServer    *client.TelnetServer
	websocketServer *client.WebsocketServer
	connectionChan  chan core.ClientConnection
)

func main() {
	log := util.Logger{}
	log.SetLogLevel(util.TraceLogLevel)

	conf := config.LoadOrCreateConfig()

	// Setup session manager, default session and start attaching clients
	sessionManager := core.NewSessionManager()
	sessionManager.Run()

	// Connection chan for when any client source start producing connections
	connectionChan := make(chan core.ClientConnection)

	// Setup telnet client server if configured
	if conf.Client.TelnetServer != nil {
		setupTelnetServer(conf.Client.TelnetServer)
	}

	// Setup websocket client server if configured
	if conf.Client.WebsocketServer != nil {
		setupWebsocketServer(conf.Client.WebsocketServer)
	}

	// Receive connections indefinitely and pass to session manager
	for {
		select {
		case conn := <-connectionChan:
			sessionManager.ConnectedChan() <- conn
		}
	}
}

func setupTelnetServer(conf *config.TelnetClientConfig) {
	address := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	telnetServer = client.NewTelnetServer(address, connectionChan)
	telnetServer.Run()
}

func setupWebsocketServer(conf *config.WebsocketClientConfig) {
	// TODO: Setup websocket client server
	websocketServer = &client.WebsocketServer{}
	websocketServer.Run()
}
