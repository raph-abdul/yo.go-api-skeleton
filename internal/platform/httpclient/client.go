// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package httpclient /youGo/internal/platform/httpclient/client.go
package httpclient

import (
	"net"
	"net/http"
	"time"
	// Import config if timeout values are stored there
	// "youGo/internal/config"
)

// DefaultTimeout is a reasonable default timeout for external HTTP calls.
const DefaultTimeout = 15 * time.Second

// NewHTTPClient creates a new *http.Client with sensible defaults.
// Customize by passing configuration options if needed.
func NewHTTPClient( /* cfg config.HTTPClientConfig */ timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	// Configure the transport with timeouts for connection establishment, etc.
	// These are lower-level timeouts compared to the client's total timeout.
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment, // Respect environment proxy settings
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second, // Timeout for establishing the connection
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,             // Prefer HTTP/2
		MaxIdleConns:          100,              // Max idle connections overall
		MaxIdleConnsPerHost:   10,               // Max idle connections per host
		IdleConnTimeout:       90 * time.Second, // Timeout for idle connections
		TLSHandshakeTimeout:   5 * time.Second,  // Timeout for TLS handshake
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{
		Timeout:   timeout,   // Total timeout for the entire request-response cycle
		Transport: transport, // Use the configured transport
	}

	return client
}

// You can add more specific clients here if needed, e.g., a client
// pre-configured with specific headers or authentication for a particular external API.
// func NewMyExternalServiceClient(apiKey string, timeout time.Duration) *http.Client {
//  client := NewHTTPClient(timeout)
//  // Add custom transport wrapper to inject API key?
//  // client.Transport = &apiKeyTransport{apiKey: apiKey, roundTripper: client.Transport}
//  return client
// }
