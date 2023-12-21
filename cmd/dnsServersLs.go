package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/jucrouzet/dsak/internal/pkg/commander"
	"github.com/jucrouzet/dsak/internal/pkg/config"
)

func init() {
	commander.Register(
		"dns>servers>ls",
		func() *cobra.Command {
			return &cobra.Command{
				Use:     "ls [flags] [alias]",
				Short:   "List a specific or all DNS server alias",
				Aliases: []string{"list"},
				ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
					if len(args) == 0 {
						return getDNSServerAliasCompletion(cmd, toComplete)
					}
					return nil, cobra.ShellCompDirectiveNoFileComp
				},
				Run: func(cmd *cobra.Command, args []string) {
					name := color.New(color.FgBlue)
					value := color.New(color.Bold, color.FgGreen)
					cfg := config.GetFromCommandContext(cmd)
					aliases := cfg.GetStringMapStringSlice(configKeyDNSServerAliases)
					for alias, list := range aliases {
						if len(args) == 0 || alias == args[0] {
							name.Fprintf(cmd.OutOrStdout(), "[%s]:\n", alias)
							for _, addr := range list {
								value.Fprintf(cmd.OutOrStdout(), "  - %s\n", addr)
							}
						}
					}
				},
			}
		},
	)
}
