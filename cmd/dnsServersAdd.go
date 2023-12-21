package cmd

import (
	"slices"

	"github.com/spf13/cobra"

	"github.com/jucrouzet/dsak/internal/pkg/commander"
	"github.com/jucrouzet/dsak/internal/pkg/config"
)

func init() {
	commander.Register(
		"dns>servers>add",
		func() *cobra.Command {
			return &cobra.Command{
				Use:     "add [flags] alias ip_or_hostname...",
				Short:   "Add one or more IPs or hostname to a DNS server alias",
				Example: "add lan 192.168.1.1 192.168.2.1",
				Args:    cobra.MinimumNArgs(2),
				ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
					if len(args) == 0 {
						return getDNSServerAliasCompletion(cmd, toComplete)
					}
					return nil, cobra.ShellCompDirectiveNoFileComp
				},
				RunE: func(cmd *cobra.Command, args []string) error {
					cfg := config.GetFromCommandContext(cmd)
					aliases := cfg.GetStringMapStringSlice(configKeyDNSServerAliases)
					list, ok := aliases[args[0]]
					if !ok {
						list = make([]string, 0)
					}
					added := 0
					for _, addr := range args[1:] {
						if !slices.Contains(list, addr) {
							list = append(list, addr)
							added++
						}
					}
					if added == 0 {
						return nil
					}
					aliases[args[0]] = list
					cfg.Set(configKeyDNSServerAliases, aliases)
					return config.Write(cfg)
				},
			}
		},
	)
}
