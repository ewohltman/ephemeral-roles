package http_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	internalHTTP "github.com/ewohltman/ephemeral-roles/internal/pkg/http"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/tracer"
)

const jaegerServiceName = "ephemeral-roles"

func TestNewClient(t *testing.T) {
	testServer := httptest.NewServer(testServerHandler())
	defer testServer.Close()

	jaegerTracer, jaegerCloser, err := tracer.New(jaegerServiceName)
	if err != nil {
		t.Fatalf("Error setting up Jaeger tracer: %s", err)
	}

	defer func() {
		closeErr := jaegerCloser.Close()
		if closeErr != nil {
			t.Errorf("Error closing Jaeger tracer: %s", closeErr)
		}
	}()

	client := internalHTTP.NewClient(internalHTTP.WrapTransport(
		internalHTTP.NewTransport(),
		internalHTTP.WrapTransportWithTracer(jaegerTracer, ""),
	))

	if client == nil {
		t.Fatal("Unexpected nil *http.Client")
	}

	if client.Transport == nil {
		t.Fatal("Unexpected nil http.RoundTripper")
	}

	err = doTestRequests(client, testServer.URL)
	if err != nil {
		t.Fatalf("Error doing test request: %s", err)
	}
}

func testServerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			_, err := io.Copy(io.Discard, r.Body)
			if err != nil {
				panic(fmt.Sprintf("Error draining test request body: %s", err))
			}
		}

		err := r.Body.Close()
		if err != nil {
			panic(fmt.Sprintf("Error closing test request body: %s", err))
		}

		_, err = w.Write([]byte{})
		if err != nil {
			panic(fmt.Sprintf("Error writing test response body: %s", err))
		}
	}
}

func doTestRequests(client *http.Client, testServerURL string) error {
	resp, err := doRequest(client, testServerURL)
	if err != nil {
		return err
	}

	err = drainCloseResponse(resp)
	if err != nil {
		return err
	}

	if resp.Request.Context() == context.Background() {
		return fmt.Errorf("request context was not set")
	}

	resp, err = doContextRequest(context.Background(), client, testServerURL)
	if err != nil {
		return err
	}

	err = drainCloseResponse(resp)
	if err != nil {
		return err
	}

	if resp.Request.Context() == context.Background() {
		return fmt.Errorf("request context was not set")
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second)
	defer cancelCtx()

	resp, err = doContextRequest(ctx, client, testServerURL)
	if err != nil {
		return err
	}

	err = drainCloseResponse(resp)
	if err != nil {
		return err
	}

	if resp.Request.Context() == context.Background() {
		return fmt.Errorf("request context was not set")
	}

	return nil
}

func doRequest(client *http.Client, testServerURL string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, testServerURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("unable to create test request: %w", err)
	}

	return client.Do(req)
}

func doContextRequest(ctx context.Context, client *http.Client, testServerURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testServerURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("unable to create test request: %w", err)
	}

	return client.Do(req)
}

func readCloseResponse(resp *http.Response) (respBytes []byte, err error) {
	defer func() {
		err = closeResponse(resp, err)
	}()

	return io.ReadAll(resp.Body)
}

func drainCloseResponse(resp *http.Response) (err error) {
	defer func() {
		err = closeResponse(resp, err)
	}()

	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		err = fmt.Errorf("error draining test response body: %w", err)
	}

	return
}

func closeResponse(resp *http.Response, err error) error {
	closeErr := resp.Body.Close()
	if closeErr != nil {
		closeErr = fmt.Errorf("error closing test response body: %w", closeErr)

		if err != nil {
			return fmt.Errorf("%s: %w", closeErr, err)
		}

		return closeErr
	}

	return err
}
