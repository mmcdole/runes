package mock

import (
	"fmt"
	"strings"

	"github.com/mmcdole/runes/internal/util"
)

// DefaultServer is a mock server intended to respond to commands
// while clients are attached to the default session.
type DefaultServer struct {
	inputChan  chan string
	outputChan chan string
	logger     util.Logger
}

func NewDefaultServer(logger util.Logger) *DefaultServer {
	return &DefaultServer{
		inputChan:  make(chan string),
		outputChan: make(chan string),
		logger:     logger,
	}
}

func (ds *DefaultServer) Connect() error {
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

func (ds *DefaultServer) Input() chan string {
	return ds.inputChan
}

func (ds *DefaultServer) Output() chan string {
	return ds.outputChan
}

func (ds *DefaultServer) Close() error {
	return nil
}

func (ds *DefaultServer) handleCommand(input string) {
	ds.logger.Trace("[DefaultServer]: Command In: '%s'", strings.TrimSpace(input))
	// TODO: respond to different commands
	ds.sendText(fmt.Sprintf("Command '%s' not found!\n", strings.TrimSpace(input)))
}

func (ds *DefaultServer) sendText(text string) {
	ds.logger.Trace("[DefaultServer]: Text Out: '%s'", strings.TrimSpace(text))

	ds.outputChan <- text
}
