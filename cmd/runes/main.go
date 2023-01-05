package main

import (
	"time"

	"github.com/mmcdole/runes/internal/config"
	"github.com/mmcdole/runes/internal/core"
	"github.com/mmcdole/runes/internal/server"
	"github.com/mmcdole/runes/internal/util"
)

var (
// telnetServer    *telnet.TelnetServer
// websocketServer *websocket.WebsocketServer
)

func main() {
	logger := util.Logger{}
	logger.SetLogLevel(util.TraceLogLevel)
	logger.Info("Runes Launched!")

	conf := config.LoadOrCreateConfig()

	// SessionManager owns all Rune's game sessions and facilitates Connection's to a Session
	sessionManager := core.NewSessionManager(logger, conf)

	// ServerManager owns all configured servers (telnet, websocket) and forwards all new client connections
	serverManager := server.NewServerManager(logger, conf, sessionManager.ConnectedChan(), sessionManager.DisconnectedChan())

	// Initiate the default session and begin handling connections
	sessionManager.Start()

	// Run all configured servers and forward any connections/disconnections
	serverManager.Start()

	for {
		time.Sleep(time.Second)
	}
}

// func setupTelnetServer(logger util.Logger, conf *config.TelnetClientConfig, onConnect chan types.Connection) {
// 	address := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
// 	telnetServer = telnet.NewTelnetServer(logger, address, onConnect)
// 	telnetServer.Run()
// }

// func setupWebsocketServer(logger util.Logger, conf *config.WebsocketClientConfig, onConnect chan types.Connection) {
// 	// TODO: Setup websocket client server
// 	websocketServer = &websocket.WebsocketServer{}
// 	websocketServer.Run()
// }
