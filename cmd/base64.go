package cmd

import (
	"encoding/base64"

	"github.com/spf13/cobra"

	"github.com/jucrouzet/dsak/internal/pkg/commander"
	"github.com/jucrouzet/dsak/internal/pkg/config"
)

const (
	configKeyBase64URLEncoding    = "base64.urlencoding"
	configKeyBase64DisablePadding = "base64.disablepadding"
)

func init() {
	config.RegisterValue(
		configKeyBase64URLEncoding,
		config.ValueTypeBool,
		config.Description("Use URL encoding mode"),
		config.Flag("url-encoding"),
		config.ShortFlag('u'),
		config.FlagIsPersistent(),
	)

	config.RegisterValue(
		configKeyBase64DisablePadding,
		config.ValueTypeBool,
		config.Description("Disable padding"),
		config.Flag("disable-padding"),
		config.ShortFlag('p'),
		config.FlagIsPersistent(),
	)

	commander.Register(
		"base64",
		func() *cobra.Command {
			return &cobra.Command{
				Use:     "base64",
				Short:   "Base64 tools",
				Aliases: []string{"b64"},
			}
		},
		commander.WithConfig(configKeyBase64URLEncoding),
		commander.WithConfig(configKeyBase64DisablePadding),
	)
}

func getBase64Encoding(cmd *cobra.Command) *base64.Encoding {
	cfg := config.GetFromCommandContext(cmd)
	url := cfg.GetBool(configKeyBase64URLEncoding)
	padding := cfg.GetBool(configKeyBase64DisablePadding)

	if url && padding {
		return base64.RawURLEncoding
	} else if padding {
		return base64.RawStdEncoding
	} else if url {
		return base64.URLEncoding
	}
	return base64.StdEncoding
}
