package http

// Close implements io.Closer.
func (r *R) Close() error {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if r.reader != nil {
		return r.reader.Close()
	}
	if r.writer != nil {
		return r.writer.Close()
	}
	return nil
}
