package httpdsak

import (
	"context"
	"fmt"
	"net/http"
)

func (c *Client) buildRequest(ctx context.Context) error {
	req, err := http.NewRequestWithContext(
		c.getTracerContext(ctx),
		c.method,
		c.url.String(),
		wrapSizedBody(c.body),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	c.req = req
	c.setRequestHeaders()
	return nil
}
