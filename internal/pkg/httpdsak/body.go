package httpdsak

import (
	"io"

	"github.com/jucrouzet/dsak/internal/pkg/resource"
)

type sizedBody struct {
	body io.ReadCloser
	size int64
}

func wrapSizedBody(body io.ReadCloser) *sizedBody {
	var size int64
	r, ok := body.(*resource.R)
	if ok {
		size = r.Size()
	}
	return &sizedBody{
		body: body,
		size: size,
	}
}

// Read implements io.Reader.
func (b *sizedBody) Read(p []byte) (int, error) {
	return b.body.Read(p)
}

// Close implements io.Closer.
func (b *sizedBody) Close() error {
	return b.body.Close()
}

// Close implements io.Closer.
func (b *sizedBody) Size() int64 {
	return b.size
}
