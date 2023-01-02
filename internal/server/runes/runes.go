package runes

import (
	"fmt"

	"github.com/mmcdole/runes/internal/util"
)

// RunesServer is a mock server intended to respond to commands
// while clients are attached to the default session.
type RunesServer struct {
	inputChan  chan string
	outputChan chan string
}

func NewDefaultServer() *RunesServer {
	return &RunesServer{
		inputChan:  make(chan string),
		outputChan: make(chan string),
	}
}

func (ds *RunesServer) Connect() error {
	go func() {
		for {
			select {
			case input := <-ds.inputChan:
				ds.handleInput(input)
			}
		}
	}()

	go func() {
		ds.outputChan <- util.WelcomeBanner
		ds.outputChan <- "Welcome to Runes default session!\n"
	}()

	return nil
}

func (ds *RunesServer) Input() chan string {
	return ds.inputChan
}

func (ds *RunesServer) Output() chan string {
	return ds.outputChan
}

func (ds *RunesServer) Close() error {
	return nil
}

func (ds *RunesServer) handleInput(input string) {
	// TODO: respond to different commands
	ds.outputChan <- fmt.Sprintf("Command '%s' not found!\n", input)
}
