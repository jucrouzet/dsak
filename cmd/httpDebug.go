package cmd

import (
	"fmt"
	"io"
	"net/http"
	"slices"
	"sort"
	"strings"

	// Image formats.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/alecthomas/chroma/styles"
	"github.com/itchyny/gojq"
	"github.com/spf13/cobra"

	"github.com/jucrouzet/dsak/internal/pkg/commander"
	"github.com/jucrouzet/dsak/internal/pkg/config"
	"github.com/jucrouzet/dsak/internal/pkg/contenttype"
	"github.com/jucrouzet/dsak/internal/pkg/httpdsak"
	"github.com/jucrouzet/dsak/internal/pkg/resource"
)

const (
	configKeyHTTPDebugInsecure           = "http.debug.insecure"
	configKeyHTTPDebugMethod             = "http.debug.method"
	configKeyHTTPDebugRequestAccept      = "http.debug.request.accept"
	configKeyHTTPDebugRequestBody        = "http.debug.request.body"
	configKeyHTTPDebugRequestBodyContent = "http.debug.request.bodycontent"
	configKeyHTTPDebugRequestContentType = "http.debug.request.contenttype"
	configKeyHTTPDebugRequestForceHTTP1  = "http.debug.request.forcehttp1"
	configKeyHTTPDebugRequestForceHTTP2  = "http.debug.request.forcehttp2"
	configKeyHTTPDebugRequestHeader      = "http.debug.request.header"
	configKeyHTTPDebugResponseForceType  = "http.debug.response.forcetype"
	configKeyHTTPDebugResponseJQ         = "http.debug.response.jq"
	configKeyHTTPDebugResponseRaw        = "http.debug.response.raw"
	configKeyHTTPDebugResponseStyle      = "http.debug.response.style"
	configKeyHTTPDebugTrace              = "http.debug.trace"
)

