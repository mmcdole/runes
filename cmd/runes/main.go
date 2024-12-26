package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/mmcdole/runes/pkg/luaengine"
	"github.com/mmcdole/runes/pkg/mud"
)

func main() {
	// Parse command line flags
	userScriptDir := flag.String("scripts", "", "Directory containing user scripts")
	flag.Parse()

	// Create the Lua engine
	engine := luaengine.New(luafiles.GetCoreScripts(), *userScriptDir)

	// Create and initialize the client
	client := mud.NewClient(engine)

	// Register the client's API with the engine
	engine.RegisterAPI(client)

	// Load all scripts
	if err := engine.LoadScripts(); err != nil {
		log.Fatalf("Failed to load scripts: %v", err)
	}

	// Handle Ctrl+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Simple input loop
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				log.Printf("Error reading input: %v", err)
				continue
			}
			input := string(buf[:n])
			client.HandleInput(input)
		}
	}()

	<-c // Wait for Ctrl+C
	log.Println("\nShutting down...")
	client.Disconnect()
}
