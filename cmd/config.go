package cmd

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/jucrouzet/dsak/internal/pkg/commander"
	"github.com/jucrouzet/dsak/internal/pkg/config"
)

const (
	configKeyConfigRaw = "config.printraw"
)

func init() {
	config.RegisterValue(
		configKeyConfigRaw,
		config.ValueTypeBool,
		config.Flag("raw"),
		config.ShortFlag('r'),
		config.DefaultValue(false),
		config.Description("Show raw configuration value. Needs a config name to be used."),
	)
	commander.Register(
		"config",
		func() *cobra.Command {
			return &cobra.Command{
				Use:   "config [flags] [name] [value]",
				Short: "Get or set a configuration value",
				ValidArgsFunction: func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
					var commands []string
					for _, v := range config.GetValues() {
						commands = append(commands, v.GetName())
					}
					var list []string
					if len(args) == 0 {
						if len(toComplete) > 0 {
							toComplete = strings.ToLower(toComplete)
							for _, name := range commands {
								if strings.Contains(strings.ToLower(name), toComplete) {
									list = append(list, name)
								}
							}
						} else {
							list = commands
						}
					}
					return list, cobra.ShellCompDirectiveNoFileComp
				},
				RunE: func(cmd *cobra.Command, args []string) error {
					cfg := config.GetFromCommandContext(cmd)
					if len(args) == 0 {
						return configPrintAllValues(cmd)
					} else if len(args) == 1 {
						v, err := config.GetValue(args[0])
						if err != nil {
							return err
						}
						return configPrintValue(cmd, v)
					}
					v, err := config.GetValue(args[0])
					if err != nil {
						return err
					}
					if err := v.Set(cmd, args[1]); err != nil {
						return fmt.Errorf("failed to set config value: %w", err)
					}
					if err := config.Write(cfg); err != nil {
						return fmt.Errorf("failed to save config file: %w", err)
					}
					return configPrintValue(cmd, v)
				},
			}
		},
		commander.WithConfig(configKeyConfigRaw),
	)
}

func configPrintAllValues(cmd *cobra.Command) error {
	logger := getLogger(cmd)
	cfg := config.GetFromCommandContext(cmd)
	if cfg.GetBool(configKeyConfigRaw) {
		return errors.New("Raw value is only supported with a config name")
	}
	values := config.GetValues()
	keys := make([]string, 0, len(values))
	for _, v := range values {
		keys = append(keys, v.GetName())
	}
	sort.Strings(keys)
	for _, k := range keys {
		v, err := config.GetValue(k)
		if err != nil {
			continue
		}
		if err := configPrintValue(cmd, v); err != nil {
			logger.
				With(zap.Error(err)).
				With(zap.String("value", v.GetName())).
				Warn("failed to print config value")
		}
	}
	return nil
}

func configPrintValue(cmd *cobra.Command, v *config.Value) error {
	cfg := config.GetFromCommandContext(cmd)
	raw := cfg.GetBool(configKeyConfigRaw)
	if raw {
		fmt.Fprintln(cmd.OutOrStdout(), v.AsRawString(cmd))
	} else {
		name := color.New(color.FgBlue)
		value := color.New(color.Bold, color.FgGreen)
		name.Fprintf(cmd.OutOrStdout(), "[%s]:\n", v.GetName())
		res := v.AsString(cmd)
		res = strings.Trim(strings.Join(strings.Split(res, "\n"), "\n\t"), "\n\t")
		value.Fprintf(cmd.OutOrStdout(), "\t%s\n", res)
	}
	return nil
}
