package shell

// ShellProxy creates a connectable entity that executes a shell command and
// allows interaction with that shell command from within a runes session.
//
// For example, you could run 'ssh user@foo.com' in a ShellProxy and then
// use runes as your interface to that command.
type ShellProxy struct{}
