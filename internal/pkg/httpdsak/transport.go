package httpdsak

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"
)

func (c *Client) buildTransport(ctx context.Context) (*http.Transport, error) {
	v, ok := ctx.Deadline()
	var tlsHandshakeTimeout time.Duration
	if ok {
		tlsHandshakeTimeout = time.Until(v)
	} else {
		tlsHandshakeTimeout = 10 * time.Second
	}
	var rtNextProto map[string]func(string, *tls.Conn) http.RoundTripper
	var cfgNextProto []string

	if c.forceHTTP1 {
		rtNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}
	}
	if c.forceHTTP2 {
		cfgNextProto = []string{"h2"}
	}

	if c.forceHTTP2 && c.url.Scheme == "http" {
		return nil, fmt.Errorf("HTTP/2 requires (de facto) a TLS connection (https://www.mnot.net/blog/2015/06/15/http2_implementation_status)")
	}

	return &http.Transport{
		DialContext:           getDialer,
		ExpectContinueTimeout: 1,
		ForceAttemptHTTP2:     true,
		IdleConnTimeout:       time.Second,
		MaxIdleConns:          1,
		MaxIdleConnsPerHost:   1,
		Proxy:                 http.ProxyFromEnvironment,
		TLSHandshakeTimeout:   tlsHandshakeTimeout,
		TLSNextProto:          rtNextProto,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: c.insecure, //nolint:gosec
			NextProtos:         cfgNextProto,
			VerifyConnection: func(cs tls.ConnectionState) error {
				if c.forceHTTP2 && cs.NegotiatedProtocol != "h2" {
					return fmt.Errorf("HTTP/2 was required")
				}
				return nil
			},
		},
	}, nil
}

func getDialer(ctx context.Context, network string, addr string) (net.Conn, error) {
	v, ok := ctx.Deadline()
	var timeout time.Duration
	if ok {
		timeout = time.Until(v)
	} else {
		timeout = 10 * time.Second
	}
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 0,
	}
	return dialer.DialContext(ctx, network, addr)
}
