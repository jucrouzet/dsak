package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/glob"
	netroute "github.com/libp2p/go-netroute"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/jucrouzet/dsak/internal/pkg/commander"
	"github.com/jucrouzet/dsak/internal/pkg/config"
)

const (
	configKeyHTTPCertDays = "http.cert.days"
)

func init() {
	config.RegisterValue(
		configKeyHTTPCertDays,
		config.ValueTypeUint,
		config.DefaultValue(uint64(0)),
		config.Flag("days"),
		config.ShortFlag('d'),
		config.Description("Set a number of days for the certificate to be valid for for NotAfter"),
	)

	commander.Register(
		"http>cert",
		func() *cobra.Command {
			return &cobra.Command{
				Use:   "cert [flags] hostname",
				Short: "Check HTTPS certificate for a given hostname",
				Args:  cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					domain, port, ips, err := httpCertResolve(cmd, args[0])
					if err != nil {
						return fmt.Errorf("failed to resolve %s: %w", args[0], err)
					}
					errs := make(map[*net.IP][]error)
					countRuns := 0
					errsMtx := sync.Mutex{}
					getResult := func(res *httpCertResult) {
						errsMtx.Lock()
						errs[&res.ip] = res.errs
						if res.runned {
							countRuns++
						}
						errsMtx.Unlock()
					}
					wg := sync.WaitGroup{}
					for _, ip := range ips {
						wg.Add(1)
						go func(ip net.IP) {
							defer wg.Done()
							getResult(httpCertOnIP(cmd, domain, ip, port))
						}(ip)
					}
					wg.Wait()

					valid := true
					for ip, err := range errs {
						ipStr := ip.String()
						if len(err) == 0 {
							fmt.Printf("%s: OK\n", ipStr)
						} else {
							for _, e := range err {
								if e != nil {
									fmt.Printf("%s: %s\n", ipStr, e)
									valid = false
								}
							}
						}
					}
					if !valid {
						cmd.SilenceUsage = true
						return errors.New("failed to validate certificate")
					}
					return nil
				},
			}
		},
		commander.WithConfig(configKeyHTTPCertDays),
	)
}

func httpCertResolve(_ *cobra.Command, arg string) (string, int, []net.IP, error) {
	domain := arg
	port := int64(443)
	if strings.Contains(arg, "://") {
		var err error
		domain, port, err = httpCertResolveURL(arg)
		if err != nil {
			return "", 0, nil, fmt.Errorf("invalid URL: %w", err)
		}
	}
	parts := strings.SplitN(domain, ":", 2)
	if len(parts) == 2 {
		var err error
		port, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return "", 0, nil, fmt.Errorf("invalid port: %w", err)
		}
		domain = parts[0]
	}
	ip := net.ParseIP(domain)
	if ip != nil {
		return domain, int(port), []net.IP{ip}, nil
	}
	ips, err := net.LookupIP(domain)
	if err != nil {
		return "", 0, nil, err
	}
	return domain, int(port), ips, nil
}

func httpCertResolveURL(uri string) (string, int64, error) {
	port := int64(443)
	u, err := url.Parse(uri)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse %s: %w", uri, err)
	}
	if u.Scheme != "https" {
		return "", 0, fmt.Errorf("%s is not an HTTPS URL", uri)
	}
	if u.Hostname() == "" {
		return "", 0, fmt.Errorf("%s has no host", uri)
	}
	if u.Port() != "" {
		port, err = strconv.ParseInt(u.Port(), 10, 64)
		if err != nil || port <= 0 || port > 65535 {
			return "", 0, fmt.Errorf("%s has invalid port: %w", uri, err)
		}
	}
	return u.Hostname(), port, nil
}

type httpCertResult struct {
	errs   []error
	runned bool
	ip     net.IP
}

func httpCertOnIPErrResult(err error, ip net.IP, runned bool) *httpCertResult {
	return &httpCertResult{
		errs:   []error{err},
		runned: runned,
		ip:     ip,
	}
}

func httpCertOnIP(cmd *cobra.Command, domain string, ip net.IP, port int) *httpCertResult {
	r, err := netroute.New()
	if err != nil {
		return httpCertOnIPErrResult(fmt.Errorf("failed to get routing table: %w", err), ip, false)
	}
	iface, _, _, err := r.Route(ip)
	if err != nil {
		return httpCertOnIPErrResult(fmt.Errorf("failed to get route for %s: %w", ip.String(), err), ip, false)
	}
	if iface == nil {
		return &httpCertResult{
			runned: false,
			ip:     ip,
		}
	}

	address := net.JoinHostPort(ip.String(), strconv.Itoa(port))

	logger := getLogger(cmd).
		With(zap.String("server_address", address)).
		With(zap.String("domain", domain))
	logger.Debug("connecting to server")

	conn, err := tls.Dial(
		"tcp",
		address,
		&tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
			ServerName:         domain,
		},
	)
	if err != nil {
		return httpCertOnIPErrResult(fmt.Errorf("cannot connect to server or is not an HTTPS server: %w", err), ip, true)
	}
	defer conn.Close()
	return &httpCertResult{
		runned: true,
		ip:     ip,
		errs:   httpCertOnIPCheck(cmd, domain, ip, port, conn.ConnectionState().PeerCertificates, logger),
	}
}

