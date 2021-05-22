// Package mock provides implementations for mocking objects and endpoints for
// unit testing.
package mock

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

// roundTripperFunc allows functions to satisfy the http.RoundTripper
// interface.
type roundTripperFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements the http.RoundTripper interface.
func (rt roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}

// NewMirrorRoundTripper returns an http.RoundTripper that mirrors the request
// body in the response body.
func NewMirrorRoundTripper() http.RoundTripper {
	return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		resp := &http.Response{
			Status:     http.StatusText(http.StatusOK),
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Request:    req,
		}

		if req.Body == nil {
			resp.ContentLength = 0
			resp.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))

			return resp, nil
		}

		reqBody, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		err = req.Body.Close()
		if err != nil {
			return nil, err
		}

		resp.ContentLength = int64(len(reqBody))
		resp.Body = ioutil.NopCloser(bytes.NewReader(reqBody))

		return resp, nil
	})
}
