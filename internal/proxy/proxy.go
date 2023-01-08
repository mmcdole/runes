package proxy

// ProxyConnection that a session interacts with
type ProxyConnection interface {
	Connect() error
	Close() error
	Input() chan string
	Output() chan string
}
