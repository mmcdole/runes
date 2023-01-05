package core

import "fmt"

type SessionCommand struct{}

// Handle "session" command
// session list - Show info about availble sessions
// session create vk telnet vikingmud.org 2001 - Create telnet session
// session create vk shell <command> - Create shell session
// session kill vk - Kill session
// session switch vk - Switch to existing session
func (c *SessionCommand) Execute(params *CommandParams) bool {
	if len(params.Args) == 0 {
		return c.handleSessionListCommand(params)
	}

	action := params.Args[0]
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
	params.Session.writeText("Sessions: ")
	sessions := params.Session.sessionManager.GetSessions()
	for _, session := range sessions {
		if session == params.Session {
			params.Session.writeText(fmt.Sprintf("  [*] %s", session.Name))
		} else {
			params.Session.writeText(fmt.Sprintf("  [ ] %s", session.Name))
		}
	}
	return true
}

// Handle "!session create <name> <type> <arg1> <arg2>" command
func (s *SessionCommand) handleSessionCreateCommand(params *CommandParams) bool {
	return true
}

func (s *SessionCommand) handleSessionCreateTelnetCommand(params *CommandParams) bool {
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
	return true
}

func (c *SessionCommand) Usage() string {
	return "Session Usage!"
}

func (c *SessionCommand) Help() string {
	return "Session Help!"
}
