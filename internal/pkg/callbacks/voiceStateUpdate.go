package callbacks

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/operations"
)

const (
	voiceStateUpdate           = "VoiceStateUpdate"
	voiceStateUpdateEventError = "Unable to process event: " + voiceStateUpdate
)

type voiceStateUpdateMetadata struct {
	Session       *discordgo.Session
	Guild         *discordgo.Guild
	Member        *discordgo.Member
	Channel       *discordgo.Channel
	EphemeralRole *discordgo.Role
}

// VoiceStateUpdate is the callback function for the VoiceStateUpdate event from Discord.
func (handler *Handler) VoiceStateUpdate(session *discordgo.Session, voiceState *discordgo.VoiceStateUpdate) {
	handler.VoiceStateUpdateCounter.Inc()

	span := handler.JaegerTracer.StartSpan(voiceStateUpdate)
	defer span.Finish()

	ctx, cancelCtx := context.WithTimeout(context.Background(), handler.ContextTimeout)
	defer cancelCtx()

	ctx = opentracing.ContextWithSpan(ctx, span)

	metadata, err := handler.parseEvent(ctx, session, voiceState)
	if err != nil {
		handler.handleParseEventError(ctx, session, err)
		return
	}

	log := handler.Log.WithFields(
		logrus.Fields{
			"guild":  metadata.Guild.Name,
			"member": metadata.Member.User.Username,
		},
	)

	if handler.memberHasRole(metadata.Member, metadata.EphemeralRole) {
		return
	}

	err = handler.removeEphemeralRoles(ctx, metadata)
	if err != nil {
		log.WithError(err).Error(voiceStateUpdateEventError)
	}

	log.Debug("Removed Ephemeral Roles")

	err = handler.addEphemeralRole(ctx, metadata)
	if err != nil {
		if operations.IsForbiddenResponse(err) {
			log.WithError(err).Debug(voiceStateUpdateEventError)
			return
		}

		log.WithError(err).Error(voiceStateUpdateEventError)

		return
	}

	log.WithField("role", metadata.EphemeralRole.Name).Debug("Added Ephemeral Role")
}

func (handler *Handler) parseEvent(
	ctx context.Context,
	session *discordgo.Session,
	voiceState *discordgo.VoiceStateUpdate,
) (
	*voiceStateUpdateMetadata,
	error,
) {
	guild, err := operations.LookupGuild(ctx, session, voiceState.GuildID)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup Guild: %w", err)
	}

	member, err := session.State.Member(voiceState.GuildID, voiceState.UserID)
	if err != nil {
		return nil, &MemberNotFound{
			Guild: guild,
			Err:   err,
		}
	}

	channel, err := session.State.Channel(voiceState.ChannelID)
	if err != nil {
		return nil, &ChannelNotFound{
			Guild:  guild,
			Member: member,
			Err:    err,
		}
	}

	err = operations.BotHasChannelPermission(ctx, session, channel)
	if err != nil {
		return nil, &InsufficientPermissions{
			Guild:   guild,
			Member:  member,
			Channel: channel,
			Err:     err,
		}
	}

	ephemeralRole, err := handler.lookupGuildRole(guild, channel)
	if err != nil {
		if !errors.Is(err, &RoleNotFound{}) {
			if operations.IsForbiddenResponse(err) {
				return nil, &InsufficientPermissions{
					Guild:   guild,
					Member:  member,
					Channel: channel,
					Err:     err,
				}
			}

			return nil, err
		}

		ephemeralRole, err = handler.createRole(ctx, guild, handler.RoleNameFromChannel(channel.Name))
		if err != nil {
			if operations.IsForbiddenResponse(err) {
				return nil, &InsufficientPermissions{
					Guild:   guild,
					Member:  member,
					Channel: channel,
					Err:     err,
				}
			}

			return nil, err
		}
	}

	return &voiceStateUpdateMetadata{
		Session:       session,
		Guild:         guild,
		Member:        member,
		Channel:       channel,
		EphemeralRole: ephemeralRole,
	}, nil
}

func (handler *Handler) handleParseEventError(ctx context.Context, session *discordgo.Session, err error) {
	var (
		memberNotFoundErr          *MemberNotFound
		channelNotFoundErr         *ChannelNotFound
		insufficientPermissionsErr *InsufficientPermissions
	)

	switch {
	case errors.As(err, &memberNotFoundErr):
		handler.logRemove(ctx, session, memberNotFoundErr)
	case errors.As(err, &channelNotFoundErr):
		handler.logRemove(ctx, session, channelNotFoundErr)
	case errors.As(err, &insufficientPermissionsErr):
		handler.logRemove(ctx, session, insufficientPermissionsErr)
	default:
		handler.Log.WithError(err).Error(voiceStateUpdateEventError)
	}
}

