package cmd

import (
	"encoding/base64"
	"io"

	"github.com/spf13/cobra"

	"github.com/jucrouzet/dsak/internal/pkg/commander"
	"github.com/jucrouzet/dsak/internal/pkg/resource"
)

func init() {
	commander.Register(
		"base64>encode",
		func() *cobra.Command {
			return &cobra.Command{
				Use:   "encode [flags] resource",
				Short: "Encode a resource with base64",
				Args:  cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					in, err := resource.New(cmd, args[0], getLogger(cmd))
					if err != nil {
						return err
					}
					defer in.Close()
					enc := base64.NewEncoder(getBase64Encoding(cmd), cmd.OutOrStdout())
					_, err = io.Copy(enc, in)
					return err
				},
			}
		},
	)
}
