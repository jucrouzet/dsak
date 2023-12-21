// Package commander handles commands execution.
package commander

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/jucrouzet/dsak/internal/pkg/config"
)

func Run(args []string, forCommands ...map[string]registeredCommand) error {
	var list map[string]registeredCommand
	if len(forCommands) > 0 {
		list = forCommands[0]
	} else {
		list = commandsRegistered
	}
	rootCmd, cmds, err := buildCommandTree(list)
	if err != nil {
		return fmt.Errorf("failed building command tree: %w", err)
	}
	rootCmd.SetContext(context.Background())
	cfg, err := config.New(args)
	if err != nil {
		return fmt.Errorf("failed initializing configuration: %w", err)
	}
	config.SetCommandContext(rootCmd, cfg)
	if err := applyConfigs(list, cmds, cfg); err != nil {
		return fmt.Errorf("failed applying config to command tree: %w", err)
	}
	if err := applyFlagCompletion(list, cmds); err != nil {
		return fmt.Errorf("failed applying flag completion to command tree: %w", err)
	}
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

// CommandFlagCompletionFunc is the type of the completion function for a flag.
type CommandFlagCompletionFunc func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

type registeredCommand struct {
	creator       CommandCreator
	configs       []string
	flagCompleter map[string]CommandFlagCompletionFunc
}

var commandsRegistered = make(map[string]registeredCommand)

var commandNamePartValid = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]+$`).MatchString

// CommandCreator is a function that returns the registered cobra.Command.
type CommandCreator func() *cobra.Command

// CommandOption is a function that can be used to configure a registered command.
type CommandOption func(*registeredCommand) error

// Register registers a new command.
// `name` should be "" if the command is the root comand or be composed like this.
// "blah" reprensents a command named "blah" child of the root command.
// "blah>bleh" represents a command named "bleh" child of the command named "blah".
// configurations is the list of configuration values needed for the command.
// Register should be called in commands init() functions.
func Register(name string, creator CommandCreator, opts ...CommandOption) {
	if _, ok := commandsRegistered[name]; ok {
		panic(fmt.Errorf("command name is already registered: %s", name))
	}
	if name != "" {
		for _, part := range strings.Split(name, ">") {
			if part == "" || !commandNamePartValid(part) {
				panic(fmt.Errorf("invalid command name: %s", name))
			}
		}
	}
	cmd := registeredCommand{
		creator:       creator,
		flagCompleter: make(map[string]CommandFlagCompletionFunc),
	}
	for _, opt := range opts {
		if err := opt(&cmd); err != nil {
			panic(fmt.Errorf("invalid option for command %s: %w", name, err))
		}
	}
	commandsRegistered[name] = cmd
}

// WithConfig adds one or more element to the list of configuration values needed for the command.
func WithConfig(configs ...string) CommandOption {
	return func(c *registeredCommand) error {
		c.configs = append(c.configs, configs...)
		return nil
	}
}

// WithFlagCompletion sets a method to be called to get a command's flag completion.
func WithFlagCompletion(configName string, f CommandFlagCompletionFunc) CommandOption {
	return func(c *registeredCommand) error {
		if _, ok := c.flagCompleter[configName]; ok {
			return fmt.Errorf("flag completion already set for config %s", configName)
		}
		c.flagCompleter[configName] = f
		return nil
	}
}

func buildCommandTree(commands map[string]registeredCommand) (*cobra.Command, map[string]*cobra.Command, error) {
	rootCreator, ok := commands[""]
	if !ok {
		return nil, nil, fmt.Errorf("root command not found")
	}
	rootCmd := rootCreator.creator()

	leveled := make(map[int][]string)
	for name := range commands {
		if name == "" {
			continue
		}
		parts := strings.Split(name, ">")
		leveled[len(parts)] = append(leveled[len(parts)], name)
	}
	counts := make([]int, 0, len(leveled))
	for k := range leveled {
		counts = append(counts, k)
	}
	sort.Ints(counts)
	cmds := make(map[string]*cobra.Command)
	cmds[""] = rootCmd
	for _, level := range counts {
		for _, command := range leveled[level] {
			var parentName string
			if level == 1 {
				parentName = ""
			} else {
				parts := strings.Split(command, ">")
				parentName = strings.Join(parts[:len(parts)-1], ">")
			}
			parent, ok := cmds[parentName]
			if !ok {
				return nil, nil, fmt.Errorf("parent command of %s not found", command)
			}
			cmd := commands[command].creator()
			cmds[command] = cmd
			parent.AddCommand(cmd)
		}
	}
	return rootCmd, cmds, nil
}

func applyConfigs(
	commands map[string]registeredCommand,
	builtCommands map[string]*cobra.Command,
	cfg *viper.Viper,
) error {
	for name, cmd := range builtCommands {
		v := commands[name]
		for _, configName := range v.configs {
			value, err := config.GetValue(configName)
			if err != nil {
				return fmt.Errorf("failed getting configuration for command %s: %w", name, err)
			}
			if err := value.Apply(cmd, cfg); err != nil {
				return fmt.Errorf("failed applying configuration for command %s: %w", name, err)
			}
		}
	}
	return nil
}

func applyFlagCompletion(
	commands map[string]registeredCommand,
	builtCommands map[string]*cobra.Command,
) error {
	for name, cmd := range builtCommands {
		v := commands[name]
		for configName, f := range v.flagCompleter {
			value, err := config.GetValue(configName)
			if err != nil {
				return fmt.Errorf("failed getting configuration %s for flag completion of %s: %w", configName, name, err)
			}
			if err = cmd.RegisterFlagCompletionFunc(value.GetFlag(), f); err != nil {
				return fmt.Errorf("failed registering flag completion for configuration %s of %s: %w", configName, name, err)
			}
		}
	}
	return nil
}
