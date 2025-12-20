package httpclient

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"github.com/quic-go/quic-go/http3"
)

// getTLSVersion converts string to tls version constant
func getTLSVersion(version string) (uint16, error) {
	switch strings.ToUpper(version) {
	case "TLS10":
		return tls.VersionTLS10, nil
	case "TLS11":
		return tls.VersionTLS11, nil
	case "TLS12":
		return tls.VersionTLS12, nil
	case "TLS13":
		return tls.VersionTLS13, nil
	default:
		return 0, fmt.Errorf("invalid TLS version: %s (valid values: TLS10, TLS11, TLS12, TLS13)", version)
	}
}

// configureHTTPTransport configures the HTTP transport based on version
func configureHTTPTransport(httpVersion string, tlsConfig *tls.Config) (http.RoundTripper, error) {
	version := strings.ToUpper(httpVersion)

	switch version {
	case "", "HTTP1.1", "HTTP/1.1":
		return &http.Transport{
			TLSClientConfig:   tlsConfig,
			ForceAttemptHTTP2: false,
		}, nil
	case "HTTP2", "HTTP/2":
		return &http.Transport{
			TLSClientConfig:   tlsConfig,
			ForceAttemptHTTP2: true,
		}, nil
	case "HTTP3", "HTTP/3":
		return &http3.Transport{
			TLSClientConfig: tlsConfig,
		}, nil
	default:
		return nil, fmt.Errorf("invalid HTTP version: %s (valid values: HTTP1.1, HTTP2)", httpVersion)
	}
}
