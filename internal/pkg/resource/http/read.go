package http

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync/atomic"

	"go.uber.org/zap"
)

// Read implements io.Reader.
func (r *R) Read(p []byte) (int, error) {
	reader, release, err := r.getReader()
	if err != nil {
		return 0, err
	}
	defer release()
	return reader.Read(p)
}

func (r *R) getReader() (io.ReadCloser, func(), error) {
	r.mtx.Lock()
	if r.reader == nil { //nolint:nestif
		if r.writer != nil {
			return nil, func() {}, errors.New("resource is already in use as a writer, cannot read")
		}
		req, err := http.NewRequest(http.MethodGet, r.url.String(), http.NoBody)
		if err != nil {
			r.mtx.Unlock()
			return nil, func() {}, err
		}
		req = req.WithContext(r.cmd.Context())
		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			r.mtx.Unlock()
			return nil, func() {}, err
		}
		if res.StatusCode >= http.StatusBadRequest {
			r.mtx.Unlock()
			return nil, func() {}, fmt.Errorf("resource returned status %d", res.StatusCode)
		}
		if res.Header.Get("Content-Length") != "" {
			size, err := strconv.ParseInt(res.Header.Get("Content-Length"), 10, 64)
			if err != nil {
				r.logger.With(zap.Error(err)).Debug("failed to parse Content-Length header")
			} else {
				atomic.StoreInt64(r.size, size)
			}
		}
		r.reader = res.Body
		return res.Body, r.mtx.Unlock, nil
	}
	return r.reader, r.mtx.Unlock, nil
}
