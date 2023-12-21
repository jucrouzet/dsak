package httpdsak

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/alecthomas/chroma/quick"
)

func (c *Client) outputJSON(ctx context.Context, res *http.Response) error {
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body : %w", err)
	}
	return c.outputJSONBytes(ctx, b)
}

func (c *Client) outputJSONBytes(_ context.Context, b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return fmt.Errorf("response is not valid json: %w", err)
	}
	str, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}
	return quick.Highlight(c.out, string(str), "json", "terminal", c.style)
}
