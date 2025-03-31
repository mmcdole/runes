package history

// DefaultMaxSize is the default maximum number of commands to keep in history
const DefaultMaxSize = 1000

// History manages a list of command history.
// It provides simple storage and retrieval of commands.
type History struct {
	commands []string
	maxSize  int
}

// New creates a new History instance with default settings
func New() *History {
	return &History{
		commands: make([]string, 0, DefaultMaxSize),
		maxSize:  DefaultMaxSize,
	}
}

// Add appends a command to history, unless it's empty or a duplicate of the last command.
// The oldest command is removed if the history exceeds its maximum size.
func (h *History) Add(cmd string) {
	// Don't add empty commands or duplicates
	if cmd == "" || (len(h.commands) > 0 && h.commands[len(h.commands)-1] == cmd) {
		return
	}

	h.commands = append(h.commands, cmd)

	// Remove oldest if exceeding max
	if len(h.commands) > h.maxSize {
		h.commands = h.commands[1:]
	}
}

// Get returns the command at the given index.
// Returns the command and true if the index is valid,
// or empty string and false if the index is out of bounds.
func (h *History) Get(index int) (string, bool) {
	if index >= 0 && index < len(h.commands) {
		return h.commands[index], true
	}
	return "", false
}

// Length returns the number of commands in history
func (h *History) Length() int {
	return len(h.commands)
}

// Clear removes all commands from history
func (h *History) Clear() {
	h.commands = make([]string, 0, h.maxSize)
}

// SetMaxSize sets the maximum number of commands to keep in history.
// If the current history exceeds the new maximum, older entries are removed.
func (h *History) SetMaxSize(max int) {
	if max < 1 {
		max = 1
	}
	h.maxSize = max
	if len(h.commands) > max {
		h.commands = h.commands[len(h.commands)-max:]
	}
}
