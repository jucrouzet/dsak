// Package dns handles DNS operations.
package dns

import (
	"context"
	"fmt"
	"strings"

	"github.com/miekg/dns"
	dnslib "github.com/miekg/dns"
	"go.uber.org/zap"
)

// Client represents a DNS client.
type Client struct {
	logger  *zap.Logger
	servers []string
}

// NewClient creates a new DNS client.
func NewClient(logger *zap.Logger, servers ...string) *Client {
	if len(servers) == 0 {
		servers = []string{"1.1.1.1", "8.8.8.8"}
	}
	logger.With(zap.Strings("dns_servers", servers)).Debug("Initializing DNS client")
	return &Client{
		logger:  logger,
		servers: servers,
	}
}

// Query performs a DNS query for the given domain and returns the response.
func (c *Client) Query(ctx context.Context, rType Type, domain string) (string, error) {
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}
	client := new(dnslib.Client)
	co, err := client.DialContext(ctx, "8.8.8.8:53")
	if err != nil {
		return "", err
	}
	defer co.Close()
	m := new(dns.Msg)
	m.SetQuestion(domain, uint16(rType))
	in, rtt, err := client.ExchangeWithConnContext(ctx, m, co)
	if err != nil {
		return "", fmt.Errorf("DNS query failed: %w", err)
	}
	c.logger.With(
		zap.String("domain", domain),
		zap.String("type", GetTypeName(rType)),
		zap.Duration("rtt", rtt),
	).Debug("DNS query succeeded")
	return in.String(), nil
}
