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

const contextTimeout = 30 * time.Second

// Config contains fields for the callback methods.
type Config struct {
	Log                     logging.Interface
	BotName                 string
	BotKeyword              string
	RolePrefix              string
	RoleColor               int
	JaegerTracer            opentracing.Tracer
	ReadyCounter            prometheus.Counter
	MessageCreateCounter    prometheus.Counter
	VoiceStateUpdateCounter prometheus.Counter
}

func mapGuildRoleIDs(guildRoles discordgo.Roles) map[string]*discordgo.Role {
	guildRoleMap := make(map[string]*discordgo.Role)

	for _, role := range guildRoles {
		guildRoleMap[role.ID] = role
	}

	return guildRoleMap
}

func lookupGuild(ctx context.Context, session *discordgo.Session, guildID string) (*discordgo.Guild, error) {
	return session.GuildWithContext(ctx, guildID)
}

func lookupGuildMember(ctx context.Context, session *discordgo.Session, guildID, userID string) (*discordgo.Member, error) {
	member, err := session.GuildMemberWithContext(ctx, guildID, userID)
	if err != nil {
		return nil, fmt.Errorf("%w", &memberNotFound{err: err})
	}

	return member, nil
}

func lookupGuildRoles(ctx context.Context, session *discordgo.Session, guildID string) (discordgo.Roles, error) {
	return session.GuildRolesWithContext(ctx, guildID)
}

func lookupGuildChannel(ctx context.Context, session *discordgo.Session, guildID, channelID string) (*discordgo.Channel, error) {
	if channelID == "" {
		return nil, nil
	}

	guildChannels, err := session.GuildChannelsWithContext(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("%w", &channelNotFound{err: err})
	}

	for _, guildChannel := range guildChannels {
		if guildChannel.ID == channelID {
			return guildChannel, nil
		}
	}

	return nil, &channelNotFound{}
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

	return role, nil
}

func addRoleToMember(ctx context.Context, session *discordgo.Session, guildID, userID, roleID string) error {
	err := session.GuildMemberRoleAddWithContext(ctx, guildID, userID, roleID)
	if err != nil {
		return fmt.Errorf("unable to grant ephemeral role: %w", err)
	}

	return nil
}

func removeRoleFromMember(ctx context.Context, session *discordgo.Session, guildID, userID, roleID string) error {
	err := session.GuildMemberRoleRemoveWithContext(ctx, guildID, userID, roleID)
	if err != nil {
		return fmt.Errorf("unable to revoke ephemeral role: %w", err)
	}

	return nil
}
