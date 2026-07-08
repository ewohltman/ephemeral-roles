package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internalHTTP "github.com/ewohltman/ephemeral-roles/internal/pkg/http"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

const (
	testPort = "8081"
	testURL  = "http://localhost:" + testPort

	serverStartupDelay = 50 * time.Millisecond

	expectedGuildsFilePath = "testdata/guilds.json"
)

func TestNewServer(t *testing.T) {
	t.Parallel()

	log := mock.NewLogger()

	session, err := mock.NewSession()
	require.NoError(t, err)

	session.State.Guilds = append(
		session.State.Guilds,
		&discordgo.Guild{Name: "testGuild2", MemberCount: 3},
		&discordgo.Guild{Name: "testGuild3", MemberCount: 4},
	)

	testServer := internalHTTP.NewServer(log, session, testPort)

	go func() {
		assert.ErrorIs(t, testServer.ListenAndServe(), http.ErrServerClosed)
	}()

	time.Sleep(serverStartupDelay)

	client := internalHTTP.NewClient(internalHTTP.NewTransport())

	testRootEndpoint(t, client)
	testGuildsEndpoint(t, client)

	ctx, cancelContext := context.WithTimeout(t.Context(), time.Second)
	defer cancelContext()

	require.NoError(t, testServer.Shutdown(ctx))
}

func testRootEndpoint(t *testing.T, client *http.Client) {
	t.Helper()

	resp, err := doRequest(t.Context(), client, testURL+internalHTTP.RootEndpoint)
	require.NoError(t, err)

	drainCloseResponse(resp)
}

func testGuildsEndpoint(t *testing.T, client *http.Client) {
	t.Helper()

	expectedGuildsFile, err := os.Open(expectedGuildsFilePath)
	require.NoError(t, err)

	defer func() { _ = expectedGuildsFile.Close() }()

	expectedGuilds := make(internalHTTP.SortableGuilds, 0)
	require.NoError(t, json.NewDecoder(expectedGuildsFile).Decode(&expectedGuilds))

	resp, err := doRequest(t.Context(), client, testURL+internalHTTP.GuildsEndpoint)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	actualGuilds := make(internalHTTP.SortableGuilds, 0)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&actualGuilds))

	assert.Equal(t, expectedGuilds, actualGuilds)
}
