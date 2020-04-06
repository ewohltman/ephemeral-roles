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

	err = addLargeTestGuild(state)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", stateCreateErrMessage, err)
	}

	return state, nil
}

func addTestGuild(state *discordgo.State) error {
	testGuild, err := addGuild(state, TestGuild)
	if err != nil {
		return err
	}

	err = addChannels(state, testGuild, mockChannels()...)
	if err != nil {
		return err
	}

	err = addRoles(state, testGuild, mockRoles())
	if err != nil {
		return err
	}

	err = addMembers(state, testGuild, mockMembers()...)
	if err != nil {
		return err
	}

	return nil
}

func addLargeTestGuild(state *discordgo.State) error {
	testGuild, err := addGuild(state, TestGuild+"2")
	if err != nil {
		return err
	}

	err = addChannels(state, testGuild, mockChannels()...)
	if err != nil {
		return err
	}

	err = addRoles(state, testGuild, mockRoles())
	if err != nil {
		return err
	}

	testMembers := make([]*discordgo.Member, largeMemberCount)

	for i := 0; i < largeMemberCount; i++ {
		testMembers[i] = mockMember(fmt.Sprintf("%s-%d", TestUser, i))
	}

	err = addMembers(state, testGuild, testMembers...)
	if err != nil {
		return err
	}

	return nil
}

func addGuild(state *discordgo.State, guildID string) (*discordgo.Guild, error) {
	guild := mockGuild(guildID)

	return guild, state.GuildAdd(guild)
}

func addChannels(state *discordgo.State, guild *discordgo.Guild, channels ...*discordgo.Channel) error {
	for _, channel := range channels {
		channel.GuildID = guild.ID

		guild.Channels = append(guild.Channels, channel)

		err := state.ChannelAdd(channel)
		if err != nil {
			return err
		}
	}

	return nil
}

func addRoles(state *discordgo.State, guild *discordgo.Guild, roles discordgo.Roles) error {
	guild.Roles = append(guild.Roles, roles...)

	for _, role := range guild.Roles {
		err := state.RoleAdd(guild.ID, role)
		if err != nil {
			return err
		}
	}

	return nil
}

func addMembers(state *discordgo.State, guild *discordgo.Guild, members ...*discordgo.Member) error {
	for _, member := range members {
		member.GuildID = guild.ID

		guild.Members = append(guild.Members, member)
		guild.MemberCount++

		err := state.MemberAdd(member)
		if err != nil {
			return err
		}
	}

	return nil
}
