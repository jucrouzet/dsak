package cmd

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jucrouzet/dsak/internal/pkg/commander"
	"github.com/jucrouzet/dsak/internal/pkg/config"
	"github.com/jucrouzet/dsak/internal/pkg/dns"
)

const (
	configKeyDNSQueryUseServers = "dns.query.useservers"
	configKeyDNSQueryType       = "dns.query.type"
)

func init() {
	config.RegisterValue(
		configKeyDNSQueryType,
		config.ValueTypeString,
		config.Flag("type"),
		config.ShortFlag('t'),
		config.DefaultValue("A"),
		config.Description("Record type to query"),
	)
	config.RegisterValue(
		configKeyDNSQueryUseServers,
		config.ValueTypeStrings,
		config.Flag("servers"),
		config.ShortFlag('s'),
		config.DefaultValue([]string{"default"}),
		config.Description("DNS server aliases or IP addresses/hostnames to use"),
	)

	commander.Register(
		"dns>query",
		func() *cobra.Command {
			return &cobra.Command{
				Use:   "query [flags] value",
				Short: "Run a dns query",
				Args:  cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					servers := dnsQueryGetServers(cmd)
					cfg := config.GetFromCommandContext(cmd)
					tName := cfg.GetString(configKeyDNSQueryType)
					t, err := dns.GetType(tName)
					if err != nil {
						return err
					}
					client := dns.NewClient(getLogger(cmd), servers...)
					res, err := client.Query(cmd.Context(), t, args[0])
					if err != nil {
						return fmt.Errorf("failed to query: %w", err)
					}
					fmt.Fprintln(cmd.OutOrStdout(), res)
					return nil
				},
			}
		},
		commander.WithConfig(configKeyDNSQueryUseServers),
		commander.WithConfig(configKeyDNSQueryType),
		commander.WithFlagCompletion(
			configKeyDNSQueryUseServers,
			func(cmd *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
			},
		),
		commander.WithFlagCompletion(
			configKeyDNSQueryType,
			func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				flags := cobra.ShellCompDirectiveNoFileComp
				var completions []string
				toComplete = strings.ToLower(toComplete)
				for _, v := range dns.GetTypeNames() {
					t, err := dns.GetType(v)
					if err != nil {
						continue
					}
					vv := strings.ToLower(v)
					if toComplete == "" || strings.Contains(vv, toComplete) {
						completions = append(completions, fmt.Sprintf("%s\t%s", v, dns.GetTypeDescription(t)))
					}
				}
				sort.Strings(completions)
				return completions, flags
			},
		),
	)
}

func dnsQueryGetServers(cmd *cobra.Command) []string {
	cfg := config.GetFromCommandContext(cmd)
	aliases := cfg.GetStringMapStringSlice(configKeyDNSServerAliases)
	list := make([]string, 0)

	for _, server := range cfg.GetStringSlice(configKeyDNSQueryUseServers) {
		aliasList, ok := aliases[server]
		if ok {
			for _, addr := range aliasList {
				if !slices.Contains(list, addr) {
					list = append(list, addr)
				}
			}
			continue
		}
		if !slices.Contains(list, server) {
			list = append(list, server)
		}
	}
	return list
}
