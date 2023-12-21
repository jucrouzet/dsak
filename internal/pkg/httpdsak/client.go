package httpdsak

import (
	"context"
	"fmt"
	"net/http"
)

func (c *Client) buildClient(ctx context.Context) error {
	tr, err := c.buildTransport(ctx)
	if err != nil {
		return fmt.Errorf("failed to build HTTP transport: %w", err)
	}
	c.client = &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return nil
}
