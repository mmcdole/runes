package proxy

// ProxyConnection that a session interacts with
type ProxyConnection interface {
	Input() chan string
	Output() chan string
	Connect() error
	Close() error
}
