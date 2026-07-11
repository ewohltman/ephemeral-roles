// Package operations provides a centralized gateway for processing requests
// on Discord API operations.
package operations

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"golang.org/x/sync/singleflight"
)

// APIErrorCodeMaxRoles is the Discord API error code for the maximum number of
// guild roles being reached.
const APIErrorCodeMaxRoles = rest.JSONErrorCodeMaximumGuildRolesReached

const (
	roleHoist   = true
	roleMention = true

	// requestTimeout bounds each Discord REST request, including the time
	// spent queued in disgo's REST rate limiter. Discord can respond to role
	// mutations with retry_after values of an hour or more; without a
	// deadline the rate limiter blocks the calling goroutine for that entire
	// window (observed wedging a shard's guild worker for 96+ minutes in
	// production). Failing fast with context.DeadlineExceeded instead lets
	// callers classify the error (KindDeadlineExceeded) and drop the
	// operation.
	requestTimeout = 1 * time.Minute
)

// RequestContext returns a context bounding a single Discord REST request,
// and its cancel function.
func RequestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), requestTimeout)
}

// Gateway is a centralized construct to process Discord API-mutating requests
// by de-duplicating identical simultaneous requests and providing the result
// to all the callers.
type Gateway struct {
	Client *bot.Client
	group  singleflight.Group
}

// NewGateway returns a new *Gateway ready to process requests.
func NewGateway(client *bot.Client) *Gateway {
	return &Gateway{Client: client}
}

// CreateRole creates a new role in the provided guild and adds it to the
// client cache. Concurrent calls for the same guild and role name are collapsed
// into a single Discord API request sharing one result.
func (gateway *Gateway) CreateRole(guildID snowflake.ID, roleName string, roleColor int) (discord.Role, error) {
	result, err, _ := gateway.group.Do(guildID.String()+"/"+roleName, func() (any, error) {
		return createRole(gateway.Client, guildID, roleName, roleColor)
	})
	if err != nil {
		return discord.Role{}, err
	}

	role, ok := result.(discord.Role)
	if !ok {
		return discord.Role{}, fmt.Errorf("unrecognized create role result type: %T", result)
	}

	return role, nil
}

// LookupGuild returns a discord.Guild from the client's cache. If the guild is
// not found in the cache, LookupGuild will query the Discord API for the guild
// and add it to the cache before returning it.
func LookupGuild(client *bot.Client, guildID snowflake.ID) (discord.Guild, error) {
	guild, ok := client.Caches.Guild(guildID)
	if ok {
		return guild, nil
	}

	ctx, cancel := RequestContext()
	defer cancel()

	restGuild, err := client.Rest.GetGuild(guildID, false, rest.WithCtx(ctx))
	if err != nil {
		return discord.Guild{}, fmt.Errorf("unable to query guild: %w", err)
	}

	client.Caches.AddGuild(restGuild.Guild)

	return restGuild.Guild, nil
}

// AddRoleToMember adds the role associated with the provided roleID to the
// user associated with the provided userID, in the guild associated with the
// provided guildID.
func AddRoleToMember(client *bot.Client, guildID, userID, roleID snowflake.ID) error {
	ctx, cancel := RequestContext()
	defer cancel()

	if err := client.Rest.AddMemberRole(guildID, userID, roleID, rest.WithCtx(ctx)); err != nil {
		return fmt.Errorf("unable to add ephemeral role: %w", err)
	}

	return nil
}

// RemoveRoleFromMember removes the role associated with the provided roleID
// from the user associated with the provided userID, in the guild associated
// with the provided guildID.
func RemoveRoleFromMember(client *bot.Client, guildID, userID, roleID snowflake.ID) error {
	ctx, cancel := RequestContext()
	defer cancel()

	if err := client.Rest.RemoveMemberRole(guildID, userID, roleID, rest.WithCtx(ctx)); err != nil {
		return fmt.Errorf("unable to remove ephemeral role: %w", err)
	}

	return nil
}

// IsDeadlineExceeded checks if the provided error wraps
// context.DeadlineExceeded.
func IsDeadlineExceeded(err error) bool {
	return errors.Is(err, context.DeadlineExceeded)
}

// IsForbiddenResponse checks if the provided error wraps *rest.Error. If it
// does, IsForbiddenResponse returns true if the response code is equal to
// http.StatusForbidden or the Discord error code indicates missing permissions.
func IsForbiddenResponse(err error) bool {
	if restErr, ok := errors.AsType[*rest.Error](err); ok {
		if restErr.Response != nil && restErr.Response.StatusCode == http.StatusForbidden {
			return true
		}
	}

	return rest.IsJSONErrorCode(err,
		rest.JSONErrorCodeMissingAccess,
		rest.JSONErrorCodeLackPermissionsToPerformAction,
	)
}

// IsMaxGuildsResponse checks if the provided error wraps *rest.Error. If it
// does, IsMaxGuildsResponse returns true if the Discord error code indicates
// the guild has reached the maximum number of roles.
func IsMaxGuildsResponse(err error) bool {
	return rest.IsJSONErrorCode(err, APIErrorCodeMaxRoles)
}

// ShouldLogDebug checks if the provided error should be logged at a debug
// level.
func ShouldLogDebug(err error) bool {
	switch {
	case IsDeadlineExceeded(err), IsForbiddenResponse(err):
		return true
	default:
		return false
	}
}

// BotHasChannelPermission checks if the bot has view permissions for the
// channel. If the bot does have the view permission, BotHasChannelPermission
// returns nil.
func BotHasChannelPermission(client *bot.Client, channel discord.GuildChannel) error {
	selfMember, ok := client.Caches.SelfMember(channel.GuildID())
	if !ok {
		return errors.New("unable to determine channel permissions: bot member not found in cache")
	}

	permissions := client.Caches.MemberPermissionsInChannel(channel, selfMember)

	if permissions&discord.PermissionViewChannel != discord.PermissionViewChannel {
		return fmt.Errorf("insufficient channel permissions: channel: %s", channel.Name())
	}

	return nil
}

func createRole(
	client *bot.Client,
	guildID snowflake.ID,
	roleName string,
	roleColor int,
) (discord.Role, error) {
	ctx, cancel := RequestContext()
	defer cancel()

	role, err := client.Rest.CreateRole(guildID, discord.RoleCreate{
		Name:        roleName,
		Color:       roleColor,
		Hoist:       roleHoist,
		Mentionable: roleMention,
	}, rest.WithCtx(ctx))
	if err != nil {
		return discord.Role{}, fmt.Errorf("unable to create ephemeral role: %w", err)
	}

	client.Caches.AddRole(*role)

	return *role, nil
}
