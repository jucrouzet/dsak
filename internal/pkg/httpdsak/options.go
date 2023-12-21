package httpdsak

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

	"github.com/alecthomas/chroma/styles"
	"github.com/itchyny/gojq"
)

// Option is a function that configures a Client.
type Option func(*Client) error

/*
	commander.WithConfig(configKeyHTTPDebugResponseStyle),
	commander.WithConfig(configKeyHTTPDebugTrace),

*/

// WithOut configures the client output writer.
func WithOut(writer io.Writer) Option {
	return func(c *Client) error {
		c.out = writer
		return nil
	}
}

// WithLog configures the client log writer.
func WithLog(writer io.Writer) Option {
	return func(c *Client) error {
		c.log = writer
		return nil
	}
}

func WithInsecure() Option {
	return func(c *Client) error {
		c.insecure = true
		return nil
	}
}

// WithMethod configures the method to use while doing the request.
func WithMethod(method string) Option {
	return func(c *Client) error {
		if method == "" {
			return errors.New("method cannot be empty")
		}
		if !regexp.MustCompile("^[a-zA-Z]+$").MatchString(method) {
			return fmt.Errorf("%s: invalid method", method)
		}
		c.method = strings.ToUpper(method)
		return nil
	}
}

// WithAccept configures the request accept header value.
func WithAccept(v string) Option {
	return func(c *Client) error {
		c.acccept = v
		return nil
	}
}

// WithBody configures the request body.
func WithBody(b io.ReadCloser) Option {
	return func(c *Client) error {
		c.body = b
		return nil
	}
}

// WithContentType configures the request content-type header value.
func WithContentType(v string) Option {
	return func(c *Client) error {
		c.contentType = v
		return nil
	}
}

// WithForceHTTP1 set client to force using HTTP/1.1.
func WithForceHTTP1() Option {
	return func(c *Client) error {
		if c.forceHTTP2 {
			return errors.New("cannot force HTTP/1.1 and HTTP/2 at the same time")
		}
		c.forceHTTP1 = true
		return nil
	}
}

// WithForceHTTP2 set client to force using HTTP/2.
func WithForceHTTP2() Option {
	return func(c *Client) error {
		if c.forceHTTP1 {
			return errors.New("cannot force HTTP/1.1 and HTTP/2 at the same time")
		}
		c.forceHTTP2 = true
		return nil
	}
}

// WithHeader adds an header to request.
func WithHeader(name, value string) Option {
	return func(c *Client) error {
		c.headers.Add(name, value)
		return nil
	}
}

// WithJQ configures the jq query to filter a JSON response.
func WithJQ(filter string) Option {
	return func(c *Client) error {
		q, err := gojq.Parse(filter)
		if err != nil {
			return fmt.Errorf("invalid jq filter: %w", err)
		}
		c.jq = q
		return nil
	}
}

// WithRaw tells client to output the raw response body.
func WithRaw() Option {
	return func(c *Client) error {
		c.raw = true
		return nil
	}
}

// WithStyle tells client to use the given style while highlighting response body.
func WithStyle(style string) Option {
	return func(c *Client) error {
		if !slices.Contains[[]string](styles.Names(), style) {
			return fmt.Errorf("%s: unknown style", style)
		}
		c.style = style
		return nil
	}
}

// WithTrace tells client to show a tracing log of the request execution.
func WithTrace() Option {
	return func(c *Client) error {
		c.trace = true
		return nil
	}
}

// WithForceType tells client to force using given content-type for response.
func WithForceType(t string) Option {
	return func(c *Client) error {
		c.forceType = t
		return nil
	}
}
