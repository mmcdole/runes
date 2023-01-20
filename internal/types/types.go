package types

type BufferOutput struct {
	Line       string
	BufferName string
}

type ClientCommand struct {
	Client Connection
	Text   string
}

type Connection interface {
	ID() string
	Name() string
	InputChan() chan *ClientCommand
	SetInputChan(chan *ClientCommand)
	OutputChan() chan string
	DisconnectChan() chan Connection
	Close() error
}
