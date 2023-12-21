package httpdsak

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/itchyny/gojq"
)

type Client struct {
	acccept     string
	body        io.ReadCloser
	client      *http.Client
	contentType string
	forceHTTP1  bool
	forceHTTP2  bool
	forceType   string
	headers     http.Header
	insecure    bool
	jq          *gojq.Query
	log         io.Writer
	method      string
	out         io.Writer
	raw         bool
	req         *http.Request
	style       string
	trace       bool
	url         *url.URL
}

func NewClient(uri string, opts ...Option) (*Client, error) {
	c := &Client{
		acccept:     "",
		body:        http.NoBody,
		contentType: "application/octet-stream",
		headers:     make(http.Header),
		insecure:    false,
		log:         os.Stderr,
		method:      "GET",
		out:         os.Stdout,
		style:       "monokai",
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	u, err := parseURL(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}
	c.url = u
	return c, nil
}

// Run runs the HTTP request.
func (c *Client) Run(ctx context.Context) error {
	err := c.buildRequest(ctx)
	if err != nil {
		return fmt.Errorf("failed to build HTTP request: %w", err)
	}
	err = c.buildClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to build HTTP client: %w", err)
	}
	res, err := c.client.Do(c.req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	c.traceInfo("Response status code is ")
	if res.StatusCode >= http.StatusBadRequest {
		c.traceErrorf("%s \n", res.Status)
	} else {
		c.traceValuef("%s \n", res.Status)
	}
	c.showResponseHeaders(res.Header)
	return c.output(ctx, res)
}
