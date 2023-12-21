package resource

import (
	"io"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/jucrouzet/dsak/internal/pkg/resource/resourcetype"
)

// Type is the type of a resource.
type Type uint

// R represents a resource in dsak.
type R struct {
	cmd      *cobra.Command
	logger   *zap.Logger
	resource resourcetype.Handler
	url      string
}

// New returns a new resource.
func New(cmd *cobra.Command, url string, logger *zap.Logger) (*R, error) {
	logger = logger.With(zap.String("resource", url))
	r := &R{
		cmd:    cmd,
		logger: logger,
		url:    url,
	}
	res, err := resourcetype.Parse(cmd, url, logger)
	if err != nil {
		return nil, err
	}
	r.resource = res
	return r, nil
}

type ioResult struct {
	count int
	err   error
}

// Read implements io.Reader.
func (r *R) Read(p []byte) (int, error) {
	ctx := r.cmd.Context()
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	res := make(chan ioResult)
	go func(r io.ReadWriteCloser, p *[]byte) {
		c, err := r.Read(*p)
		res <- ioResult{c, err}
	}(r.resource, &p)
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case res := <-res:
		return res.count, res.err
	}
}

// Write implements io.Writer.
func (r *R) Write(p []byte) (int, error) {
	ctx := r.cmd.Context()
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	res := make(chan ioResult)
	go func(r io.ReadWriteCloser, p *[]byte) {
		c, err := r.Write(*p)
		res <- ioResult{c, err}
	}(r.resource, &p)
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case res := <-res:
		return res.count, res.err
	}
}

// Close implements io.Closer.
func (r *R) Close() error {
	ctx := r.cmd.Context()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	res := make(chan error)
	go func(r io.ReadWriteCloser) {
		res <- r.Close()
	}(r.resource)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case res := <-res:
		return res
	}
}

// Size returns the size of the resource.
func (r *R) Size() int64 {
	return r.resource.Size()
}
