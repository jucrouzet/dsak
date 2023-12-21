package resourcetype

import "io"

// Handler is a resource handler.
type Handler interface {
	io.ReadWriteCloser
	Size() int64
}
