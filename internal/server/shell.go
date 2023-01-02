package server

// ShellServer creates a connectable entity that executes a shell command and
// allows interaction with that shell command from within a runes session.
//
// For example, you could run 'ssh user@foo.com' in a ShellServer and then
// use runes as your interface to that command.
type ShellServer struct{}
