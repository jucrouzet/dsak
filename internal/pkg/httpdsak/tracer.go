package httpdsak

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http/httptrace"
	"strings"
	"time"
)

func (c *Client) getTracerContext(ctx context.Context) context.Context {
	if !c.trace {
		return ctx
	}
	var (
		starDialing     time.Time
		startConnect    time.Time
		startDNSResolve time.Time
		startHandshake  time.Time
		startHeaders    time.Time
		startRequest    time.Time
		startTTFB       time.Time
	)

	tracer := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			starDialing = time.Now()
			c.traceInfo("Dialing ")
			c.traceValue(hostPort)
			c.traceInfoln("...")
		},
		GotConn: func(infos httptrace.GotConnInfo) {
			c.traceInfo("Got a valid connection to ")
			c.traceValue(infos.Conn.RemoteAddr().String())
			c.traceInfo(" in ")
			c.traceValueln(time.Since(starDialing).String())
			startRequest = time.Now()
		},
		ConnectStart: func(network, addr string) {
			startConnect = time.Now()
			c.traceInfo("Starting connection to ")
			c.traceValuef("%s:%s", network, addr)
			c.traceInfoln("...")
		},
		ConnectDone: func(network, addr string, err error) {
			addr = fmt.Sprintf("%s:%s", network, addr)
			dur := time.Since(startConnect).String()
			if err != nil {
				c.traceInfo("Could not establish a connection to ")
				c.traceValue(addr)
				c.traceInfo(" after ")
				c.traceValue(dur)
				c.traceInfo(" : ")
				c.traceErrorEln(err)
				return
			}
			c.traceInfo("Established a connection to ")
			c.traceValue(addr)
			c.traceInfo(" after ")
			c.traceValueln(dur)
		},
		DNSStart: func(infos httptrace.DNSStartInfo) {
			startDNSResolve = time.Now()
			c.traceInfo("Resolving ")
			c.traceValuef(infos.Host)
			c.traceInfoln("...")
		},
		DNSDone: func(infos httptrace.DNSDoneInfo) {
			res := make([]string, len(infos.Addrs))
			for i, v := range infos.Addrs {
				res[i] = v.String()
			}
			c.traceInfo("Resolved hostname to ")
			c.traceValuef(`["%s"]`, strings.Join(res, ", "))
			c.traceInfo(" in ")
			c.traceValueln(time.Since(startDNSResolve).String())
		},
		TLSHandshakeStart: func() {
			startHandshake = time.Now()
			c.traceInfoln("Starting TLS handshake...")
		},

		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			if err != nil {
				c.traceInfo("TLS handshake failed : ")
				c.traceErrorE(err)
			} else {
				c.traceInfo("TLS handshake successful")
			}
			c.traceInfo(" in ")
			c.traceValueln(time.Since(startHandshake).String())
			if state.NegotiatedProtocol != "" {
				c.traceInfo("ALPN negotiated protocol is ")
				c.traceValueln(state.NegotiatedProtocol)
			}
			tlsVersion := ""
			switch state.Version {
			case tls.VersionTLS10:
				tlsVersion = "TLS 1.0"
			case tls.VersionTLS11:
				tlsVersion = "TLS 1.1"
			case tls.VersionTLS12:
				tlsVersion = "TLS 1.2"
			case tls.VersionTLS13:
				tlsVersion = "TLS 1.3"
			}
			if tlsVersion != "" {
				c.traceInfo("TLS version is ")
				c.traceValueln(tlsVersion)
			}
			if len(state.PeerCertificates) > 0 {
				c.checkCerts(state.PeerCertificates)
			}
			if len(state.VerifiedChains) == 0 && !c.insecure {
				c.traceErrorln("Certificate is not verified")
			}
		},

		WroteHeaderField: func(key string, value []string) {
			if startHeaders.IsZero() {
				startHeaders = time.Now()
			}
			c.traceInfo("Sent header ")
			c.traceValuef("%q", key)
			c.traceInfo(" : ")
			if len(value) > 1 {
				c.traceValuef("%q\n", value)
			} else {
				c.traceValuef("%q\n", value[0])
			}
		},

		WroteHeaders: func() {
			if !startHeaders.IsZero() {
				c.traceInfo("All headers have been sent in ")
				c.traceValuef("%s\n", time.Since(startHeaders).String())
			}
		},

		WroteRequest: func(infos httptrace.WroteRequestInfo) {
			if infos.Err != nil {
				c.traceInfo("Failed to write request : ")
				c.traceErrorE(infos.Err)
				c.traceInfo(" ")
			} else {
				c.traceInfo("Request is sent ")
			}
			c.traceInfo("in ")
			c.traceValuef("%s\n", time.Since(startRequest).String())
			startTTFB = time.Now()
		},

		GotFirstResponseByte: func() {
			c.traceInfo("Received first byte in ")
			c.traceValuef("%s\n", time.Since(startTTFB).String())
		},
	}
	return httptrace.WithClientTrace(ctx, tracer)
}
