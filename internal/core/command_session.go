package core

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
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
	case "connect":
		return c.handleSessionConnectCommand(params)
	default:
		return false
	}
}

// Handle "session list"
func (s *SessionCommand) handleSessionListCommand(params *CommandParams) bool {
	params.writeToExecutor("Sessions:\n")
	sessions := params.Session.sessionManager.GetSessions()
	for _, session := range sessions {
		if session == params.Session {
			params.writeToExecutor(fmt.Sprintf(" [%s] %s\n", color.HiGreenString("*"), session.Name))
		} else {
			params.writeToExecutor(fmt.Sprintf(" [ ] %s\n", session.Name))
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
	params.Session.sessionManager.AttachClientSession(params.Executor, sessionName)
	params.writeToExecutor(fmt.Sprintf("Attached to session: %s\n", sessionName))

	session.Start()
	params.writeToExecutor(fmt.Sprintf("Connecting to '%s:%s'...\n", telnetHost, telnetPort))
	if err := telnetProxy.Connect(); err != nil {
		params.writeToExecutor("Connection failed!\n")
	} else {
		params.writeToExecutor("Connected successfully!\n")
	}

	return true
}

func (s *SessionCommand) handleSessionCreateShellCommand(params *CommandParams) bool {
	// TODO: Implement shell command server type
	return true
}

// Handle "!session kill <name>" command
func (s *SessionCommand) handleSessionKillCommand(params *CommandParams) bool {
	// TODO: Implement session kill command
	return true
}

// Handle "!session switch <name>" command
func (s *SessionCommand) handleSessionSwitchCommand(params *CommandParams) bool {
	if len(params.Args) < 2 {
		return false
	}

	sessionName := strings.ToLower(params.Args[1])
	if params.Session.Name == sessionName {
		params.writeToExecutor(fmt.Sprintf("Session '%s' already active!\n", params.Session.Name))
		return true
	}

	params.writeToExecutor(fmt.Sprintf("Switching to '%s' session.\n", sessionName))
	params.Session.sessionManager.AttachClientSession(params.Executor, sessionName)

	return true
}

func (s *SessionCommand) handleSessionConnectCommand(params *CommandParams) bool {
	params.writeToExecutor("Reconnecting...\n")
	err := params.Session.proxyConnection.Connect()
	if err != nil {
		params.writeToExecutor("Connection failed!\n")
	} else {
		params.writeToExecutor("Connected successfully!\n")
	}
	return true
}

func (c *SessionCommand) Usage() string {
	// TODO: Session Usage
	return "Session Usage!"
}

func (c *SessionCommand) Help() string {
	// TODO: Session Help
	return "Session Help!"
}
