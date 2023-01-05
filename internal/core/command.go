package core

import "github.com/mmcdole/runes/internal/client"

type Command interface {
	Execute(params *CommandParams) bool
	Usage() string
	Help() string
}

type CommandParams struct {
	Command     string
	Args        []string
	Session     *Session
	Executor    *client.ClientConnection
	FullCommand string
}
