package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"zgo.at/tz"

	"github.com/jucrouzet/dsak/internal/pkg/commander"
	"github.com/jucrouzet/dsak/internal/pkg/config"
)

const (
	configKeyTimestampTimezone = "timestamp.timezone"
	configKeyTimestampMSecs    = "timestamp.msecs"
	configKeyTimestampFormat   = "timestamp.format"
)

func init() {
	config.RegisterValue(
		configKeyTimestampTimezone,
		config.ValueTypeString,
		config.DefaultValue(time.Local.String()),
		config.Flag("timezone"),
		config.ShortFlag('z'),
		config.FlagIsPersistent(),
		config.Description("Use this timezone"),
	)
	config.RegisterValue(
		configKeyTimestampMSecs,
		config.ValueTypeBool,
		config.Flag("msecs"),
		config.ShortFlag('m'),
		config.FlagIsPersistent(),
		config.Description("Use milliseconds instead of seconds for timestamps"),
	)
	config.RegisterValue(
		configKeyTimestampFormat,
		config.ValueTypeString,
		config.DefaultValue("2006-01-02 15:04:05.000"),
		config.Flag("format"),
		config.ShortFlag('f'),
		config.FlagIsPersistent(),
		config.Description("Use this golang time.Format layout string"),
	)
	commander.Register(
		"timestamp",
		func() *cobra.Command {
			return &cobra.Command{
				Use:     "timestamp",
				Aliases: []string{"ts"},
				Short:   "Timestamp tools",
				Long: `Timestamp tools.

The --msecs flag tells that timestamp values are interpreted as milliseconds since epoch.
The --timezone flag tells to use a timezone other than the machine's default.
	Timezones flag has completion.
The --format flag tells to use a golang time.Format layout string for date results and parsing.
	See https://go.dev/src/time/format.go`,
				RunE: func(cmd *cobra.Command, args []string) error {
					return cmd.Usage()
				},
			}
		},
		commander.WithConfig(configKeyTimestampTimezone),
		commander.WithFlagCompletion(
			configKeyTimestampTimezone,
			func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				flags := cobra.ShellCompDirectiveNoFileComp
				list := make([]string, 0, len(tz.Zones))
				toComplete = strings.ToLower(toComplete)
				for _, v := range tz.Zones {
					if toComplete == "" || strings.Contains(strings.ToLower(v.Zone), toComplete) {
						list = append(list, fmt.Sprintf("%s\t%s", v.Zone, v.Display()))
					}
				}
				sort.Strings(list)
				return list, flags
			},
		),
		commander.WithConfig(configKeyTimestampMSecs),
		commander.WithConfig(configKeyTimestampFormat),
		commander.WithFlagCompletion(
			configKeyTimestampFormat,
			func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
				flags := cobra.ShellCompDirectiveNoFileComp
				formats := [][2]string{
					{time.ANSIC, "ANSIC format"},
					{time.UnixDate, "Unix date"},
					{time.RubyDate, "Ruby date"},
					{time.RFC822, "RFC822"},
					{time.RFC822Z, "RFC822Z"},
					{time.RFC850, "RFC850"},
					{time.RFC1123, "RFC1123"},
					{time.RFC1123Z, "RFC1123Z"},
					{time.RFC3339, "RFC3339"},
					{time.Stamp, "Timestamp"},
					{time.StampMilli, "Timestamp with milliseconds"},
					{time.DateTime, "DateTime"},
				}
				keys := make([]string, 0, len(formats))
				for _, f := range formats {
					keys = append(keys, f[1])
				}
				sort.Strings(keys)
				list := make([]string, 0, len(keys))
				for _, k := range keys {
					for _, f := range formats {
						if k == f[1] {
							list = append(list, fmt.Sprintf("%s\t%s", f[0], f[1]))
							break
						}
					}
				}
				return list, flags
			},
		),
	)
}

func init() {
	for _, z := range tz.Zones {
		var err error
		z.Location, err = time.LoadLocation(z.Zone)
		if err != nil {
			if strings.Contains(err.Error(), "unknown time zone") {
				fmt.Fprintf(os.Stderr, "warning: timezone.init: %s; you probably need to update your tzdata/zoneinfo\n", err)
			}
		}
	}
}
