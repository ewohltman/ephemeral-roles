// Package mock provides implementations for mocking objects and endpoints for
// unit testing.
package mock

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

// Fixture snowflake IDs used to build the mock cache. They replace the string
// identifiers previously provided by github.com/ewohltman/discordgo-mock.
const (
	TestGuild          snowflake.ID = 1000
	TestGuildLarge     snowflake.ID = 2000
	TestChannel        snowflake.ID = 1001
	TestChannel2       snowflake.ID = 1002
	TestPrivateChannel snowflake.ID = 1003
	TestRole           snowflake.ID = 1004
	TestEphemeralRole  snowflake.ID = 1005
	TestUser           snowflake.ID = 1006
	TestUserBot        snowflake.ID = 1007
)

// Fixture names used to build the mock cache.
const (
	rolePrefix = "{eph}"

	TestGuildName          = "testGuild"
	TestGuildLargeName     = "testGuildLarge"
	TestChannelName        = "testChannel"
	TestChannel2Name       = "testChannel2"
	TestPrivateChannelName = "testPrivateChannel"
	TestRoleName           = "testRole"
	TestUserName           = "testUser"
	TestUserBotName        = "testUserBot"

	// EphemeralRoleName is the name of the ephemeral role associated with
	// TestChannel.
	EphemeralRoleName = rolePrefix + " " + TestChannelName

	largeGuildSize = 3000
)

// NewSession provides a *bot.Client instance to be used in unit testing with a
// pre-populated cache and a fake REST client.
func NewSession() (*bot.Client, error) {
	caches := cache.New(cache.WithCaches(
		cache.FlagGuilds,
		cache.FlagChannels,
		cache.FlagRoles,
		cache.FlagVoiceStates,
		cache.FlagMembers,
	))

	caches.SetSelfUser(discord.OAuth2User{
		User: discord.User{
			ID:       TestUserBot,
			Username: TestUserBotName,
			Bot:      true,
		},
	})

	if err := addGuild(caches, TestGuild, TestGuildName, false); err != nil {
		return nil, err
	}

	if err := addGuild(caches, TestGuildLarge, TestGuildLargeName, true); err != nil {
		return nil, err
	}

	client, err := disgo.New(mockToken(),
		bot.WithCaches(caches),
		bot.WithRest(newMockRest(caches)),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// mockToken builds a syntactically valid bot token. disgo derives the
// application ID from the token's first '.'-separated segment, which must be
// the raw-standard base64 encoding of the application ID's string form.
func mockToken() string {
	appID := base64.RawStdEncoding.EncodeToString([]byte(TestUserBot.String()))

	return appID + ".mock.token"
}

func addGuild(caches cache.Caches, guildID snowflake.ID, guildName string, large bool) error {
	addRoles(caches, guildID)
	addMembers(caches, guildID, large)

	if err := addChannels(caches, guildID); err != nil {
		return err
	}

	memberCount := 2
	if large {
		memberCount += largeGuildSize
	}

	caches.AddGuild(discord.Guild{
		ID:          guildID,
		Name:        guildName,
		MemberCount: memberCount,
	})

	return nil
}

func addRoles(caches cache.Caches, guildID snowflake.ID) {
	// The role whose ID equals the guild ID is the @everyone role; its
	// permissions form the base permissions of every member.
	caches.AddRole(discord.Role{
		ID:          guildID,
		GuildID:     guildID,
		Name:        "@everyone",
		Permissions: discord.PermissionViewChannel,
	})

	caches.AddRole(discord.Role{
		ID:          TestRole,
		GuildID:     guildID,
		Name:        TestRoleName,
		Permissions: discord.PermissionViewChannel,
	})

	caches.AddRole(discord.Role{
		ID:          TestEphemeralRole,
		GuildID:     guildID,
		Name:        EphemeralRoleName,
		Permissions: discord.PermissionViewChannel,
	})
}

func addMembers(caches cache.Caches, guildID snowflake.ID, large bool) {
	roleIDs := []snowflake.ID{TestRole, TestEphemeralRole}

	caches.AddMember(discord.Member{
		GuildID: guildID,
		User:    discord.User{ID: TestUserBot, Username: TestUserBotName, Bot: true},
		RoleIDs: roleIDs,
	})

	caches.AddMember(discord.Member{
		GuildID: guildID,
		User:    discord.User{ID: TestUser, Username: TestUserName},
		RoleIDs: roleIDs,
	})

	if !large {
		return
	}

	for i := range largeGuildSize {
		memberID := snowflake.ID(uint64(TestUser) + uint64(i) + 1)

		caches.AddMember(discord.Member{
			GuildID: guildID,
			User:    discord.User{ID: memberID, Username: fmt.Sprintf("%s%d", TestUserName, i)},
			RoleIDs: roleIDs,
		})
	}
}

func addChannels(caches cache.Caches, guildID snowflake.ID) error {
	channels := []struct {
		id      snowflake.ID
		name    string
		denyBot bool
	}{
		{TestChannel, TestChannelName, false},
		{TestChannel2, TestChannel2Name, false},
		{TestPrivateChannel, TestPrivateChannelName, true},
	}

	for _, ch := range channels {
		channel, err := newVoiceChannel(ch.id, guildID, ch.name, ch.denyBot)
		if err != nil {
			return err
		}

		caches.AddChannel(channel)
	}

	return nil
}

func newVoiceChannel(id, guildID snowflake.ID, name string, denyBot bool) (discord.GuildVoiceChannel, error) {
	overwrites := ""
	if denyBot {
		overwrites = fmt.Sprintf(
			`,"permission_overwrites":[{"type":%d,"id":"%d","allow":"0","deny":"%d"}]`,
			discord.PermissionOverwriteTypeMember, TestUserBot, discord.PermissionViewChannel,
		)
	}

	raw := fmt.Sprintf(
		`{"id":"%d","guild_id":"%d","name":%q,"type":%d%s}`,
		id, guildID, name, discord.ChannelTypeGuildVoice, overwrites,
	)

	var channel discord.GuildVoiceChannel
	if err := json.Unmarshal([]byte(raw), &channel); err != nil {
		return discord.GuildVoiceChannel{}, fmt.Errorf("unable to build mock voice channel: %w", err)
	}

	return channel, nil
}
