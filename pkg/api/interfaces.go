package api

// MudAPI defines the interface for MUD client operations
type MudAPI interface {
	// Runes API
	Connect(host string, port int) error
	Disconnect() error
	Display(text string, buffer string)
	ListBuffers() []string
	SwitchBuffer(name string)

	// MUD API
	SendCommand(cmd string)  // Direct send
	QueueCommand(cmd string) // Queued send
	SendRaw(cmd string)      // Raw send
}
