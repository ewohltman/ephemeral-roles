// Package client provides an HTTP client with a wrapping http.RoundTripper for
// additional features.
package client

import (
	"context"
	"net/http"
	"reflect"
	"time"
)

const contextTimeout = 15 * time.Second

// New returns a new *HTTP client.
func New() *http.Client {
	client := &http.Client{}

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
		if reflect.DeepEqual(r.Context(), context.Background()) {
			ctx, cancelCtx := context.WithTimeout(context.Background(), contextTimeout)
			defer cancelCtx()

			r = r.Clone(ctx)
		}

		return next.RoundTrip(r)
	}
}
