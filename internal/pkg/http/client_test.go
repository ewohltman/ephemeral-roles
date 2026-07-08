package http_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	internalHTTP "github.com/ewohltman/ephemeral-roles/internal/pkg/http"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/tracer"
)

const jaegerServiceName = "ephemeral-roles"

func TestNewClient(t *testing.T) {
	t.Parallel()

	testServer := httptest.NewServer(testServerHandler())
	defer testServer.Close()

	jaegerTracer, jaegerCloser, err := tracer.New(jaegerServiceName)
	require.NoError(t, err)

	defer func() { _ = jaegerCloser.Close() }()

	client := internalHTTP.NewClient(tracer.RoundTripper(
		jaegerTracer,
		"",
		internalHTTP.NewTransport(),
	))

	require.NotNil(t, client)
	require.NotNil(t, client.Transport)

	resp, err := doRequest(t.Context(), client, testServer.URL)
	require.NoError(t, err)

	drainCloseResponse(resp)
}

func testServerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
		_, _ = w.Write(nil)
	}
}
