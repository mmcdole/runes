package runes

import (
	"fmt"
	"strings"

	"github.com/mmcdole/runes/internal/util"
)

// RunesServer is a mock server intended to respond to commands
// while clients are attached to the default session.
type RunesServer struct {
	inputChan  chan string
	outputChan chan string
	logger     util.Logger
}

func NewRunesServer(logger util.Logger) *RunesServer {
	return &RunesServer{
		inputChan:  make(chan string),
		outputChan: make(chan string),
		logger:     logger,
	}
}

func (ds *RunesServer) Connect() error {
	go func() {
		for {
			select {
			case input := <-ds.inputChan:
				ds.handleCommand(input)
			}
		}
	}()

	go func() {
		welcomeLines := strings.Split(util.WelcomeBanner, "\n")
		for _, line := range welcomeLines {
			ds.sendText(line + "\n")
		}
		ds.sendText("\n")
		ds.sendText("Welcome to Runes default session!\n")
		ds.sendText("\n")
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

func (ds *RunesServer) handleCommand(input string) {
	ds.logger.Trace("[DefaultServer]: Command In: '%s'", strings.TrimSpace(input))
	// TODO: respond to different commands
	ds.sendText(fmt.Sprintf("Command '%s' not found!\n", strings.TrimSpace(input)))
}

func (ds *RunesServer) sendText(text string) {
	ds.logger.Trace("[DefaultServer]: Text Out: '%s'", strings.TrimSpace(text))

	ds.outputChan <- text
}
