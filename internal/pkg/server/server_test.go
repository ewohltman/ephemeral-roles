package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

const (
	testPort = "8080"
	testURL  = "http://localhost:" + testPort

	httpClientTimeout = time.Second
	contextTimeout    = time.Second

	expectedGuildsFile = "testdata/guilds.json"
)

func TestNew(t *testing.T) {
	log := mock.NewLogger()

	session, err := mock.NewSession()
	if err != nil {
		t.Fatalf("Error obtaining mock session: %s", err)
	}

	testServer := New(log, session, testPort)

	go func() {
		serverErr := testServer.ListenAndServe()
		if !errors.Is(serverErr, http.ErrServerClosed) {
			t.Errorf("Test server error: %s", serverErr)
		}
	}()

	client := &http.Client{Timeout: httpClientTimeout}

	testRootEndpoint(t, client)
	testGuildsEndpoint(t, client)

	ctx, cancelContext := context.WithTimeout(context.Background(), contextTimeout)
	defer cancelContext()

	err = testServer.Shutdown(ctx)
	if err != nil {
		t.Errorf("Error closing test server: %s", err)
	}
}

func testRootEndpoint(t *testing.T, client *http.Client) {
	resp, err := doRequest(client, rootEndpoint)
	if err != nil {
		t.Fatal(err)
	}

	err = drainCloseResponse(resp)
	if err != nil {
		t.Fatal(err)
	}
}

func testGuildsEndpoint(t *testing.T, client *http.Client) {
	expectedGuildsBytes, err := ioutil.ReadFile(expectedGuildsFile)
	if err != nil {
		t.Fatal(err)
	}

	expectedGuilds := make(sortableGuilds, 0)

	err = json.Unmarshal(expectedGuildsBytes, &expectedGuilds)
	if err != nil {
		t.Fatalf("Error unmarshaling expected guild data: %s", err)
	}

	resp, err := doRequest(client, guildsEndpoint)
	if err != nil {
		t.Fatal(err)
	}

	actualGuildsBytes, err := readCloseResponse(resp)
	if err != nil {
		t.Fatal(err)
	}

	actualGuilds := make(sortableGuilds, 0)

	err = json.Unmarshal(actualGuildsBytes, &actualGuilds)
	if err != nil {
		t.Fatalf("Error unmarshaling actual guild data: %s", err)
	}

	if !reflect.DeepEqual(actualGuilds, expectedGuilds) {
		t.Errorf(
			"Unexpected response:\nGot:\n%s\n\nExpected:\n%s",
			string(actualGuildsBytes),
			string(expectedGuildsBytes),
		)
	}
}

func doRequest(client *http.Client, endpoint string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, testURL+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating test request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing test request: %w", err)
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
