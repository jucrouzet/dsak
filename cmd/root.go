package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/hako/durafmt"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/term"

	"github.com/jucrouzet/dsak/internal/pkg/commander"
	"github.com/jucrouzet/dsak/internal/pkg/config"
	"github.com/jucrouzet/dsak/internal/pkg/resource"
)

const (
	configKeyGlobalJSONLogs = "global.jsonlogs"
	configKeyGlobalTimeout  = "global.timeout"
	configKeyGlobalVerbose  = "global.verbose"
	configKeyGlobalOutput   = "global.output"
	configKeyGlobalNoColor  = "global.nocolor"
)

func init() {
	config.RegisterValue(
		configKeyGlobalTimeout,
		config.ValueTypeUint,
		config.DefaultValue(uint64(10000)),
		config.Flag("timeout"),
		config.FlagIsPersistent(),
		config.Description("Timeout for command in milliseconds, 0 for unlimited"),
		config.Stringer(func(cmd *cobra.Command) string {
			cfg := config.GetFromCommandContext(cmd)
			msecs := cfg.GetUint64(configKeyGlobalTimeout)
			if msecs == 0 {
				return "unlimited"
			}
			return durafmt.Parse(time.Duration(msecs) * time.Millisecond).String()
		}),
	)

	config.RegisterValue(
		configKeyGlobalVerbose,
		config.ValueTypeBool,
		config.DefaultValue(false),
		config.Flag("verbose"),
		config.FlagIsPersistent(),
		config.Description("Run command verbosely"),
	)

	config.RegisterValue(
		configKeyGlobalJSONLogs,
		config.ValueTypeBool,
		config.DefaultValue(false),
		config.Flag("jsonlogs"),
		config.FlagIsPersistent(),
		config.Description("Log output in JSON format"),
	)

	config.RegisterValue(
		configKeyGlobalOutput,
		config.ValueTypeString,
		config.DefaultValue("stdout"),
		config.Flag("output"),
		config.FlagIsPersistent(),
		config.Description("Command output resource"),
	)

	config.RegisterValue(
		configKeyGlobalNoColor,
		config.ValueTypeBool,
		config.Flag("no-color"),
		config.FlagIsPersistent(),
		config.Description("Diable color in output"),
	)

	commander.Register(
		"",
		func() *cobra.Command {
			return &cobra.Command{
				Use:   "dsak",
				Short: "Dave's Swiss Army Knife",
				Long: `Dave's a developer.
Dave needs tools to work.
Dave loves cli.
This is Dave's Swiss Army Knife.
All the tools he needs in one cli.

When a command or a flag expects a resource, the resource can be stdin, stdout, stderr, a file or an URL.

You can use dsak command -h to get information about a command or its flags.`,

				PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
					return errors.Join(
						loggerInitializer(cmd),
						outputInitializer(cmd),
						timeoutInitializer(cmd),
					)
				},
				PersistentPostRun: func(cmd *cobra.Command, _ []string) {
					cancelTimeout(cmd)
				},
				RunE: func(cmd *cobra.Command, _ []string) error {
					return cmd.Usage()
				},
				PersistentPostRunE: func(cmd *cobra.Command, _ []string) error {
					closer, ok := cmd.OutOrStdout().(io.Closer)
					if ok {
						if err := closer.Close(); err != nil {
							return fmt.Errorf("failed to close output: %w", err)
						}
					}
					return nil
				},
			}
		},
		commander.WithConfig(configKeyGlobalJSONLogs),
		commander.WithConfig(configKeyGlobalTimeout),
		commander.WithConfig(configKeyGlobalVerbose),
		commander.WithConfig(configKeyGlobalOutput),
		commander.WithConfig(configKeyGlobalNoColor),
	)
}

type cmdContextLoggerKeyType string

var cmdContextLoggerKey = cmdContextLoggerKeyType("logger")

func loggerInitializer(cmd *cobra.Command) error {
	var zapCfg zap.Config
	cfg := config.GetFromCommandContext(cmd)

	verbose := cfg.GetBool(configKeyGlobalVerbose)
	encoding := "console"
	encodingConfig := zap.NewDevelopmentEncoderConfig()
	level := zap.InfoLevel
	if verbose {
		level = zap.DebugLevel
	}

	if cfg.GetBool(configKeyGlobalJSONLogs) {
		encoding = "json"
		encodingConfig = zap.NewProductionEncoderConfig()
	}
	zapCfg = zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      verbose,
		Encoding:         encoding,
		EncoderConfig:    encodingConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
	logger, err := zapCfg.Build()
	if err != nil {
		return fmt.Errorf("failed to build logger: %w", err)
	}
	logger = logger.With(zap.String("app", "dsak"))
	ctx := context.WithValue(cmd.Context(), cmdContextLoggerKey, logger)
	cmd.SetContext(ctx)
	return nil
}

func getLogger(cmd *cobra.Command) *zap.Logger {
	return getLoggerFromContext(cmd.Context())
}

func getLoggerFromContext(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(cmdContextLoggerKey).(*zap.Logger)
	if !ok {
		return zap.NewNop()
	}
	return logger
}

func outputInitializer(cmd *cobra.Command) error {
	cfg := config.GetFromCommandContext(cmd)
	if cfg.GetBool(configKeyGlobalNoColor) || !term.IsTerminal(syscall.Stdout) {
		color.NoColor = true
	}
	switch cfg.GetString(configKeyGlobalOutput) {
	case "stdout", "stderr":
	default:
		color.NoColor = true
	}
	out, err := resource.New(cmd, cfg.GetString(configKeyGlobalOutput), getLogger(cmd))
	if err != nil {
		return fmt.Errorf("cannot create output resource: %w", err)
	}
	cmd.SetOut(out)
	return nil
}

type cmdContextTimeoutCancelKeyType string

var cmdContextTimeoutCancel = cmdContextTimeoutCancelKeyType("timeout cancel")

func timeoutInitializer(cmd *cobra.Command) error { //nolint:unparam
	cfg := config.GetFromCommandContext(cmd)
	timeoutMS := cfg.GetUint16(configKeyGlobalTimeout)
	var timeout time.Duration
	if timeoutMS == 0 {
		timeout = 86400 * time.Hour
	} else {
		timeout = time.Duration(timeoutMS) * time.Millisecond
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
	ctx = context.WithValue(ctx, cmdContextTimeoutCancel, cancel)
	cmd.SetContext(ctx)
	return nil
}

func cancelTimeout(cmd *cobra.Command) {
	cancel, ok := cmd.Context().Value(cmdContextTimeoutCancel).(context.CancelFunc)
	if ok {
		cancel()
	}
}
