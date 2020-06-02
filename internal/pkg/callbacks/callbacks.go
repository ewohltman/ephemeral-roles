// Package callbacks provides callback implementations for Discord API events.
package callbacks

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
)

const guildMembersPageLimit = 1000

// Config contains fields for the callback methods.
type Config struct {
	Log                     logging.Interface
	BotName                 string
	BotKeyword              string
	RolePrefix              string
	RoleColor               int
	JaegerTracer            opentracing.Tracer
	ContextTimeout          time.Duration
	ReadyCounter            prometheus.Counter
	MessageCreateCounter    prometheus.Counter
	VoiceStateUpdateCounter prometheus.Counter
}

func lookupGuild(ctx context.Context, session *discordgo.Session, guildID string) (*discordgo.Guild, error) {
	guild, err := session.State.Guild(guildID)
	if err != nil {
		return queryGuild(ctx, session, guildID)
	}

	return guild, nil
}

func queryGuild(ctx context.Context, session *discordgo.Session, guildID string) (*discordgo.Guild, error) {
	guild, err := session.GuildWithContext(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("unable to query guild: %w", err)
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
		return nil, fmt.Errorf("unable to add guild to session cache: %w", err)
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
		return nil, err
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

	guildMembers = append(guildMembers, nextGuildMembers...)

	return guildMembers, nil
}

func createGuildRole(ctx context.Context, session *discordgo.Session, guildID, roleName string, roleColor int) (*discordgo.Role, error) {
	const hoist = true

	role, err := session.GuildRoleCreateWithContext(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("unable to create ephemeral role: %w", err)
	}

	role, err = session.GuildRoleEditWithContext(
		ctx,
		guildID, role.ID,
		roleName, roleColor,
		hoist, role.Permissions, role.Mentionable,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to edit ephemeral role: %w", err)
	}

	err = session.State.RoleAdd(guildID, role)
	if err != nil {
		return nil, fmt.Errorf("unable to add ephemeral role to session cache: %w", err)
	}

	return role, nil
}

func addRoleToMember(ctx context.Context, session *discordgo.Session, guildID, userID, ephemeralRoleID string) error {
	err := session.GuildMemberRoleAddWithContext(ctx, guildID, userID, ephemeralRoleID)
	if err != nil {
		return fmt.Errorf("unable to grant ephemeral role: %w", err)
	}

	return nil
}

func removeRoleFromMember(ctx context.Context, session *discordgo.Session, guildID, userID, ephemeralRoleID string) error {
	err := session.GuildMemberRoleRemoveWithContext(ctx, guildID, userID, ephemeralRoleID)
	if err != nil {
		return fmt.Errorf("unable to revoke ephemeral role: %w", err)
	}

	return nil
}
