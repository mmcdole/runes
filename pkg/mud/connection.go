package mud

import (
	"io"
)

type Connection interface {
	Connect(host string, port int) error
	Close() error
	io.ReadWriter
}
