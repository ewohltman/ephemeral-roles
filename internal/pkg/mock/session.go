package mock

import (
	"fmt"
	"testing"

	"github.com/bwmarrin/discordgo"
)

// NewSession provides a *discordgo.Session instance to be used in unit
// testing with pre-populated initial state.
func NewSession() (*discordgo.Session, error) {
	state, err := NewState()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", sessionCreateErrMessage, err)
	}

	session := &discordgo.Session{
		State:        state,
		StateEnabled: true,
		Ratelimiter:  discordgo.NewRatelimiter(),
		Client:       restClient(),
	}

	return session, nil
}

// NewSessionEmptyState provides a *discordgo.Session instance to be used in
// unit testing with an empty initial state.
func NewSessionEmptyState() (*discordgo.Session, error) {
	session := &discordgo.Session{
		State:        discordgo.NewState(),
		StateEnabled: true,
		Ratelimiter:  discordgo.NewRatelimiter(),
		Client:       restClient(),
	}

	return session, nil
}

// SessionClose closes a *discordgo.Session instance and if an error is encountered,
// the provided testingInstance logs the error and marks the test as failed.
func SessionClose(testingInstance testing.TB, session *discordgo.Session) {
	err := session.Close()
	if err != nil {
		testingInstance.Error(err)
	}
}
