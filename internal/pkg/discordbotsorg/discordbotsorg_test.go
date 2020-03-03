package discordbotsorg

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	internalHTTP "github.com/ewohltman/ephemeral-roles/internal/pkg/http"
)

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (s roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return s(r)
}

func TestUpdate(t *testing.T) {
	client := internalHTTP.NewClient(roundTripFunc(testResponse), nil, nil)

	err := Update(client, "", "", -1)
	if err != nil {
		if !strings.HasSuffix(err.Error(), `{"error":"Unauthorized"}`) {
			t.Error(err)
		}
	}
}

func testResponse(r *http.Request) (*http.Response, error) {
	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       ioutil.NopCloser(strings.NewReader("{}")),
	}, nil
}
