package config

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type cmdContextConfigKeyType string

var cmdContextConfigKey = cmdContextConfigKeyType("configuration")

// SetCommandContext sets the command context to contain a new configuration.
func SetCommandContext(cmd *cobra.Command, config *viper.Viper) {
	cmd.SetContext(context.WithValue(cmd.Context(), cmdContextConfigKey, config))
}

// GetFromCommandContext gets the configuration from a command context.
// If no configuration is found, a new configuration is created.
func GetFromCommandContext(cmd *cobra.Command) *viper.Viper {
	cfg, ok := cmd.Context().Value(cmdContextConfigKey).(*viper.Viper)
	if ok {
		return cfg
	}
	return viper.New()
}
