package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jucrouzet/dsak/internal/pkg/commander"
)

func init() {
	commander.Register(
		"dns>servers",
		func() *cobra.Command {
			return &cobra.Command{
				Use:   "servers",
				Short: "Get or set DNS server aliases",
			}
		},
	)
}
