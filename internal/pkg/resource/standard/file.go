package standard

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
)

// File represents a file as an io.ReadWriteCloser.
type File struct {
	logger    *zap.Logger
	path      string
	reader    io.ReadCloser
	readerMtx sync.Mutex
	writer    io.WriteCloser
	writerMtx sync.Mutex
}

func NewFile(path string, logger *zap.Logger) (*File, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	logger = logger.
		With(zap.String("resource_type", "file")).
		With(zap.String("path", path))
	return &File{
		logger:    logger,
		path:      path,
		readerMtx: sync.Mutex{},
		writerMtx: sync.Mutex{},
	}, nil
}

// Read implements io.Reader.
func (r *File) Read(p []byte) (int, error) {
	f, err := r.getReader()
	if err != nil {
		return 0, err
	}
	return f.Read(p)
}

// Close implements io.Closer.
func (r *File) Close() error {
	var errs []error
	r.readerMtx.Lock()
	defer r.readerMtx.Unlock()
	if r.reader != nil {
		r.logger.Debug("closing file for reading")
		errs = append(errs, r.reader.Close())
	}
	r.writerMtx.Lock()
	defer r.writerMtx.Unlock()
	if r.writer != nil {
		r.logger.Debug("closing file for writing")
		errs = append(errs, r.writer.Close())
	}
	return errors.Join(errs...)
}

// Write implements io.Writer.
func (r *File) Write(b []byte) (int, error) {
	f, err := r.getWriter()
	if err != nil {
		return 0, err
	}
	return f.Write(b)
}

// Size implements resourcetype.Handler.
func (r *File) Size() int64 {
	s, err := os.Stat(r.path)
	if err != nil {
		r.logger.With(zap.Error(err)).Debug("failed to get file size")
		return 0
	}
	return s.Size()
}

func (r *File) getReader() (io.Reader, error) {
	r.readerMtx.Lock()
	defer r.readerMtx.Unlock()
	if r.reader != nil {
		return r.reader, nil
	}
	r.logger.Debug("opening file for reading")
	f, err := os.Open(r.path)
	if err != nil {
		return nil, fmt.Errorf("failed opening file for reading: %w", err)
	}
	r.reader = f
	return f, nil
}

func (r *File) getWriter() (io.Writer, error) {
	r.writerMtx.Lock()
	defer r.writerMtx.Unlock()
	if r.writer != nil {
		return r.writer, nil
	}
	r.logger.Debug("opening file for writing")
	f, err := os.Create(r.path)
	if err != nil {
		return nil, fmt.Errorf("failed opening file for writing: %w", err)
	}
	r.writer = f
	return f, nil
}