func (handler *Handler) logRemove(ctx context.Context, session *discordgo.Session, callbackError CallbackError) {
	log := logrus.NewEntry(handler.Log.WrappedLogger())

	guild := callbackError.InGuild()
	if guild != nil {
		log = log.WithField("guild", guild.Name)
	}

	member := callbackError.ForMember()
	if member != nil {
		log = log.WithField("member", member.User.Username)
	}

	channel := callbackError.InChannel()
	if channel != nil {
		log = log.WithField("channel", channel.Name)
	}

	log.WithError(callbackError).Debug(voiceStateUpdateEventError)

	if member == nil {
		return
	}

	metadata := &voiceStateUpdateMetadata{
		Session: session,
		Guild:   callbackError.InGuild(),
		Member:  callbackError.ForMember(),
	}

	err := handler.removeEphemeralRoles(ctx, metadata)
	if err != nil {
		log.WithError(err).Debug(voiceStateUpdateEventError)
	}
}

func (handler *Handler) memberHasRole(member *discordgo.Member, role *discordgo.Role) bool {
	memberRoles := make([]string, len(member.Roles))

	copy(memberRoles, member.Roles)

	sort.Slice(memberRoles, func(i, j int) bool {
		return memberRoles[i] < memberRoles[j]
	})

	index := sort.Search(len(memberRoles), func(i int) bool {
		return memberRoles[i] >= role.ID
	})

	return index != len(memberRoles)
}

func (handler *Handler) lookupGuildRole(guild *discordgo.Guild, channel *discordgo.Channel) (*discordgo.Role, error) {
	guildRoles := make([]*discordgo.Role, len(guild.Roles))
	ephemeralRoleName := handler.RoleNameFromChannel(channel.Name)

	copy(guildRoles, guild.Roles)

	sort.Slice(guildRoles, func(i, j int) bool {
		return guildRoles[i].Name < guildRoles[j].Name
	})

	index := sort.Search(len(guildRoles), func(i int) bool {
		return guildRoles[i].Name >= ephemeralRoleName
	})

	if index != len(guildRoles) {
		return guildRoles[index], nil
	}

	return nil, &RoleNotFound{}
}

func (handler *Handler) createRole(ctx context.Context, guild *discordgo.Guild, roleName string) (*discordgo.Role, error) {
	resultChannel := operations.NewResultChannel()

	handler.OperationsNexus.Process(ctx, resultChannel, &operations.Request{
		Type: operations.CreateRole,
		CreateRole: &operations.CreateRoleRequest{
			Guild:     guild,
			RoleName:  roleName,
			RoleColor: handler.RoleColor,
		},
	})

	result := <-resultChannel

	switch typedResult := result.(type) {
	case *discordgo.Role:
		return typedResult, nil
	case error:
		return nil, typedResult
	default:
		return nil, fmt.Errorf("unrecognized operations result type: %T", typedResult)
	}
}

func (handler *Handler) addEphemeralRole(ctx context.Context, metadata *voiceStateUpdateMetadata) error {
	return operations.AddRoleToMember(ctx, metadata.Session, metadata.Guild.ID, metadata.Member.User.ID, metadata.EphemeralRole.ID)
}

func (handler *Handler) removeEphemeralRoles(ctx context.Context, metadata *voiceStateUpdateMetadata) error {
	var err error

	for _, memberRoleID := range metadata.Member.Roles {
		removeError := handler.removeEphemeralRole(ctx, metadata, memberRoleID)
		if removeError != nil {
			if err == nil {
				err = removeError
				continue
			}

			err = fmt.Errorf("%s: %w", err, removeError)
		}
	}

	return err
}

func (handler *Handler) removeEphemeralRole(ctx context.Context, metadata *voiceStateUpdateMetadata, memberRoleID string) error {
	role, err := metadata.Session.State.Role(metadata.Guild.ID, memberRoleID)
	if err != nil {
		if errors.Is(err, discordgo.ErrStateNotFound) {
			return nil
		}

		return fmt.Errorf("unable to remove ephemeral role: %w", err)
	}

	if !strings.HasPrefix(role.Name, handler.RolePrefix) {
		return nil
	}

	err = operations.RemoveRoleFromMember(ctx, metadata.Session, metadata.Guild.ID, metadata.Member.User.ID, role.ID)
	if err != nil {
		if !operations.IsForbiddenResponse(err) {
			return err
		}
	}

	return nil
}
