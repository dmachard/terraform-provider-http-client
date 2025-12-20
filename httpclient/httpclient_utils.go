package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/quic-go/quic-go/http3"
)

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

func configureHTTPTransport(httpVersion string, tlsConfig *tls.Config, timeout time.Duration) (http.RoundTripper, error) {
	version := strings.ToUpper(httpVersion)

	switch version {
	case "", "HTTP1.1", "HTTP/1.1":
		return &http.Transport{
			TLSClientConfig:       tlsConfig,
			ForceAttemptHTTP2:     false,
			DialContext:           (&net.Dialer{Timeout: timeout}).DialContext,
			TLSHandshakeTimeout:   timeout,
			ResponseHeaderTimeout: timeout,
		}, nil
	case "HTTP2", "HTTP/2":
		return &http.Transport{
			TLSClientConfig:       tlsConfig,
			ForceAttemptHTTP2:     true,
			DialContext:           (&net.Dialer{Timeout: timeout}).DialContext,
			TLSHandshakeTimeout:   timeout,
			ResponseHeaderTimeout: timeout,
		}, nil
	case "HTTP3", "HTTP/3":
		if tlsConfig.MinVersion < tls.VersionTLS13 {
			return nil, fmt.Errorf("HTTP/3 requires TLS 1.3")
		}
		return &http3.Transport{
			TLSClientConfig: tlsConfig,
		}, nil
	default:
		return nil, fmt.Errorf("invalid HTTP version: %s (valid values: HTTP1.1, HTTP2)", httpVersion)
	}
}

// executeRequest executes an HTTP request based on configuration and returns response
type RequestConfig struct {
	Ctx                 context.Context
	URL                 string
	Method              string
	Headers             map[string]string
	Body                []byte
	Username            string
	Password            string
	Timeout             time.Duration
	Insecure            bool
	TLSMinVersion       string
	ClientCertPEM       string
	ClientKeyPEM        string
	CACertPEM           string
	HTTPVersion         string
	ExpectedStatusCodes []int
	FailOnHTTPError     bool
}

type RequestResult struct {
	ResponseCode    int
	ResponseHeaders map[string]string
	ResponseBody    []byte
}

func ExecuteRequest(cfg RequestConfig) (*RequestResult, error) {
	// Build TLS config
	minVersion := uint16(tls.VersionTLS12)
	if cfg.TLSMinVersion != "" {
		v, err := getTLSVersion(cfg.TLSMinVersion)
		if err != nil {
			return nil, err
		}
		minVersion = v
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.Insecure,
		MinVersion:         minVersion,
	}

	// Client certificate (mTLS)
	if cfg.ClientCertPEM != "" && cfg.ClientKeyPEM != "" {
		cert, err := tls.X509KeyPair([]byte(cfg.ClientCertPEM), []byte(cfg.ClientKeyPEM))
		if err != nil {
			return nil, fmt.Errorf("failed to parse client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// CA certificate
	if cfg.CACertPEM != "" {
		caCertPool := x509.NewCertPool()
		rest := []byte(cfg.CACertPEM)
		for {
			block, r := pem.Decode(rest)
			if block == nil {
				break
			}
			if block.Type == "CERTIFICATE" {
				cert, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					return nil, fmt.Errorf("failed to parse CA certificate: %w", err)
				}
				caCertPool.AddCert(cert)
			}
			rest = r
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Build HTTP request
	req, err := http.NewRequestWithContext(cfg.Ctx, cfg.Method, cfg.URL, bytes.NewBuffer(cfg.Body))
	if err != nil {
		return nil, err
	}

	// Headers
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	// Basic auth
	if cfg.Username != "" {
		req.SetBasicAuth(cfg.Username, cfg.Password)
	}

	// Transport + client
	transport, err := configureHTTPTransport(cfg.HTTPVersion, tlsConfig, cfg.Timeout)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
	}

	// Do request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check expected status codes
	if cfg.FailOnHTTPError {
		valid := false
		for _, code := range cfg.ExpectedStatusCodes {
			if resp.StatusCode == code {
				valid = true
				break
			}
		}
		if !valid && len(cfg.ExpectedStatusCodes) > 0 {
			return nil, fmt.Errorf("unexpected HTTP status code %d", resp.StatusCode)
		}
	}

	// Parse headers
	headers := make(map[string]string)
	for k, v := range resp.Header {
		headers[k] = strings.Join(v, ", ")
	}

	return &RequestResult{
		ResponseCode:    resp.StatusCode,
		ResponseHeaders: headers,
		ResponseBody:    respBody,
	}, nil
}
