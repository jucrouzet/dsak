package resourcetype

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/jucrouzet/dsak/internal/pkg/resource/http"
	"github.com/jucrouzet/dsak/internal/pkg/resource/standard"
)

var (
	ErrUnknownType = errors.New("unknown resource type")
)

const (
	rTypeUnknown = ""
	rTypeStdIn   = "stdin"
	rTypeStdOut  = "stdout"
	rTypeStdErr  = "stderr"
	rTypeFile    = "file"
	rTypeHTTP    = "http"
)

func Parse(cmd *cobra.Command, s string, logger *zap.Logger) (Handler, error) {
	if strings.EqualFold(s, rTypeStdIn) || strings.EqualFold(s, "-") {
		logger.Debug("resource is stdin")
		return standard.NewStdIn()
	}
	if strings.EqualFold(s, rTypeStdOut) {
		logger.Debug("resource is stdout")
		return standard.NewStdOut()
	}
	if strings.EqualFold(s, rTypeStdErr) {
		logger.Debug("resource is stderr")
		return standard.NewStdErr()
	}
	if !strings.Contains(s, "://") {
		logger.Debug("resource is a file")
		return standard.NewFile(s, logger)
	}
	uri, err := url.Parse(s)
	if err != nil {
		logger.With(zap.Error(err)).Warn("failed to parse resource url")
		return nil, ErrUnknownType
	}
	switch uri.Scheme {
	case rTypeHTTP, "https":
		logger.Debug("resource is http(s)")
		return http.New(cmd, s, logger)
	default:
		return nil, fmt.Errorf("%w: unsupported scheme: %s", ErrUnknownType, uri.Scheme)
	}
}
