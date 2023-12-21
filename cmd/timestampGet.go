package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"zgo.at/tz"

	"github.com/jucrouzet/dsak/internal/pkg/commander"
	"github.com/jucrouzet/dsak/internal/pkg/config"
)

func init() {
	commander.Register(
		"timestamp>get",
		func() *cobra.Command {
			return &cobra.Command{
				Use:   "get [flags] [value]",
				Short: "Get a timestamp",
				Long: `Get a timestamp.

If no argument is specified, dsak will return current timestamp.
If a value is specified, dsak will try to parse the value as a date (see timestamp --format flag).`,
				RunE: func(cmd *cobra.Command, args []string) error {
					cfg := config.GetFromCommandContext(cmd)
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
					if len(args) == 0 {
						timestampGetPrint(cmd, time.Now())
						return nil
					}
					t, err := time.ParseInLocation(
						cfg.GetString(configKeyTimestampFormat),
						args[0],
						loc,
					)
					if err == nil {
						timestampGetPrint(cmd, t)
						return nil
					}
					return err
				},
			}
		},
	)
}

func timestampGetPrint(cmd *cobra.Command, t time.Time) {
	if config.GetFromCommandContext(cmd).GetBool(configKeyTimestampMSecs) {
		fmt.Println(t.UnixMilli())
		return
	}
	fmt.Println(t.Unix())
}
