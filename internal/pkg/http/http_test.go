package http_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func doRequest(ctx context.Context, client *http.Client, testServerURL string) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testServerURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("unable to create test request: %w", err)
	}

	return client.Do(req)
}

func drainCloseResponse(resp *http.Response) {
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}
