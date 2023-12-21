package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"zgo.at/tz"

	"github.com/jucrouzet/dsak/internal/pkg/commander"
	"github.com/jucrouzet/dsak/internal/pkg/config"
)

func init() {
	commander.Register(
		"timestamp>parse",
		func() *cobra.Command {
			return &cobra.Command{
				Use:   "parse [flags] value",
				Short: "Parses a timestamp and print it into a human readable format",
				Args:  cobra.MinimumNArgs(1),
				Long: `Parses a timestamp and print it into a human readable format.

Value can be a unix timestamp, a unix timestamp with milliseconds, positive or negative.
To use a negative value, from the command line, use -- before value.
Eg: dsak timestamp -- -1609459
If - is specified as value, value is read from stdin.`,
				RunE: func(cmd *cobra.Command, args []string) error {
					arg, err := getStdinOrValue(cmd, args[0], true)
					if err != nil {
						return err
					}
					ts, err := strconv.ParseInt(arg, 10, 64)
					if err != nil {
						return fmt.Errorf("failed to parse timestamp value: %w", err)
					}
					cfg := config.GetFromCommandContext(cmd)
					var t time.Time
					if cfg.GetBool(configKeyTimestampMSecs) {
						t = time.UnixMilli(ts)
					} else {
						t = time.Unix(ts, 0)
					}
					loc := time.Local
					askedTZ := cfg.GetString(configKeyTimestampTimezone)
					if askedTZ != "" && askedTZ != loc.String() {
						var found *time.Location
						for _, v := range tz.Zones {
							if strings.EqualFold(v.Zone, askedTZ) {
								found = v.Location
								break
							}
						}
						if found == nil {
							return fmt.Errorf("unhandled timezone: %s", askedTZ)
						}
						loc = found
					}
					fmt.Println(t.In(loc).Format(cfg.GetString(configKeyTimestampFormat)))
					return nil
				},
			}
		},
	)
}
