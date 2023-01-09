package core

import (
	"github.com/mmcdole/runes/internal/types"
)

type Command interface {
	Execute(params *CommandParams) bool
	Usage() string
	Help() string
}

type CommandParams struct {
	Command     string
	Args        []string
	Session     *Session
	Executor    types.Connection
	FullCommand string
}

// Utility function to write to the executing client
func (cp *CommandParams) writeToExecutor(text string) {
	cp.Session.writeClientLine(cp.Executor, text)
}

// Utility function to write to the current's sessions named buffer
func (cp *CommandParams) writeToBuffer(text string, bufferName string) {
	cp.Session.writeBufferText(bufferName, text)
}
