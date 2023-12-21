package cmd

import (
	"errors"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jucrouzet/dsak/internal/pkg/commander"
	"github.com/jucrouzet/dsak/internal/pkg/config"
)

func init() {
	commander.Register(
		"dns>servers>rm",
		func() *cobra.Command {
			return &cobra.Command{
				Use:     "rm [flags] alias [ip_or_hostname...]",
				Short:   "Remove one or more IPs or hostname from a DNS server alias or delete the alias",
				Aliases: []string{"remove"},
				Example: "rm lan 192.168.1.1 192.168.2.1",
				Args:    cobra.MinimumNArgs(1),
				ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
					if len(args) == 0 {
						return getDNSServerAliasCompletion(cmd, toComplete)
					}
					cfg := config.GetFromCommandContext(cmd)
					aliases := cfg.GetStringMapStringSlice(configKeyDNSServerAliases)
					list, ok := aliases[args[0]]
					if !ok {
						return nil, cobra.ShellCompDirectiveNoFileComp
					}
					if len(toComplete) > 0 {
						toComplete = strings.ToLower(toComplete)
						newList := make([]string, 0, len(list))
						for _, addr := range list {
							if strings.Contains(strings.ToLower(addr), toComplete) {
								newList = append(newList, addr)
							}
						}
						list = newList
					}
					if len(args) > 1 {
						newList := make([]string, 0, len(list))
						for _, addr := range list {
							if !slices.Contains(args[1:], addr) {
								newList = append(newList, addr)
							}
						}
						list = newList
					}
					return list, cobra.ShellCompDirectiveNoFileComp
				},
				RunE: func(cmd *cobra.Command, args []string) error {
					if len(args) == 1 {
						return dnsServersRmAlias(cmd, args[0])
					}
					return dnsServersRmEntries(cmd, args[0], args[1:]...)
				},
			}
		},
	)
}

func dnsServersRmAlias(cmd *cobra.Command, alias string) error {
	if alias == "default" {
		return errors.New("cannot remove the default DNS server alias")
	}
	cfg := config.GetFromCommandContext(cmd)
	aliases := cfg.GetStringMapStringSlice(configKeyDNSServerAliases)
	if _, ok := aliases[alias]; !ok {
		return nil
	}
	delete(aliases, alias)
	cfg.Set(configKeyDNSServerAliases, aliases)
	return config.Write(cfg)
}

func dnsServersRmEntries(cmd *cobra.Command, alias string, entries ...string) error {
	cfg := config.GetFromCommandContext(cmd)
	aliases := cfg.GetStringMapStringSlice(configKeyDNSServerAliases)
	list, ok := aliases[alias]
	if !ok {
		return nil
	}
	newList := make([]string, 0, len(list))
	for _, addr := range list {
		if !slices.Contains(entries, addr) {
			newList = append(newList, addr)
		}
	}
	if len(newList) == 0 {
		if alias == "default" {
			return errors.New("the default DNS server alias must have at least one entry")
		}
		delete(aliases, alias)
	} else {
		aliases[alias] = newList
	}
	cfg.Set(configKeyDNSServerAliases, aliases)
	return config.Write(cfg)
}
