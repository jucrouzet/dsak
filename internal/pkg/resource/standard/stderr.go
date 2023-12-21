package standard

import (
	"fmt"
	"os"
)

// StdErr represents stderr as an io.ReadWriteCloser.
type StdErr int

func NewStdErr() (*StdErr, error) {
	return new(StdErr), nil
}

// Read implements io.Reader.
func (r *StdErr) Read(_ []byte) (int, error) {
	return 0, fmt.Errorf("cannot read from stderr")
}

// Close implements io.Closer.
func (r *StdErr) Close() error {
	return os.Stderr.Close()
}

// Write implements io.Writer.
func (r *StdErr) Write(b []byte) (int, error) {
	return os.Stderr.Write(b)
}

// Size implements resourcetype.Handler.
func (r *StdErr) Size() int64 {
	return 0
}
