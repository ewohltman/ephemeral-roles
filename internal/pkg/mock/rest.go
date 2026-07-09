package mock

import (
	"context"
	"errors"
	"slices"
	"sync/atomic"

	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

// errGuildNotFound is returned by the fake GetGuild when a guild is not present
// in the cache.
var errGuildNotFound = errors.New("guild not found")

// mockRest is a fake rest.Rest implementation backed by the cache. Only the
// methods exercised by the bot are overridden; any other method is inherited
// from the embedded (nil) rest.Rest interface and will panic if called, which
// keeps the fake honest about what the code under test actually uses.
type mockRest struct {
	rest.Rest

	caches     cache.Caches
	nextRoleID atomic.Uint64
}

func newMockRest(caches cache.Caches) *mockRest {
	mock := &mockRest{caches: caches}
	mock.nextRoleID.Store(uint64(TestEphemeralRole) + 1)

	return mock
}

// Close is a no-op override so bot.Client.Close does not dispatch to the
// embedded (nil) rest.Rest interface.
func (*mockRest) Close(_ context.Context) {}

// CreateRole creates a role, adds it to the cache and returns it.
//
//nolint:gocritic // signature is dictated by the rest.Rest interface
func (m *mockRest) CreateRole(
	guildID snowflake.ID,
	createRole discord.RoleCreate,
	_ ...rest.RequestOpt,
) (*discord.Role, error) {
	role := discord.Role{
		ID:          snowflake.ID(m.nextRoleID.Add(1)),
		GuildID:     guildID,
		Name:        createRole.Name,
		Color:       createRole.Color,
		Hoist:       createRole.Hoist,
		Mentionable: createRole.Mentionable,
	}

	m.caches.AddRole(role)

	return &role, nil
}

// DeleteRole removes a role from the cache.
func (m *mockRest) DeleteRole(guildID, roleID snowflake.ID, _ ...rest.RequestOpt) error {
	m.caches.RemoveRole(guildID, roleID)

	return nil
}

// AddMemberRole adds a role to a member in the cache.
func (m *mockRest) AddMemberRole(guildID, userID, roleID snowflake.ID, _ ...rest.RequestOpt) error {
	member, ok := m.caches.Member(guildID, userID)
	if !ok {
		return nil
	}

	if !slices.Contains(member.RoleIDs, roleID) {
		member.RoleIDs = append(member.RoleIDs, roleID)
		m.caches.AddMember(member)
	}

	return nil
}

// RemoveMemberRole removes a role from a member in the cache.
func (m *mockRest) RemoveMemberRole(guildID, userID, roleID snowflake.ID, _ ...rest.RequestOpt) error {
	member, ok := m.caches.Member(guildID, userID)
	if !ok {
		return nil
	}

	filtered := member.RoleIDs[:0]

	for _, id := range member.RoleIDs {
		if id != roleID {
			filtered = append(filtered, id)
		}
	}

	member.RoleIDs = filtered
	m.caches.AddMember(member)

	return nil
}

// GetGuild returns the guild from the cache wrapped in a *discord.RestGuild.
func (m *mockRest) GetGuild(
	guildID snowflake.ID,
	_ bool,
	_ ...rest.RequestOpt,
) (*discord.RestGuild, error) {
	guild, ok := m.caches.Guild(guildID)
	if !ok {
		return nil, errGuildNotFound
	}

	return &discord.RestGuild{Guild: guild}, nil
}
