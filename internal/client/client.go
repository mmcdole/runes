package client

// Wrapper struct to pass input client owner along with the input
type ClientInput struct {
	Client ClientConnection
	Text   string
}

// ClientConnection to interact with a session
type ClientConnection interface {
	Name() string
	InputChan() chan ClientInput
	SetInputChan(chan ClientInput)
	OutputChan() chan string
	DisconnectChan() chan bool
	Close() error
}
