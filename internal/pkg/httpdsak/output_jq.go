package httpdsak

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) outputJQ(ctx context.Context, res *http.Response) error {
	var decoded any
	if err := json.NewDecoder(res.Body).Decode(&decoded); err != nil {
		return fmt.Errorf("response is not valid json: %w", err)
	}
	iter := c.jq.RunWithContext(ctx, decoded)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return err
		}
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		if err := c.outputJSONBytes(ctx, b); err != nil {
			return err
		}
		fmt.Fprintln(c.out, "")
	}
	return nil
}
