package client

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	client := New()

	if client == nil {
		t.Fatal("Unexpected nil *http.Client")
	}

	if client.Transport == nil {
		t.Fatal("Unexpected nil transport http.RoundTripper")
	}

	err := doTestRequest(client)
	if err != nil {
		t.Fatalf("Error doing test request: %s", err)
	}
}

func TestSetTransport(t *testing.T) {
	client := &http.Client{}

	if client.Transport != nil {
		t.Fatal("Unexpected non-nil transport http.RoundTripper")
	}

	SetTransport(client)

	if client.Transport == nil {
		t.Fatal("Unexpected nil transport http.RoundTripper")
	}

	err := doTestRequest(client)
	if err != nil {
		t.Fatalf("Error doing test request: %s", err)
	}
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

func doTestRequest(client *http.Client) (err error) {
	testServer := httptest.NewServer(testServerHandler())
	defer testServer.Close()

	req, err := http.NewRequest(http.MethodGet, testServer.URL, nil)
	if err != nil {
		err = fmt.Errorf("unable to create test request: %w", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("unable to perform test request: %w", err)
		return
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
		return
	}

	return nil
}
