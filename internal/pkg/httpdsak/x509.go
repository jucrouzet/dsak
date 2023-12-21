package httpdsak

import (
	"crypto/x509"
	"net"
	"net/url"
	"slices"
	"strings"

	"github.com/gobwas/glob"
)

func (c *Client) checkCerts(certs []*x509.Certificate) {
	hostname := strings.ToLower(c.url.Hostname())
	names, ips, urls := httpDebugGetCertsSANs(certs)
	if len(names) == 0 && len(ips) == 0 && len(urls) == 0 {
		c.traceErrorln("Certificate has no Subject Alternate Name")
		return
	}
	if len(names) > 0 {
		c.traceInfo("Certificate Subject Alternate Name domains are ")
		c.traceValuef("%q\n", names)
		for _, name := range names {
			g, err := glob.Compile(name)
			if err != nil {
				continue
			}
			if g.Match(hostname) {
				c.traceInfo("Certificate Subject Alternate Name domain ")
				c.traceValue(name)
				c.traceInfo(" matches hostname ")
				c.traceValueln(c.url.Hostname())
				return
			}
		}
	}
	if len(ips) > 0 {
		c.traceInfo("Certificate Subject Alternate Name ips are ")
		c.traceValuef("%q\n", ips)
		hostIP := net.ParseIP(hostname)
		if hostIP != nil {
			for _, ip := range ips {
				if !ip.Equal(hostIP) {
					continue
				}
				c.traceInfo("Certificate Subject Alternate Name ip ")
				c.traceValue(ip.String())
				c.traceInfoln(" matches hostname")
				return
			}
		}
	}
	if len(urls) > 0 {
		c.traceInfo("Certificate Subject Alternate Name uris are ")
		c.traceValuef("%q\n", urls)
		givenURL := c.url.String()
		for _, url := range urls {
			if !strings.HasPrefix(givenURL, url.String()) {
				continue
			}
			c.traceInfo("Certificate Subject Alternate Name uri ")
			c.traceValue(url.String())
			c.traceInfoln(" matches url")
			return
		}
	}
	c.traceErrorln("None of the Certificate Subject Alternate Name are valid")
}

func httpDebugGetCertsSANs(certs []*x509.Certificate) ([]string, []net.IP, []*url.URL) {
	var names []string
	var ips []net.IP
	var urls []*url.URL
	for _, cert := range certs {
		for _, name := range cert.DNSNames {
			if !slices.Contains(names, name) {
				names = append(names, name)
			}
		}
	ipLoop:
		for _, givenIP := range cert.IPAddresses {
			for _, ip := range ips {
				if ip.Equal(givenIP) {
					continue ipLoop
				}
			}
			ips = append(ips, givenIP)
		}

	urlsLoop:
		for _, givenURL := range cert.URIs {
			for _, url := range urls {
				if url.String() == givenURL.String() {
					continue urlsLoop
				}
			}
			urls = append(urls, givenURL)
		}
	}
	return names, ips, urls
}
