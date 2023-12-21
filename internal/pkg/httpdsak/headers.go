package httpdsak

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/jucrouzet/dsak/internal/pkg/version"
)

func (c *Client) setRequestHeaders() {
	headers := c.req.Header.Clone()
	headers.Set("user-agent", fmt.Sprintf("dsak/%s", version.GetFullVersion()))
	if c.acccept != "" {
		headers.Set("accept", c.acccept)
	}
	if c.contentType != "" {
		headers.Set("content-type", c.contentType)
	}
	for n, vs := range c.headers {
		for _, v := range vs {
			headers.Add(n, v)
		}
	}
	c.req.ContentLength = -1
	wb, ok := c.req.Body.(*sizedBody)
	if ok {
		c.req.ContentLength = wb.Size()
	}
	c.req.Header = headers
}

func (c *Client) showResponseHeaders(headers http.Header) {
	for k, v := range headers {
		c.traceInfo("Response header ")
		c.traceValuef("%q", k)
		c.traceInfo(" : ")
		if len(v) > 1 {
			c.traceValuef("%q\n", v)
		} else {
			c.traceValuef("%q\n", v[0])
		}
		if c.forceType != "" && strings.EqualFold(k, "content-type") {
			c.traceInfo("Forcing Content-Type to ")
			c.traceValuef("%s\n", c.forceType)
		}
	}
	c.traceInfoln("")
}
