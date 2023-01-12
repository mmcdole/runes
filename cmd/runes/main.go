package main

import (
	"time"

	"github.com/fatih/color"
	"github.com/mmcdole/runes/internal/config"
	"github.com/mmcdole/runes/internal/core"
	"github.com/mmcdole/runes/internal/server"
	"github.com/mmcdole/runes/internal/util"
)

func main() {
	logger := util.Logger{LogLevel: util.TraceLogLevel}
	logger.Info("Runes Launched!")

	conf := config.LoadOrCreateConfig()
	color.NoColor = !conf.Core.EnableColors

	// SessionManager owns all Rune's game sessions and facilitates Connection's to a Session
	sessionManager := core.NewSessionManager(logger, conf)

	// ServerManager owns all configured servers (telnet, websocket) and forwards all new client connections
	serverManager := server.NewServerManager(logger, conf, sessionManager.ConnectedChan(), sessionManager.DisconnectedChan())

	// Run all configured servers and forward any connections/disconnections
	serverManager.Start()

	// Initiate the default session and begin handling connections
	sessionManager.Start()

	for {
		time.Sleep(time.Second)
	}
}
