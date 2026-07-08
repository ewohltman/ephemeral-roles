package http

import (
	"net"
	"net/http"
	"time"
)

const (
	clientTimeout         = 1 * time.Minute
	dialerTimeout         = 30 * time.Second
	tlsHandshakeTimeout   = 30 * time.Second
	responseHeaderTimeout = 1 * time.Minute
)

const poolSize = 100

// NewClient returns a new preconfigured *http.Client.
func NewClient(transport http.RoundTripper) *http.Client {
	return &http.Client{
		Transport: transport,
		Timeout:   clientTimeout,
	}
}

// NewTransport returns a new pre-configured *http.Transport.
func NewTransport() *http.Transport {
	return &http.Transport{
		DialContext:           (&net.Dialer{Timeout: dialerTimeout}).DialContext,
		TLSHandshakeTimeout:   tlsHandshakeTimeout,
		MaxIdleConns:          poolSize,
		MaxIdleConnsPerHost:   poolSize,
		MaxConnsPerHost:       poolSize,
		ResponseHeaderTimeout: responseHeaderTimeout,
	}
}
