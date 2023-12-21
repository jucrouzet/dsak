package standard

import (
	"fmt"
	"os"
)

// StdOut represents stderr as an io.ReadWriteCloser.
type StdOut int

func NewStdOut() (*StdOut, error) {
	return new(StdOut), nil
}

// Read implements io.Reader.
func (r *StdOut) Read(_ []byte) (int, error) {
	return 0, fmt.Errorf("cannot read from stdout")
}

// Close implements io.Closer.
func (r *StdOut) Close() error {
	return os.Stdout.Close()
}

// Write implements io.Writer.
func (r *StdOut) Write(b []byte) (int, error) {
	return os.Stdout.Write(b)
}

// Size implements resourcetype.Handler.
func (r *StdOut) Size() int64 {
	return 0
}
