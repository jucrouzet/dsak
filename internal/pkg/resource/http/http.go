package http

import (
	"fmt"
	"io"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// R represents an HTTP resource as an io.ReadWriteCloser.
type R struct {
	cmd    *cobra.Command
	logger *zap.Logger
	url    *url.URL
	reader io.ReadCloser
	size   *int64
	writer io.WriteCloser
	mtx    sync.Mutex
}

func New(cmd *cobra.Command, uri string, logger *zap.Logger) (*R, error) {
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if parsedURI.Scheme != "http" && parsedURI.Scheme != "https" {
		return nil, fmt.Errorf("not a valid HTTP URL: %s", uri)
	}
	return &R{
		cmd:    cmd,
		logger: logger.With(zap.String("resource_type", "http")),
		url:    parsedURI,
		mtx:    sync.Mutex{},
		size:   new(int64),
	}, nil
}

// Size implements resourcetype.Handler.
func (r *R) Size() int64 {
	if v := atomic.LoadInt64(r.size); v != 0 {
		return v
	}
	_, release, err := r.getReader()
	if err != nil {
		return 0
	}
	defer release()
	return atomic.LoadInt64(r.size)
}
