package mud

import (
	"io"
)

// Connection represents a connection to a MUD server
type Connection interface {
	io.ReadWriteCloser
}
