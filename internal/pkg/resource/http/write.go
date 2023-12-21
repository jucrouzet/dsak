package http

import (
	"errors"
	"io"
)

// Write implements io.Writer.
func (r *R) Write(p []byte) (int, error) {
	reader, release, err := r.getWriter()
	if err != nil {
		return 0, err
	}
	defer release()
	return reader.Write(p)
}

func (r *R) getWriter() (io.WriteCloser, func(), error) {
	r.mtx.Lock()
	if r.writer == nil {
		if r.reader != nil {
			return nil, func() {}, errors.New("resource is already in use as a reader, cannot write")
		}
		r.writer = newMemoryWriter(r.cmd.Context(), r.url.String())
	}
	return r.writer, r.mtx.Unlock, nil
}
