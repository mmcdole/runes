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
