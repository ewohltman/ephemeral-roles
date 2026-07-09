package callbacks

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/operations"
)

const (
	voiceStateUpdate           = "VoiceStateUpdate"
	voiceStateUpdateEventError = unableToProcessEvent + voiceStateUpdate
)

type voiceStateUpdateMetadata struct {
	Client        *bot.Client
	Guild         *discord.Guild
	Member        *discord.Member
	Channel       discord.GuildChannel
	EphemeralRole *discord.Role
}

// VoiceStateUpdate is the callback function for the VoiceStateUpdate event from Discord.
func (handler *Handler) VoiceStateUpdate(event *events.GuildVoiceStateUpdate) {
	handler.VoiceStateUpdateCounter.Inc()

	metadata, err := handler.parseEvent(event.Client(), event.VoiceState, &event.Member)
	if err != nil {
		handler.handleParseEventError(event.Client(), err)
		return
	}

	log := handler.Log.With(
		"guild", metadata.Guild.Name,
		"member", metadata.Member.User.Username,
	)

	if metadata.EphemeralRole != nil && slices.Contains(metadata.Member.RoleIDs, metadata.EphemeralRole.ID) {
		return
	}

	if err := handler.removeEphemeralRoles(metadata); err != nil {
		log.Error(voiceStateUpdateEventError, "error", err)
	}

	if metadata.Channel == nil {
		return
	}

	if err := handler.addEphemeralRole(metadata); err != nil {
		if operations.ShouldLogDebug(err) {
			log.Debug(voiceStateUpdateEventError, "error", err)
			return
		}

		log.Error(voiceStateUpdateEventError, "error", err)
	}
}

func (handler *Handler) parseEvent(
	client *bot.Client,
	voiceState discord.VoiceState,
	member *discord.Member,
) (*voiceStateUpdateMetadata, error) {
	guild, err := operations.LookupGuild(client, voiceState.GuildID)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup Guild: %w", err)
	}

	if voiceState.ChannelID == nil {
		return &voiceStateUpdateMetadata{
			Client: client,
			Guild:  &guild,
			Member: member,
		}, nil
	}

	channel, ok := client.Caches.Channel(*voiceState.ChannelID)
	if !ok {
		return nil, &EventError{Kind: KindChannelNotFound, Guild: &guild, Member: member}
	}

	if err := operations.BotHasChannelPermission(client, channel); err != nil {
		return nil, &EventError{Kind: KindInsufficientPermissions, Guild: &guild, Member: member, Channel: channel, Err: err}
	}

	ephemeralRole, err := handler.ephemeralRoleForChannel(client, &guild, member, channel)
	if err != nil {
		return nil, err
	}

	return &voiceStateUpdateMetadata{
		Client:        client,
		Guild:         &guild,
		Member:        member,
		Channel:       channel,
		EphemeralRole: ephemeralRole,
	}, nil
}

func (handler *Handler) ephemeralRoleForChannel(
	client *bot.Client,
	guild *discord.Guild,
	member *discord.Member,
	channel discord.GuildChannel,
) (*discord.Role, error) {
	ephemeralRoleName := handler.RoleNameFromChannel(channel.Name())

	if role, ok := lookupGuildRole(client, guild.ID, ephemeralRoleName); ok {
		return &role, nil
	}

	role, err := handler.OperationsGateway.CreateRole(guild.ID, ephemeralRoleName, handler.RoleColor)
	if err != nil {
		eventErr := &EventError{Guild: guild, Member: member, Channel: channel, Err: err}

		switch {
		case operations.IsDeadlineExceeded(err):
			eventErr.Kind = KindDeadlineExceeded
		case operations.IsForbiddenResponse(err):
			eventErr.Kind = KindInsufficientPermissions
		case operations.IsMaxGuildsResponse(err):
			eventErr.Kind = KindMaxNumberOfRoles
		default:
			return nil, err
		}

		return nil, eventErr
	}

	return &role, nil
}

func (handler *Handler) handleParseEventError(client *bot.Client, err error) {
	eventErr, ok := errors.AsType[*EventError](err)
	if !ok {
		handler.Log.Error(voiceStateUpdateEventError, "error", err)
		return
	}

	log := handler.newEventErrorLogger(eventErr)

	log.Debug(voiceStateUpdateEventError, "error", eventErr)

	if eventErr.Kind == KindDeadlineExceeded || eventErr.Guild == nil || eventErr.Member == nil {
		return
	}

	metadata := &voiceStateUpdateMetadata{
		Client: client,
		Guild:  eventErr.Guild,
		Member: eventErr.Member,
	}

	if err := handler.removeEphemeralRoles(metadata); err != nil {
		log.Debug(voiceStateUpdateEventError, "error", err)
	}
}

func (handler *Handler) newEventErrorLogger(eventErr *EventError) *slog.Logger {
	log := handler.Log

	if eventErr.Guild != nil {
		log = log.With("guild", eventErr.Guild.Name)
	}

	if eventErr.Member != nil {
		log = log.With("member", eventErr.Member.User.Username)
	}

	if eventErr.Channel != nil {
		log = log.With("channel", eventErr.Channel.Name())
	}

	return log
}

func lookupGuildRole(client *bot.Client, guildID snowflake.ID, roleName string) (discord.Role, bool) {
	for role := range client.Caches.Roles(guildID) {
		if role.Name == roleName {
			return role, true
		}
	}

	return discord.Role{}, false
}

func (*Handler) addEphemeralRole(metadata *voiceStateUpdateMetadata) error {
	return operations.AddRoleToMember(metadata.Client, metadata.Guild.ID, metadata.Member.User.ID, metadata.EphemeralRole.ID)
}

func (handler *Handler) removeEphemeralRoles(metadata *voiceStateUpdateMetadata) error {
	var err error

	for _, roleID := range metadata.Member.RoleIDs {
		err = errors.Join(err, handler.removeEphemeralRole(metadata, roleID))
	}

	return err
}

func (handler *Handler) removeEphemeralRole(metadata *voiceStateUpdateMetadata, roleID snowflake.ID) error {
	role, ok := metadata.Client.Caches.Role(metadata.Guild.ID, roleID)
	if !ok {
		return nil
	}

	if !strings.HasPrefix(role.Name, handler.RolePrefix) {
		return nil
	}

	if err := operations.RemoveRoleFromMember(metadata.Client, metadata.Guild.ID, metadata.Member.User.ID, role.ID); err != nil {
		if !operations.IsForbiddenResponse(err) {
			return err
		}
	}

	return nil
}
