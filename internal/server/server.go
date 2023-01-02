package server

// ServerConnection that a session interacts with
type ServerConnection interface {
	Input() chan string
	Output() chan string
	Connect() error
	Close() error
}
