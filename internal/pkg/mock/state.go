package mock

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

const stateCreateErrMessage = "unable to create new state"

// NewState provides a *discordgo.State instance to be used in unit testing.
func NewState() (*discordgo.State, error) {
	state := discordgo.NewState()

	state.User = &discordgo.User{
		ID:       TestSession,
		Username: TestSession,
		Bot:      true,
	}

	err := addTestGuild(state)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", stateCreateErrMessage, err)
	}

	return state, nil
}

func addTestGuild(state *discordgo.State) error {
	testGuild := mockGuild(TestGuild)

	err := state.GuildAdd(testGuild)
	if err != nil {
		return err
	}

	return nil
}
