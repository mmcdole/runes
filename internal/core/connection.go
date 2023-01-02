package core

// ClientSource facilitates clients connecting to runes and producing
// ClientConnections
type ClientSource interface {
	Connected() chan ClientConnection
}

// ClientConnection to interact with a session
type ClientConnection interface {
	Name() string
	InputChan() chan string
	OutputChan() chan string
	DisconnectChan() chan bool
	Close() error
}

// ServerConnection that a session interacts with
type ServerConnection interface {
	Input() chan string
	Output() chan string
	Connect() error
	Close() error
}
