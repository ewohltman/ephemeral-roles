package http

import (
	"net"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/tracer"
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

// TransportWrapper is a function that returns an http.RoundTripper that wraps
// the next http.RoundTripper by calling its RoundTrip method.
type TransportWrapper func(next http.RoundTripper) http.RoundTripper

// WrapTransport returns an http.RoundTripper with all the provided
// TransportWrapper functions wrapping the provided http.RoundTripper in order.
func WrapTransport(roundTripper http.RoundTripper, transportWrappers ...TransportWrapper) http.RoundTripper {
	for _, transportWrapper := range transportWrappers {
		roundTripper = transportWrapper(roundTripper)
	}

	return roundTripper
}

// WrapTransportWithTracer wraps the next http.RoundTripper with a Jaeger
// tracer.
func WrapTransportWithTracer(jaegerTracer opentracing.Tracer, instanceName string) TransportWrapper {
	return func(next http.RoundTripper) http.RoundTripper {
		return tracer.RoundTripper(jaegerTracer, instanceName, next)
	}
}
