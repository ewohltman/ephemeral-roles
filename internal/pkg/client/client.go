// Package client provides an HTTP client with a wrapping http.RoundTripper for
// additional features.
package client

import (
	"context"
	"net/http"
	"time"
)

const (
	clientTimeout = 20 * time.Second

	contextTimeout = 15 * time.Second
)

// New returns a new *HTTP client.
func New() *http.Client {
	client := &http.Client{Timeout: clientTimeout}

	SetTransport(client)

	return client
}

// SetTransport takes an *http.Client and sets its Transport to a custom
// RoundTripper, wrapping an existing *http.Transport if already set or
// allocating a new *http.Transport if not already set.
func SetTransport(client *http.Client) {
	if client.Transport != nil {
		transport, ok := client.Transport.(*http.Transport)
		if ok {
			client.Transport = roundTripper(transport)
			return
		}
	}

	transport := &http.Transport{}

	client.Transport = roundTripper(transport)
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements the RoundTripper interface.
func (rt roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt(r)
}

func roundTripper(next http.RoundTripper) roundTripperFunc {
	return func(r *http.Request) (*http.Response, error) {
		ctx, cancelCtx := context.WithTimeout(context.Background(), contextTimeout)
		defer cancelCtx()

		return next.RoundTrip(r.WithContext(ctx))
	}
}
