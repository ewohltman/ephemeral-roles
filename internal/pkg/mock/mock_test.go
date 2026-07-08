package mock_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

func TestNewMirrorRoundTripper(t *testing.T) {
	t.Parallel()

	mirror := mock.NewMirrorRoundTripper()

	reqBodyContent := []byte("Test message")
	reqBody := bytes.NewReader(reqBodyContent)

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "", reqBody)
	require.NoError(t, err)

	resp, err := mirror.RoundTrip(req)
	require.NoError(t, err)

	respBodyContent, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.NoError(t, resp.Body.Close())

	require.Equal(t, reqBodyContent, respBodyContent)
}
