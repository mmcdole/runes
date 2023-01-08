package core

import (
	"fmt"
	"strings"

	"github.com/mmcdole/runes/internal/proxy/telnet"
)

type SessionCommand struct{}

// Handle "session" command
// session list - Show info about availble sessions
// session create <name> telnet vikingmud.org 2001 - Create telnet session
// session create <name> shell <command> - Create shell session
// session kill <name> - Kill session
// session switch <name> - Switch to existing session
func (c *SessionCommand) Execute(params *CommandParams) bool {
	if len(params.Args) == 0 {
		return c.handleSessionListCommand(params)
	}

	action := strings.ToLower(params.Args[0])
	switch action {
	case "list":
		return c.handleSessionListCommand(params)
	case "create":
		return c.handleSessionCreateCommand(params)
	case "kill":
		return c.handleSessionKillCommand(params)
	case "switch":
		return c.handleSessionSwitchCommand(params)
	default:
		return false
	}
}

// Handle "session list"
func (s *SessionCommand) handleSessionListCommand(params *CommandParams) bool {
	params.Session.writeClientText(params.Executor, "Sessions: ")
	sessions := params.Session.sessionManager.GetSessions()
	for _, session := range sessions {
		if session == params.Session {
			params.Session.writeClientText(params.Executor, fmt.Sprintf("  [*] %s", session.Name))
		} else {
			params.Session.writeClientText(params.Executor, fmt.Sprintf("  [ ] %s", session.Name))
		}
	}
	return true
}

// Handle "session create <name> telnet <host> <port>" command
// Handle "session create <name> shell <command> <arg1> <arg2>"
func (s *SessionCommand) handleSessionCreateCommand(params *CommandParams) bool {
	if len(params.Args) < 4 {
		return false
	}

	sessionType := strings.ToLower(params.Args[2])
	switch sessionType {
	case "telnet":
		return s.handleSessionCreateTelnetCommand(params)
	case "shell":
		return s.handleSessionCreateShellCommand(params)
	default:
		return false
	}
}

// Handle "session create <name> telnet <host> <port>" command
func (s *SessionCommand) handleSessionCreateTelnetCommand(params *CommandParams) bool {
	if len(params.Args) < 5 {
		return false
	}

	sessionName := params.Args[1]
	telnetHost := params.Args[3]
	telnetPort := params.Args[4]

	// TODO: Validate telnet parameters/args

	telnetProxy := telnet.NewTelnetProxy(&params.Session.log, telnetHost, telnetPort)
	session := params.Session.sessionManager.CreateSession(sessionName, telnetProxy)
	session.Start()
	telnetProxy.Connect()
	params.Session.sessionManager.AttachClientSession(params.Executor, sessionName)

	return true
}

func (s *SessionCommand) handleSessionCreateShellCommand(params *CommandParams) bool {
	// TODO: Implement shell command server type
	return true
}

// Handle "!session kill <name>" command
func (s *SessionCommand) handleSessionKillCommand(params *CommandParams) bool {
	return true
}

// Handle "!session switch <name>" command
func (s *SessionCommand) handleSessionSwitchCommand(params *CommandParams) bool {
	if len(params.Args) < 2 {
		return false
	}

	sessionName := strings.ToLower(params.Args[1])
	if params.Session.Name == sessionName {
		params.Session.writeClientText(params.Executor, fmt.Sprintf("Session '%s' already active!", params.Session.Name))
		return true
	}

	params.Session.writeClientText(params.Executor, fmt.Sprintf("Switching to '%s' session.", sessionName))
	params.Session.sessionManager.AttachClientSession(params.Executor, sessionName)

	return true
}

func (c *SessionCommand) Usage() string {
	return "Session Usage!"
}

func (c *SessionCommand) Help() string {
	return "Session Help!"
}
