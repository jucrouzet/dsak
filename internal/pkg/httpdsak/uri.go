package httpdsak

import (
	"fmt"
	"net/url"
	"strings"
)

func parseURL(uri string) (*url.URL, error) {
	if !strings.Contains(uri, "://") {
		uri = "http://" + uri
	}
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("not an http url")
	}
	if u.Host == "" {
		return nil, fmt.Errorf("no hostname in url")
	}
	return u, nil
}
