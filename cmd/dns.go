package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/jucrouzet/dsak/internal/pkg/commander"
	"github.com/jucrouzet/dsak/internal/pkg/config"
)

const (
	configKeyDNSServerAliases = "dns.serveraliases"
)

func init() {
	config.RegisterValue(
		configKeyDNSServerAliases,
		config.ValueTypeStringsMap,
		config.DefaultValue(map[string][]string{
			"default": {
				"1.1.1.1",
				"8.8.8.8",
			},
		}),
		config.Description("List of DNS server aliases"),
	)
	commander.Register(
		"dns",
		func() *cobra.Command {
			return &cobra.Command{
				Use:   "dns",
				Short: "DNS Tools",
			}
		},
		commander.WithConfig(configKeyDNSServerAliases),
	)
}

func getDNSServerAliasCompletion(cmd *cobra.Command, toComplete string) ([]string, cobra.ShellCompDirective) {
	cfg := config.GetFromCommandContext(cmd)
	aliases := cfg.GetStringMapStringSlice(configKeyDNSServerAliases)
	flags := cobra.ShellCompDirectiveNoFileComp

	toComplete = strings.ToLower(toComplete)
	completions := make([]string, 0, len(aliases))
	for v := range aliases {
		v = strings.ToLower(v)
		if toComplete == "" || strings.Contains(v, toComplete) {
			completions = append(completions, v)
		}
	}
	return completions, flags
}