func init() { //nolint:gocyclo
	config.RegisterValue(
		configKeyHTTPDebugInsecure,
		config.ValueTypeBool,
		config.Flag("insecure"),
		config.ShortFlag('k'),
		config.Description("Run in insecure mode, do not verify TLS certificates"),
	)

	config.RegisterValue(
		configKeyHTTPDebugMethod,
		config.ValueTypeString,
		config.DefaultValue("GET"),
		config.Flag("method"),
		config.ShortFlag('m'),
		config.Description("HTTP method to use"),
	)

	config.RegisterValue(
		configKeyHTTPDebugRequestAccept,
		config.ValueTypeString,
		config.Flag("accept"),
		config.ShortFlag('a'),
		config.Description("Set request's accept header value"),
	)

	config.RegisterValue(
		configKeyHTTPDebugRequestBody,
		config.ValueTypeString,
		config.Flag("request-body"),
		config.ShortFlag('b'),
		config.Description("Use this resource as the body of the HTTP request"),
	)

	config.RegisterValue(
		configKeyHTTPDebugRequestBodyContent,
		config.ValueTypeString,
		config.Flag("request-body-content"),
		config.ShortFlag('d'),
		config.Description("Use given string as the body of the HTTP request, if set, the --request-body flag is ignored"),
	)

	config.RegisterValue(
		configKeyHTTPDebugRequestContentType,
		config.ValueTypeString,
		config.DefaultValue("application/octet-stream"),
		config.Flag("request-content-type"),
		config.ShortFlag('c'),
		config.Description("Set request's content type header value"),
	)

	config.RegisterValue(
		configKeyHTTPDebugRequestForceHTTP1,
		config.ValueTypeBool,
		config.Flag("force-http1"),
		config.ShortFlag('1'),
		config.Description("Force request to use HTTP/1.1"),
	)

	config.RegisterValue(
		configKeyHTTPDebugRequestForceHTTP2,
		config.ValueTypeBool,
		config.Flag("force-http2"),
		config.ShortFlag('2'),
		config.Description("Force request to use HTTP/2"),
	)

	config.RegisterValue(
		configKeyHTTPDebugRequestHeader,
		config.ValueTypeStrings,
		config.Flag("header"),
		config.ShortFlag('H'),
		config.Description("Add header to the request"),
	)

	config.RegisterValue(
		configKeyHTTPDebugResponseForceType,
		config.ValueTypeString,
		config.Flag("force-type"),
		config.ShortFlag('T'),
		config.Description("Force response to use given content-type"),
	)

	config.RegisterValue(
		configKeyHTTPDebugResponseJQ,
		config.ValueTypeString,
		config.Flag("jq"),
		config.ShortFlag('j'),
		config.Description("Parse response body as json and apply given jq expression"),
	)

	config.RegisterValue(
		configKeyHTTPDebugResponseRaw,
		config.ValueTypeBool,
		config.Flag("raw-response"),
		config.ShortFlag('r'),
		config.Description("Get raw response body, do not beautify it"),
		config.Setter(func(_ *cobra.Command, val string) error {
			if val == "" {
				return nil
			}
			_, err := gojq.Parse(val)
			if err != nil {
				return fmt.Errorf("invalid jq filter: %w", err)
			}
			return nil
		}),
	)

	config.RegisterValue(
		configKeyHTTPDebugResponseStyle,
		config.ValueTypeString,
		config.DefaultValue("monokai"),
		config.Flag("style"),
		config.ShortFlag('s'),
		config.Description("Style for the response body syntax highlighting, see completion for available values"),
		config.Setter(func(_ *cobra.Command, val string) error {
			if val == "" {
				return nil
			}
			if !slices.Contains[[]string](styles.Names(), val) {
				return fmt.Errorf("%s: unnown style", val)
			}
			return nil
		}),
	)

	config.RegisterValue(
		configKeyHTTPDebugTrace,
		config.ValueTypeBool,
		config.Flag("trace"),
		config.ShortFlag('t'),
		config.Description("Trace HTTP request"),
	)

	commander.Register(
		"http>debug",
		func() *cobra.Command {
			return &cobra.Command{
				Use:   "debug [flags] url",
				Short: "Debug an HTTP url by sending a request and see output",
				Args:  cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					cfg := config.GetFromCommandContext(cmd)
					body, err := httpDebugGetBody(cmd)
					if err != nil {
						return fmt.Errorf("failed to get request body: %w", err)
					}
					opts := []httpdsak.Option{
						httpdsak.WithOut(cmd.OutOrStdout()),
						httpdsak.WithLog(cmd.ErrOrStderr()),
						httpdsak.WithMethod(cfg.GetString(configKeyHTTPDebugMethod)),
						httpdsak.WithAccept(cfg.GetString(configKeyHTTPDebugRequestAccept)),
						httpdsak.WithBody(body),
						httpdsak.WithAccept(cfg.GetString(configKeyHTTPDebugRequestContentType)),
					}
					if cfg.GetBool(configKeyHTTPDebugInsecure) {
						opts = append(opts, httpdsak.WithInsecure())
					}
					if cfg.GetBool(configKeyHTTPDebugRequestForceHTTP1) {
						opts = append(opts, httpdsak.WithForceHTTP1())
					}
					if cfg.GetBool(configKeyHTTPDebugRequestForceHTTP2) {
						opts = append(opts, httpdsak.WithForceHTTP2())
					}
					for _, h := range cfg.GetStringSlice(configKeyHTTPDebugRequestHeader) {
						parts := strings.SplitN(h, ":", 2)
						if len(parts) != 2 {
							return fmt.Errorf("invalid header: %s", h)
						}
						opts = append(opts, httpdsak.WithHeader(parts[0], parts[1]))
					}

					if cfg.GetString(configKeyHTTPDebugResponseForceType) != "" {
						opts = append(opts, httpdsak.WithForceType(cfg.GetString(configKeyHTTPDebugResponseForceType)))
					}
					if cfg.GetString(configKeyHTTPDebugResponseJQ) != "" {
						opts = append(opts, httpdsak.WithJQ(cfg.GetString(configKeyHTTPDebugResponseJQ)))
					}
					if cfg.GetBool(configKeyHTTPDebugResponseRaw) {
						opts = append(opts, httpdsak.WithRaw())
					}
					if cfg.GetString(configKeyHTTPDebugResponseStyle) != "" {
						opts = append(opts, httpdsak.WithStyle(cfg.GetString(configKeyHTTPDebugResponseStyle)))
					}
					if cfg.GetBool(configKeyHTTPDebugTrace) {
						opts = append(opts, httpdsak.WithTrace())
					}
					client, err := httpdsak.NewClient(args[0], opts...)
					if err != nil {
						return fmt.Errorf("failed to initialize HTTP client: %w", err)
					}
					cmd.SilenceUsage = true
					return client.Run(cmd.Context())
				},
			}
		},

		commander.WithConfig(configKeyHTTPDebugInsecure),
		commander.WithConfig(configKeyHTTPDebugMethod),
		commander.WithConfig(configKeyHTTPDebugRequestAccept),
		commander.WithConfig(configKeyHTTPDebugRequestBody),
		commander.WithConfig(configKeyHTTPDebugRequestBodyContent),
		commander.WithConfig(configKeyHTTPDebugRequestContentType),
		commander.WithConfig(configKeyHTTPDebugRequestForceHTTP1),
		commander.WithConfig(configKeyHTTPDebugRequestForceHTTP2),
		commander.WithConfig(configKeyHTTPDebugRequestHeader),
		commander.WithConfig(configKeyHTTPDebugResponseForceType),
		commander.WithConfig(configKeyHTTPDebugResponseJQ),
		commander.WithConfig(configKeyHTTPDebugResponseRaw),
		commander.WithConfig(configKeyHTTPDebugResponseStyle),
		commander.WithConfig(configKeyHTTPDebugTrace),

		commander.WithFlagCompletion(
			configKeyHTTPDebugMethod,
			func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				flags := cobra.ShellCompDirectiveNoFileComp
				methods := []string{
					http.MethodConnect,
					http.MethodDelete,
					http.MethodGet,
					http.MethodHead,
					http.MethodOptions,
					http.MethodPatch,
					http.MethodPost,
					http.MethodPut,
					http.MethodTrace,
				}
				list := make([]string, 0, len(methods))
				toComplete = strings.ToLower(toComplete)
				for _, v := range methods {
					if toComplete == "" || strings.Contains(strings.ToLower(v), toComplete) {
						list = append(list, v)
					}
				}
				sort.Strings(list)
				return list, flags
			},
		),

		commander.WithFlagCompletion(
			configKeyHTTPDebugRequestAccept,
			func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				flags := cobra.ShellCompDirectiveNoFileComp
				list := make([]string, 0, len(contenttype.List))
				toComplete = strings.ToLower(toComplete)
				for _, v := range contenttype.List {
					if toComplete == "" || strings.Contains(strings.ToLower(v), toComplete) {
						list = append(list, v)
					}
				}
				return list, flags
			},
		),

		commander.WithFlagCompletion(
			configKeyHTTPDebugResponseStyle,
			func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				flags := cobra.ShellCompDirectiveNoFileComp
				list := make([]string, 0, len(styles.Names()))
				toComplete = strings.ToLower(toComplete)
				for _, s := range styles.Names() {
					if toComplete == "" || strings.Contains(strings.ToLower(s), toComplete) {
						list = append(list, s)
					}
				}
				return list, flags
			},
		),

		commander.WithFlagCompletion(
			configKeyHTTPDebugRequestContentType,
			func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				flags := cobra.ShellCompDirectiveNoFileComp
				list := make([]string, 0, len(contenttype.List))
				toComplete = strings.ToLower(toComplete)
				for _, v := range contenttype.List {
					if toComplete == "" || strings.Contains(strings.ToLower(v), toComplete) {
						list = append(list, v)
					}
				}
				return list, flags
			},
		),
	)
}

func httpDebugGetBody(cmd *cobra.Command) (io.ReadCloser, error) {
	cfg := config.GetFromCommandContext(cmd)
	var body io.ReadCloser
	if cfg.GetString(configKeyHTTPDebugRequestBodyContent) != "" {
		body = io.NopCloser(strings.NewReader(cfg.GetString(configKeyHTTPDebugRequestBodyContent)))
	} else if cfg.GetString(configKeyHTTPDebugRequestBody) == "" {
		body = http.NoBody
	} else {
		var err error
		body, err = resource.New(cmd, cfg.GetString(configKeyHTTPDebugRequestBody), getLogger(cmd))
		if err != nil {
			return nil, err
		}
	}
	return body, nil
}
