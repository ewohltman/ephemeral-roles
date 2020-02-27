package http

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/tracer"
)

const (
	jaegerServiceName       = "ephemeral-roles"
	jaegerSampleProbability = 1
)

func TestNewClient(t *testing.T) {
	testServer := httptest.NewServer(testServerHandler())
	defer testServer.Close()

	jaegerTracer, jaegerCloser, err := tracer.New(mock.NewLogger(), jaegerServiceName, jaegerSampleProbability)
	if err != nil {
		t.Fatalf("Error setting up Jaeger tracer: %s", err)
	}

	defer func() {
		closeErr := jaegerCloser.Close()
		if closeErr != nil {
			t.Errorf("Error closing Jaeger tracer: %s", closeErr)
		}
	}()

	client := NewClient(nil, jaegerTracer)

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

func TestSetTransport(t *testing.T) {
	testServer := httptest.NewServer(testServerHandler())
	defer testServer.Close()

	log := mock.NewLogger()

	clientNilTransport := &http.Client{}

	err := testSetTransport(log, clientNilTransport, testServer.URL)
	if err != nil {
		t.Fatalf("Error testing nil http.RoundTripper: %s", err)
	}

	clientWithTransport := &http.Client{
		Transport: http.DefaultTransport,
	}

	err = testSetTransport(log, clientWithTransport, testServer.URL)
	if err != nil {
		t.Fatalf("Error testing nil http.RoundTripper: %s", err)
	}
}

func testSetTransport(log logging.Interface, client *http.Client, testServerURL string) (err error) {
	jaegerTracer, jaegerCloser, err := tracer.New(log, jaegerServiceName, jaegerSampleProbability)
	if err != nil {
		return fmt.Errorf("error setting up Jaeger tracer: %w", err)
	}

	defer func() {
		closeErr := jaegerCloser.Close()
		if closeErr != nil {
			if err != nil {
				err = fmt.Errorf("%s: error closing Jaeger tracer: %w", err, closeErr)
				return
			}

			err = fmt.Errorf("error closing Jaeger tracer: %w", closeErr)
		}
	}()

	SetTransport(client, jaegerTracer)

	if client.Transport == nil {
		return fmt.Errorf("unexpected nil http.RoundTripper")
	}

	err = doTestRequests(client, testServerURL)
	if err != nil {
		return fmt.Errorf("error doing test request: %s", err)
	}

	return nil
}

func testServerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			_, err := io.Copy(ioutil.Discard, r.Body)
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
	resp, err := doRequest(context.Background(), client, testServerURL)
	if err != nil {
		return err
	}

	err = drainCloseResponse(resp)
	if err != nil {
		return err
	}

	if reflect.DeepEqual(resp.Request.Context(), context.Background()) {
		return fmt.Errorf("request context was not set")
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), contextTimeout)
	defer cancelCtx()

	resp, err = doRequest(ctx, client, testServerURL)
	if err != nil {
		return err
	}

	err = drainCloseResponse(resp)
	if err != nil {
		return err
	}

	if reflect.DeepEqual(resp.Request.Context(), context.Background()) {
		return fmt.Errorf("request context was not set")
	}

	return nil
}

func doRequest(ctx context.Context, client *http.Client, testServerURL string) (resp *http.Response, err error) {
	var req *http.Request

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, testServerURL, nil)
	if err != nil {
		err = fmt.Errorf("unable to create test request: %w", err)
		return nil, err
	}

	resp, err = client.Do(req)
	if err != nil {
		err = fmt.Errorf("unable to perform test request: %w", err)
		return nil, err
	}

	return resp, nil
}

func readCloseResponse(resp *http.Response) (respBytes []byte, err error) {
	defer func() {
		err = closeResponse(resp, err)
	}()

	return ioutil.ReadAll(resp.Body)
}

func drainCloseResponse(resp *http.Response) (err error) {
	defer func() {
		err = closeResponse(resp, err)
	}()

	_, err = io.Copy(ioutil.Discard, resp.Body)
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
