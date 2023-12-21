// Package config contains the configuration handling for dsak.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// New creates parses configuration file and return a new configuration instance.
func New(args []string) (*viper.Viper, error) {
	configFile, err := getConfigFile(args)
	if err != nil {
		return nil, err
	}
	config := viper.New()

	config.SetConfigFile(configFile)
	config.SetConfigType("yaml")

	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		return config, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat configuration file %s: %w", configFile, err)
	}

	if err := config.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read configuration file %s: %w", configFile, err)
	}
	return config, nil
}

// Write writes the given configuration to file.
func Write(cfg *viper.Viper) error {
	if cfg.ConfigFileUsed() == "" {
		return fmt.Errorf("no configuration file specified")
	}
	_, err := os.Stat(cfg.ConfigFileUsed())
	if os.IsNotExist(err) {
		f, err := os.Create(cfg.ConfigFileUsed())
		if err != nil {
			return fmt.Errorf("failed creating configuration file: %w", err)
		}
		f.Close()
	}
	if err := cfg.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}
	return nil
}

func getDefaultConfigFile() (string, error) {
	defaultConfigFile := os.Getenv("DSAK_CONFIGFILE")
	if defaultConfigFile == "" {
		dirname, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		defaultConfigFile = filepath.Join(dirname, ".dsak.yaml")
	}
	return defaultConfigFile, nil
}

func getConfigFile(args []string) (string, error) {
	file, err := getDefaultConfigFile()
	if err != nil {
		return "", err
	}
	for i, arg := range args {
		if arg == "--configfile" {
			if i == len(args)-1 {
				return "", fmt.Errorf("--configfile requires an argument")
			}
			file = args[i+1]
		}
	}
	return file, nil
}
