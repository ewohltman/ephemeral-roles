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

type roleID string

type roleIDMap map[roleID]*discordgo.Role

func mapGuildRoleIDs(guildRoles discordgo.Roles) roleIDMap {
	guildRoleMap := make(roleIDMap)

	for _, role := range guildRoles {
		guildRoleMap[roleID(role.ID)] = role
	}

	return guildRoleMap
}

func lookupGuild(ctx context.Context, session *discordgo.Session, guildID string) (*discordgo.Guild, error) {
	guild, err := session.GuildWithContext(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", guildNotFoundMessage, err)
	}

	err = session.State.GuildAdd(guild)
	if err != nil {
		return nil, fmt.Errorf("unable to add guild to session cache: %w", err)
	}

	return guild, nil
}

func lookupGuildMember(ctx context.Context, session *discordgo.Session, guildID, userID string) (*discordgo.Member, error) {
	guildMember, err := session.State.Member(guildID, userID)
	if err != nil {
		return queryGuildMember(ctx, session, guildID, userID)
	}

	return guildMember, nil
}

func queryGuildMember(ctx context.Context, session *discordgo.Session, guildID, userID string) (*discordgo.Member, error) {
	guildMember, err := session.GuildMemberWithContext(ctx, guildID, userID)
	if err != nil {
		return nil, fmt.Errorf("%w", &memberNotFound{err: err})
	}

	err = session.State.MemberAdd(guildMember)
	if err != nil {
		return nil, fmt.Errorf("unable to add guild member to session cache: %w", err)
	}

	return guildMember, nil
}

func lookupGuildChannel(ctx context.Context, session *discordgo.Session, guildID, channelID string) (*discordgo.Channel, error) {
	if channelID == "" {
		return nil, nil
	}

	channel, err := session.State.Channel(channelID)
	if err != nil {
		return queryGuildChannel(ctx, session, guildID, channelID)
	}

	return channel, nil
}

func queryGuildChannel(ctx context.Context, session *discordgo.Session, guildID, channelID string) (*discordgo.Channel, error) {
	guildChannels, err := session.GuildChannelsWithContext(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("%w", &channelNotFound{err: err})
	}

	var channel *discordgo.Channel

	for _, guildChannel := range guildChannels {
		err = session.State.ChannelAdd(guildChannel)
		if err != nil {
			return nil, fmt.Errorf("unable to add channel to session cache: %w", err)
		}

		if guildChannel.ID == channelID {
			channel = guildChannel
		}
	}

	if channel == nil {
		return nil, &channelNotFound{}
	}

	return channel, nil
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
