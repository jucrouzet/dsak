package httpdsak

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"syscall"

	"github.com/alecthomas/chroma/quick"
	"golang.org/x/term"
)

func (c *Client) output(ctx context.Context, res *http.Response) error {
	if c.raw || !term.IsTerminal(syscall.Stdout) {
		return c.outputRaw(ctx, res)
	}
	if c.jq != nil {
		return c.outputJQ(ctx, res)
	}
	parts := strings.SplitN(res.Header.Get("content-type"), ";", 2)
	var err error
	mimeType := c.forceType
	if mimeType == "" {
		mimeType = strings.ToLower(strings.TrimSpace(parts[0]))
	}
	switch mimeType {
	case "application/json":
		err = c.outputJSON(ctx, res)
	case "application/xml", "text/xml":
		err = c.outputHighlighted(ctx, res, "xml")
	case "text/html":
		err = c.outputHighlighted(ctx, res, "html")
	case "image/png", "image/jpeg", "image/gif":
		err = c.outputImage(ctx, res)
	default:
		parts := strings.SplitN(mimeType, "/", 2)
		switch parts[0] {
		case "audio", "video":
			err = c.outputMedia(ctx, res)
		default:
			err = c.outputRaw(ctx, res)
		}
	}
	if err != nil {
		return fmt.Errorf("error while rendering response: %w", err)
	}
	return nil
}

func (c *Client) outputHighlighted(_ context.Context, res *http.Response, lexer string) error {
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body : %w", err)
	}
	return quick.Highlight(c.out, string(b), lexer, "terminal", c.style)
}
