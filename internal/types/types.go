package types

type ConnectionInput struct {
	Client Connection
	Text   string
}

type Connection interface {
	ID() string
	Name() string
	InputChan() chan *ConnectionInput
	SetInputChan(chan *ConnectionInput)
	OutputChan() chan string
	DisconnectChan() chan Connection
	Close() error
}
