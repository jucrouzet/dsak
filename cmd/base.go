package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/spf13/cobra"

	"github.com/jucrouzet/dsak/internal/pkg/commander"
	"github.com/jucrouzet/dsak/internal/pkg/config"
)

const (
	configKeyBaseInput = "base.input"
)

func init() {
	config.RegisterValue(
		configKeyBaseInput,
		config.ValueTypeString,
		config.DefaultValue("decimal"),
		config.Flag("base"),
		config.ShortFlag('b'),
		config.Description("Integer base for input"),
	)

	commander.Register(
		"base",
		func() *cobra.Command {
			return &cobra.Command{
				Use:     "base [flags] destination_base integer",
				Short:   "Transforms an integer from a base to another",
				Example: "base hex -- -42",
				Long: `Transforms an integer from a base to another.

Base can be binary, octal, decimal, or hexadecimal or any other positive integer value from 2 to 36.
"binary" or "bin" means binary.
"octal" or "oct"" means octal.
"decimal" or "dec" means decimal.
"hexadecimal", "hexa" or "hex" means hexadecimal.

Integers must be in the range of -9223372036854775808 to 9223372036854775807 (int64).
If integer is negative, prepend value with --.
They can be specified by just typing the value, like "123", which means 123 in decimal.
They can be specified with a base prefix :
- "0b" for binary : "0b101010" means 101010 in binary,
- "0o" for octal : "01234" means 1234 in octal,
- "0x" for hexadecimal : "0x1234" means 1234 in hexadecimal.

Alternatively, you can specify the integer base with the --base flag but if you do so,
integer should not be prefixed with "0b", "0o" or "0x".`,
				Args: cobra.ExactArgs(2),
				ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
					if len(args) == 0 {
						return getBaseCompletion(cmd, args, toComplete)
					}
					return nil, cobra.ShellCompDirectiveNoFileComp
				},
				RunE: func(cmd *cobra.Command, args []string) error {
					cfg := config.GetFromCommandContext(cmd)
					givenBase, err := getBase(cfg.GetString(configKeyBaseInput))
					if err != nil {
						return fmt.Errorf("invalid input base: %w", err)
					}
					destBase, err := getBase(args[0])
					if err != nil {
						return fmt.Errorf("invalid destination base: %w", err)
					}
					input, err := parseInput(args[1], givenBase)
					if err != nil {
						return fmt.Errorf("cannot parse integer: %w", err)
					}
					_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", strconv.FormatInt(input, destBase))
					return err
				},
			}
		},
		commander.WithConfig(configKeyBaseInput),
		commander.WithFlagCompletion(configKeyBaseInput, getBaseCompletion),
	)
}

func getBaseCompletion(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	baseList := []string{"binary", "octal", "decimal", "hexadecimal"}
	list := make([]string, 0, len(baseList))
	for _, base := range baseList {
		if toComplete == "" || strings.HasPrefix(base, toComplete) {
			list = append(list, base)
		}
	}
	return list, cobra.ShellCompDirectiveNoFileComp
}

func getBase(base string) (int, error) {
	switch strings.ToLower(base) {
	case "binary", "bin":
		return 2, nil
	case "octal", "oct":
		return 8, nil
	case "decimal", "dec":
		return 10, nil
	case "hexadecimal", "hexa", "hex":
		return 16, nil
	default:
		for _, c := range base {
			if !unicode.IsDigit(c) {
				return 0, fmt.Errorf("invalid base: %s", base)
			}
		}
		v, err := strconv.ParseInt(base, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid base: %s", base)
		}
		if v < 2 || v > 36 {
			return 0, errors.New("base must be in the range of 2 to 36")
		}
		return int(v), nil
	}
}

func parseInput(input string, givenBase int) (int64, error) {
	input = strings.TrimSpace(strings.ToLower(input))
	base := givenBase
	if givenBase == 10 && len(input) >= 3 {
		if strings.HasPrefix(input, "0b") {
			base = 2
			input = input[2:]
		}
		if strings.HasPrefix(input, "0o") {
			base = 8
			input = input[2:]
		}
		if strings.HasPrefix(input, "0x") {
			base = 16
			input = input[2:]
		}
	}
	return strconv.ParseInt(input, base, 64)
}
