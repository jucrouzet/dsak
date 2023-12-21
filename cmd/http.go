package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jucrouzet/dsak/internal/pkg/commander"
)

func init() {
	commander.Register(
		"http",
		func() *cobra.Command {
			return &cobra.Command{
				Use:   "http",
				Short: "HTTP Tools",
			}
		},
	)
}
