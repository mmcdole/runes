package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/mmcdole/runes/pkg/client"
	"github.com/mmcdole/runes/pkg/events"
)

func main() {
	// Define command line flags
	scriptDir := flag.String("scripts", "", "Directory containing Lua scripts")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Create event processor
	eventProcessor := events.New()

	// Create client with script directory and debug flag
	client, err := client.NewClient(*scriptDir, eventProcessor, *debug)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}

	// Start the TUI
	if err := client.Run(); err != nil {
		fmt.Printf("Failed to run client: %v\n", err)
		client.Close()
		os.Exit(1)
	}

	// Handle Ctrl+C gracefully
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Wait for interrupt signal
	go func() {
		<-c
		client.Close()
		os.Exit(0)
	}()

	// Keep the main goroutine alive
	select {}
}
