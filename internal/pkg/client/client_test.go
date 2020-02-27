package client

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	testServer := httptest.NewServer(testServerHandler())
	defer testServer.Close()

	client := New()

	if client == nil {
		t.Fatal("Unexpected nil *http.Client")
	}

	if client.Transport == nil {
		t.Fatal("Unexpected nil http.RoundTripper")
	}

	err := doTestRequests(client, testServer.URL)
	if err != nil {
		t.Fatalf("Error doing test request: %s", err)
	}
}

func TestSetTransport(t *testing.T) {
	testServer := httptest.NewServer(testServerHandler())
	defer testServer.Close()

	clientNilTransport := &http.Client{}

	err := testSetTransport(clientNilTransport, testServer.URL)
	if err != nil {
		t.Fatalf("Error testing nil http.RoundTripper: %s", err)
	}

	clientWithTransport := &http.Client{
		Transport: http.DefaultTransport,
	}

	err = testSetTransport(clientWithTransport, testServer.URL)
	if err != nil {
		t.Fatalf("Error testing nil http.RoundTripper: %s", err)
	}
}

func testSetTransport(client *http.Client, testServerURL string) error {
	SetTransport(client)

	if client.Transport == nil {
		return fmt.Errorf("unexpected nil http.RoundTripper")
	}

	err := doTestRequests(client, testServerURL)
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
	resp, err := doRequest(context.Background(), client, testServerURL) //nolint:bodyclose // body is closed
	if err != nil {
		return err
	}

	if reflect.DeepEqual(resp.Request.Context(), context.Background()) {
		return fmt.Errorf("request context was not set")
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), contextTimeout)
	defer cancelCtx()

	resp, err = doRequest(ctx, client, testServerURL) //nolint:bodyclose // body is closed
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

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr == nil {
			return
		}

		if err != nil {
			err = fmt.Errorf(
				"%s: unable to close test response body: %w",
				err,
				closeErr,
			)

			return
		}

		err = closeErr
	}()

	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		err = fmt.Errorf("unable to drain test response body: %w", err)
		return nil, err
	}

	return resp, nil
}
