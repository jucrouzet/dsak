package httpdsak

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"syscall"
	"unicode"

	"golang.org/x/term"
)

func (c *Client) outputRaw(_ context.Context, res *http.Response) error {
	if !c.raw && term.IsTerminal(syscall.Stdout) {
		p := make([]byte, 1024)
		n, err := res.Body.Read(p)
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		for _, b := range p[:n] {
			if !unicode.IsGraphic(rune(b)) {
				fmt.Fprintln(c.out, "Output contains binary data, not showing it to preserve terminal")
				return nil
			}
		}
		_, err = c.out.Write(p[:n])
		if err != nil {
			return fmt.Errorf("failed to write response body: %w", err)
		}
	}
	_, err := io.Copy(c.out, res.Body)
	return err
}
