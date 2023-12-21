package standard

import (
	"fmt"
	"os"
)

// StdIn represents stdin as an io.ReadWriteCloser.
type StdIn int

func NewStdIn() (*StdIn, error) {
	return new(StdIn), nil
}

// Read implements io.Reader.
func (r *StdIn) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

// Close implements io.Closer.
func (r *StdIn) Close() error {
	return os.Stdin.Close()
}

// Write implements io.Writer.
func (r *StdIn) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("cannot write to stdin")
}

// Size implements resourcetype.Handler.
func (r *StdIn) Size() int64 {
	return 0
}
