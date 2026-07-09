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

	span := handler.JaegerTracer.StartSpan(voiceStateUpdate)
	defer span.Finish()

	metadata, err := handler.parseEvent(event.Client(), event.VoiceState, &event.Member)
	if err != nil {
		handler.handleParseEventError(event.Client(), err)
		return
	}

	log := handler.Log.With(
		"guild", metadata.Guild.Name,
		"member", metadata.Member.User.Username,
	)

	if metadata.EphemeralRole != nil {
		if handler.memberHasRole(metadata.Member, metadata.EphemeralRole) {
			return
		}
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

		return
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
		return nil, &ChannelNotFoundError{
			Guild:  &guild,
			Member: member,
		}
	}

	err = operations.BotHasChannelPermission(client, channel)
	if err != nil {
		return nil, &InsufficientPermissionsError{
			Guild:   &guild,
			Member:  member,
			Channel: channel,
			Err:     err,
		}
	}

	ephemeralRole, err := handler.lookupGuildRole(client, guild.ID, channel)
	if errors.Is(err, &RoleNotFoundError{}) {
		ephemeralRole, err = handler.createRole(guild.ID, handler.RoleNameFromChannel(channel.Name()))
		if err != nil {
			switch {
			case operations.IsDeadlineExceeded(err):
				return nil, &DeadlineExceededError{Guild: &guild, Member: member, Channel: channel, Err: err}
			case operations.IsForbiddenResponse(err):
				return nil, &InsufficientPermissionsError{Guild: &guild, Member: member, Channel: channel, Err: err}
			case operations.IsMaxGuildsResponse(err):
				return nil, &MaxNumberOfRolesError{Guild: &guild, Member: member, Channel: channel, Err: err}
			default:
				return nil, err
			}
		}
	}

	return &voiceStateUpdateMetadata{
		Client:        client,
		Guild:         &guild,
		Member:        member,
		Channel:       channel,
		EphemeralRole: ephemeralRole,
	}, nil
}

func (handler *Handler) handleParseEventError(client *bot.Client, err error) {
	var (
		channelNotFoundErr         *ChannelNotFoundError
		insufficientPermissionsErr *InsufficientPermissionsError
		maxNumberOfRolesErr        *MaxNumberOfRolesError
		deadlineExceededErr        *DeadlineExceededError
	)

	switch {
	case errors.As(err, &channelNotFoundErr):
		handler.logCleanup(client, channelNotFoundErr)
	case errors.As(err, &insufficientPermissionsErr):
		handler.logCleanup(client, insufficientPermissionsErr)
	case errors.As(err, &maxNumberOfRolesErr):
		handler.logCleanup(client, maxNumberOfRolesErr)
	case errors.As(err, &deadlineExceededErr):
		handler.logParseEventError(deadlineExceededErr)
	default:
		handler.Log.Error(voiceStateUpdateEventError, "error", err)
	}
}

func (handler *Handler) logCleanup(client *bot.Client, callbackError CallbackError) {
	handler.logParseEventError(callbackError)
	handler.cleanupParseEventError(client, callbackError)
}

func (handler *Handler) logParseEventError(callbackError CallbackError) {
	handler.newCallbackErrorLogger(callbackError).Debug(voiceStateUpdateEventError, "error", callbackError)
}

func (handler *Handler) cleanupParseEventError(client *bot.Client, callbackError CallbackError) {
	metadata := &voiceStateUpdateMetadata{
		Client: client,
		Guild:  callbackError.InGuild(),
		Member: callbackError.ForMember(),
	}

	if metadata.Member == nil {
		return
	}

	if err := handler.removeEphemeralRoles(metadata); err != nil {
		handler.newCallbackErrorLogger(callbackError).Debug(voiceStateUpdateEventError, "error", err)
	}
}

func (handler *Handler) newCallbackErrorLogger(callbackError CallbackError) *slog.Logger {
	log := handler.Log

	guild := callbackError.InGuild()
	if guild != nil {
		log = log.With("guild", guild.Name)
	}

	member := callbackError.ForMember()
	if member != nil {
		log = log.With("member", member.User.Username)
	}

	channel := callbackError.InChannel()
	if channel != nil {
		log = log.With("channel", channel.Name())
	}

	return log
}

func (*Handler) memberHasRole(member *discord.Member, role *discord.Role) bool {
	return slices.Contains(member.RoleIDs, role.ID)
}

func (handler *Handler) lookupGuildRole(
	client *bot.Client,
	guildID snowflake.ID,
	channel discord.GuildChannel,
) (*discord.Role, error) {
	ephemeralRoleName := handler.RoleNameFromChannel(channel.Name())

	for role := range client.Caches.Roles(guildID) {
		if role.Name == ephemeralRoleName {
			foundRole := role
			return &foundRole, nil
		}
	}

	return nil, &RoleNotFoundError{}
}

func (handler *Handler) createRole(guildID snowflake.ID, roleName string) (*discord.Role, error) {
	result := <-handler.OperationsGateway.Process(&operations.Request{
		Type: operations.CreateRole,
		CreateRole: &operations.CreateRoleRequest{
			GuildID:   guildID,
			RoleName:  roleName,
			RoleColor: handler.RoleColor,
		},
	})
	if result.Err != nil {
		return nil, result.Err
	}

	val, ok := result.Val.(discord.Role)
	if !ok {
		return nil, fmt.Errorf("unrecognized operations result type: %T", result.Val)
	}

	return &val, nil
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
