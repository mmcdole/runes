package mock

import (
	"fmt"
	"strings"

	"github.com/mmcdole/runes/internal/util"
)

// DefaultProxy is a mock server intended to respond to commands
// while clients are attached to the default session.
type DefaultProxy struct {
	inputChan  chan string
	outputChan chan string
	logger     util.Logger
}

func NewDefaultProxy(logger util.Logger) *DefaultProxy {
	return &DefaultProxy{
		inputChan:  make(chan string),
		outputChan: make(chan string),
		logger:     logger,
	}
}

func (ds *DefaultProxy) Connect() error {
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

func (ds *DefaultProxy) Input() chan string {
	return ds.inputChan
}

func (ds *DefaultProxy) Output() chan string {
	return ds.outputChan
}

func (ds *DefaultProxy) Close() error {
	return nil
}

func (ds *DefaultProxy) handleCommand(input string) {
	ds.logger.Trace("[DefaultProxy]: Command In: %s", strings.TrimSpace(input))
	// TODO: respond to different commands
	ds.sendText(fmt.Sprintf("Command '%s' not found!\n", strings.TrimSpace(input)))
}

func (ds *DefaultProxy) sendText(text string) {
	ds.logger.Trace("[DefaultProxy]: Text Out: %s", strings.TrimSpace(text))

	ds.outputChan <- text
}
