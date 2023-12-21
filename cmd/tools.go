package cmd

import (
	"bytes"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func getStdinOrValue(cmd *cobra.Command, v string, trimLines ...bool) (string, error) {
	if v == "-" {
		b, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return "", fmt.Errorf("failed to read from stdin: %w", err)
		}
		if len(trimLines) > 0 && trimLines[0] {
			b = bytes.TrimRight(b, "\n")
		}
		return string(b), nil
	}
	return v, nil
}
