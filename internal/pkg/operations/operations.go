// Package operations provides a centralized package for performing Discord
// Session/State operations.
package operations

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

const (
	guildMembersPageLimit = 1000

	roleHoist   = true
	roleMention = true
)

// LookupGuild returns a *discordgo.Guild from the session's internal state
// cache. If the guild is not found in the state cache, LookupGuild will query
// the Discord API for the guild and add it to the state cache before returning
// it.
func LookupGuild(ctx context.Context, session *discordgo.Session, guildID string) (*discordgo.Guild, error) {
	guild, err := session.State.Guild(guildID)
	if err != nil {
		guild, err = updateStateGuilds(ctx, session, guildID)
		if err != nil {
			return nil, fmt.Errorf("unable to query guild: %w", err)
		}
	}

	return guild, nil
}

// CreateRole will create an Ephemeral Role in the provided guild using the
// provided roleID and roleColor.
func CreateRole(
	ctx context.Context,
	session *discordgo.Session,
	guild *discordgo.Guild,
	roleName string,
	roleColor int,
) (*discordgo.Role, error) {
	role, err := session.GuildRoleCreateWithContext(ctx, guild.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to create ephemeral role: %w", err)
	}

	role, err = session.GuildRoleEditWithContext(
		ctx,
		guild.ID, role.ID,
		roleName, roleColor,
		roleHoist, role.Permissions, roleMention,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to edit ephemeral role: %w", err)
	}

	err = session.State.RoleAdd(guild.ID, role)
	if err != nil {
		return nil, fmt.Errorf("unable to add ephemeral role to state cache: %w", err)
	}

	return role, nil
}

// AddRoleToMember adds the role associated with the provided roleID to the
// user associated with the provided userID, in the guild associated with the
// provided guildID.
func AddRoleToMember(ctx context.Context, session *discordgo.Session, guildID, userID, roleID string) error {
	err := session.GuildMemberRoleAddWithContext(ctx, guildID, userID, roleID)
	if err != nil {
		return fmt.Errorf("unable to add ephemeral role: %w", err)
	}

	return nil
}

// RemoveRoleFromMember removes the role associated with the provided roleID
// from the user associated with the provided userID, in the guild associated
// with the provided guildID.
func RemoveRoleFromMember(ctx context.Context, session *discordgo.Session, guildID, userID, roleID string) error {
	err := session.GuildMemberRoleRemoveWithContext(ctx, guildID, userID, roleID)
	if err != nil {
		return fmt.Errorf("unable to remove ephemeral role: %w", err)
	}

	return nil
}

// IsForbiddenResponse checks if the provided error wraps *discordgo.RESTError.
// If it does, IsForbiddenResponse returns true if the response code is equal
// to http.StatusForbidden. Otherwise, IsForbiddenResponse returns false.
func IsForbiddenResponse(err error) bool {
	var restErr *discordgo.RESTError

	if errors.As(err, &restErr) {
		if restErr.Response.StatusCode == http.StatusForbidden {
			return true
		}
	}

	return false
}

// BotHasChannelPermission checks if the bot has view permissions for the
// channel. If the bot does have the view permission, BotHasChannelPermission
// returns nil.
func BotHasChannelPermission(ctx context.Context, session *discordgo.Session, channel *discordgo.Channel) error {
	permissions, err := session.UserChannelPermissions("@me", channel.ID)
	if err != nil {
		return fmt.Errorf("unable to determine channel permissions: %w", err)
	}

	if permissions&discordgo.PermissionViewChannel != discordgo.PermissionViewChannel {
		return fmt.Errorf("insufficient channel permissions: channel: %s", channel.Name)
	}

	return nil
}

func updateStateGuilds(ctx context.Context, session *discordgo.Session, guildID string) (*discordgo.Guild, error) {
	guild, err := session.GuildWithContext(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("error senging guild query request: %w", err)
	}

	roles, err := session.GuildRolesWithContext(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("unable to query guild channels: %w", err)
	}

	channels, err := session.GuildChannelsWithContext(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("unable to query guild channels: %w", err)
	}

	members, err := recursiveGuildMembersWithContext(ctx, session, guildID, "", guildMembersPageLimit)
	if err != nil {
		return nil, fmt.Errorf("unable to query guild members: %w", err)
	}

	guild.Roles = roles
	guild.Channels = channels
	guild.Members = members
	guild.MemberCount = len(members)

	err = session.State.GuildAdd(guild)
	if err != nil {
		return nil, fmt.Errorf("unable to add guild to state cache: %w", err)
	}

	return guild, nil
}

func recursiveGuildMembersWithContext(
	ctx context.Context,
	session *discordgo.Session,
	guildID, after string,
	limit int,
) ([]*discordgo.Member, error) {
	guildMembers, err := session.GuildMembersWithContext(ctx, guildID, after, limit)
	if err != nil {
		return nil, fmt.Errorf("error sending recursive guild members request: %w", err)
	}

	if len(guildMembers) < guildMembersPageLimit {
		return guildMembers, nil
	}

	nextGuildMembers, err := recursiveGuildMembersWithContext(
		ctx,
		session,
		guildID,
		guildMembers[len(guildMembers)-1].User.ID,
		guildMembersPageLimit,
	)
	if err != nil {
		return nil, err
	}

	return append(guildMembers, nextGuildMembers...), nil
}
