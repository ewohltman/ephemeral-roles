package mock

import (
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/discordgo-mock/mockchannel"
	"github.com/ewohltman/discordgo-mock/mockconstants"
	"github.com/ewohltman/discordgo-mock/mockguild"
	"github.com/ewohltman/discordgo-mock/mockmember"
	"github.com/ewohltman/discordgo-mock/mockrest"
	"github.com/ewohltman/discordgo-mock/mockrole"
	"github.com/ewohltman/discordgo-mock/mocksession"
	"github.com/ewohltman/discordgo-mock/mockstate"
	"github.com/ewohltman/discordgo-mock/mockuser"
)

const (
	rolePrefix     = "{eph}"
	testUserBot    = mockconstants.TestUser + "Bot"
	largeGuildSize = 3000
)

// NewSession provides a *discordgo.Session instance to be used in unit
// testing with pre-populated initial state.
func NewSession() (*discordgo.Session, error) {
	role := mockrole.New(
		mockrole.WithID(mockconstants.TestRole),
		mockrole.WithName(mockconstants.TestRole),
		mockrole.WithPermissions(discordgo.PermissionViewChannel),
	)

	ephRole := mockrole.New(
		mockrole.WithID(fmt.Sprintf("%s %s", rolePrefix, mockconstants.TestChannel)),
		mockrole.WithName(fmt.Sprintf("%s %s", rolePrefix, mockconstants.TestChannel)),
		mockrole.WithPermissions(discordgo.PermissionViewChannel),
	)

	botUser := mockuser.New(
		mockuser.WithID(testUserBot),
		mockuser.WithUsername(testUserBot),
		mockuser.WithBotFlag(true),
	)

	state, err := mockstate.New(
		mockstate.WithUser(botUser),
		mockstate.WithGuilds(
			smallGuild(botUser, role, ephRole),
			largeGuild(botUser, role, ephRole),
		),
	)
	if err != nil {
		return nil, err
	}

	return mocksession.New(
		mocksession.WithState(state),
		mocksession.WithClient(&http.Client{
			Transport: mockrest.NewTransport(state),
		}),
	)
}

func smallGuild(botUser *discordgo.User, role, ephRole *discordgo.Role) *discordgo.Guild {
	botMember := mockmember.New(
		mockmember.WithUser(botUser),
		mockmember.WithGuildID(mockconstants.TestGuild),
		mockmember.WithRoles(role, ephRole),
	)

	userMember := mockmember.New(
		mockmember.WithUser(mockuser.New(
			mockuser.WithID(mockconstants.TestUser),
			mockuser.WithUsername(mockconstants.TestUser),
		)),
		mockmember.WithGuildID(mockconstants.TestGuild),
		mockmember.WithRoles(role, ephRole),
	)

	channel1 := mockchannel.New(
		mockchannel.WithID(mockconstants.TestChannel),
		mockchannel.WithGuildID(mockconstants.TestGuild),
		mockchannel.WithName(mockconstants.TestChannel),
		mockchannel.WithType(discordgo.ChannelTypeGuildVoice),
	)

	channel2 := mockchannel.New(
		mockchannel.WithID(mockconstants.TestChannel2),
		mockchannel.WithGuildID(mockconstants.TestGuild),
		mockchannel.WithName(mockconstants.TestChannel2),
		mockchannel.WithType(discordgo.ChannelTypeGuildVoice),
	)

	privateChannel := mockchannel.New(
		mockchannel.WithID(mockconstants.TestPrivateChannel),
		mockchannel.WithGuildID(mockconstants.TestGuild),
		mockchannel.WithName(mockconstants.TestPrivateChannel),
		mockchannel.WithType(discordgo.ChannelTypeGuildVoice),
		mockchannel.WithPermissionOverwrites(&discordgo.PermissionOverwrite{
			ID:   botMember.User.ID,
			Type: discordgo.PermissionOverwriteTypeMember,
			Deny: discordgo.PermissionViewChannel,
		}),
	)

	return mockguild.New(
		mockguild.WithID(mockconstants.TestGuild),
		mockguild.WithName(mockconstants.TestGuild),
		mockguild.WithRoles(role, ephRole),
		mockguild.WithChannels(channel1, channel2, privateChannel),
		mockguild.WithMembers(botMember, userMember),
	)
}

func largeGuild(botUser *discordgo.User, role, ephRole *discordgo.Role) *discordgo.Guild {
	guild := smallGuild(botUser, role, ephRole)

	largeGuildMembers := make([]*discordgo.Member, largeGuildSize)

	for i := 0; i < largeGuildSize; i++ {
		largeGuildMembers[i] = mockmember.New(
			mockmember.WithUser(mockuser.New(
				mockuser.WithID(fmt.Sprintf("%s%d", mockconstants.TestUser, i)),
				mockuser.WithUsername(fmt.Sprintf("%s%d", mockconstants.TestUser, i)),
			)),
			mockmember.WithGuildID(mockconstants.TestGuildLarge),
			mockmember.WithRoles(role, ephRole),
		)
	}

	guild.ID = mockconstants.TestGuildLarge
	guild.Name = mockconstants.TestGuildLarge
	guild.Members = append(guild.Members, largeGuildMembers...)
	guild.MemberCount = len(guild.Members)

	return guild
}