func httpCertOnIPCheck(
	cmd *cobra.Command,
	domain string,
	ip net.IP,
	port int,
	certs []*x509.Certificate,
	logger *zap.Logger,
) []error {
	if len(certs) == 0 {
		return []error{
			fmt.Errorf("no certificate found for %s:%d", ip.String(), port),
		}
	}
	var res []error
	var root *x509.CertPool
	var inter *x509.CertPool
	if len(certs) > 1 {
		root = x509.NewCertPool()
		root.AddCert(certs[len(certs)-1])
	}
	if len(certs) > 2 {
		inter = x509.NewCertPool()
		for _, c := range certs[1 : len(certs)-1] {
			inter.AddCert(c)
		}
	}

	res = append(res, httpCertOnIPCheckCert(cmd, domain, certs[0], logger, "server certificate")...)
	if inter != nil {
		for i := 1; i < len(certs)-2; i++ {
			res = append(res, httpCertOnIPCheckCert(cmd, domain, certs[0], logger, "intermediate certificate")...)
		}
	}
	if root != nil {
		res = append(res, httpCertOnIPCheckCert(cmd, domain, certs[len(certs)-1], logger, "root certificate")...)
	}

	_, err := certs[0].Verify(x509.VerifyOptions{
		DNSName:       domain,
		Roots:         root,
		Intermediates: inter,
	})
	if err != nil {
		res = append(res, err)
	}

	return res
}

func httpCertOnIPCheckCert(
	cmd *cobra.Command,
	domain string,
	cert *x509.Certificate,
	logger *zap.Logger,
	certType string,
) []error {
	logger = logger.
		With(zap.String("cert_type", certType)).
		With(zap.String("cert_subject", cert.Subject.CommonName))
	logger.Debug("checking certificate")

	var res []error

	if certType == "server certificate" {
		res = append(res, httpCertOnIPCheckCertSAN(domain, cert, logger))
	}

	res = append(res, httpCertOnIPCheckCertDates(cmd, cert, certType)...)
	return res
}

func httpCertOnIPCheckCertSAN( //nolint:gocyclo
	domain string,
	cert *x509.Certificate,
	logger *zap.Logger,
) error {
	var names []string
	ips := make([]net.IP, 0)
	urls := make([]*url.URL, 0)

	for _, name := range cert.DNSNames {
		if !slices.Contains(names, name) {
			names = append(names, name)
		}
	}
	for _, givenIP := range cert.IPAddresses {
		for _, ip := range ips {
			if ip.Equal(givenIP) {
				continue
			}
		}
		ips = append(ips, givenIP)
	}
	for _, givenURL := range cert.URIs {
		for _, url := range urls {
			if url.String() == givenURL.String() {
				continue
			}
		}
		urls = append(urls, givenURL)
	}
	if len(names) == 0 && len(ips) == 0 && len(urls) == 0 {
		return errors.New("server certificate has no Subject Alternate Name")
	}
	if len(names) > 0 {
		logger.With(zap.Strings("SAN Domains", names)).Debug("Checking Subject Alternate Name domains")
		for _, name := range names {
			g, err := glob.Compile(name)
			if err != nil {
				continue
			}
			if g.Match(domain) {
				return nil
			}
		}
	}
	if len(ips) > 0 {
		ipsStrs := make([]string, len(ips))
		for i, ip := range ips {
			ipsStrs[i] = ip.String()
		}
		logger.With(zap.Strings("SAN IPs", ipsStrs)).Debug("Checking Subject Alternate Name ips")
		hostIP := net.ParseIP(domain)
		if hostIP != nil {
			for _, ip := range ips {
				if !ip.Equal(hostIP) {
					continue
				}
				return nil
			}
		}
	}
	return errors.New("server certificate does not match any Subject Alternate Name")
}

func httpCertOnIPCheckCertDates(cmd *cobra.Command, cert *x509.Certificate, certType string) []error {
	days := config.GetFromCommandContext(cmd).GetUint64(configKeyHTTPCertDays)
	afterDate := time.Now().Add(time.Duration(days) * 24 * time.Hour)
	var errs []error
	if cert.NotBefore.After(time.Now()) {
		errs = append(errs, fmt.Errorf("%s %q is not valid before %s", certType, cert.Subject, cert.NotBefore.Format(time.RFC3339)))
	}
	if cert.NotAfter.Before(afterDate) {
		errs = append(errs, fmt.Errorf("%s %q expires on %s", certType, cert.Subject, cert.NotAfter.Format(time.RFC3339)))
	}
	return errs
}
