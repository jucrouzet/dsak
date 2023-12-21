package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrTooLarge = errors.New("http resource: request body too large")
)

type memoryWriter struct {
	buffer  *bytes.Buffer
	ctx     context.Context
	maxSize int
	url     string
}

func newMemoryWriter(ctx context.Context, url string) *memoryWriter {
	return &memoryWriter{
		buffer:  new(bytes.Buffer),
		ctx:     ctx,
		maxSize: 10 * 1024 * 1024,
		url:     url,
	}
}

// Write implements io.Writer.
func (w *memoryWriter) Write(p []byte) (int, error) {
	if w.buffer.Len()+len(p) > w.maxSize {
		return 0, ErrTooLarge
	}
	return w.buffer.Write(p)
}

// Close implements io.Closer.
func (w *memoryWriter) Close() error {
	req, err := http.NewRequest(http.MethodPost, w.url, w.buffer)
	if err != nil {
		return err
	}
	req = req.WithContext(w.ctx)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("resource returned status %d", res.StatusCode)
	}
	return res.Body.Close()
}
