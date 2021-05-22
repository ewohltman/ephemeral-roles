package callbacks

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
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

	metadata, err := handler.parseEvent(session, voiceState)
	if err != nil {
		handler.handleParseEventError(session, err)
		return
	}

	log := handler.Log.WithFields(
		logrus.Fields{
			"guild":  metadata.Guild.Name,
			"member": metadata.Member.User.Username,
		},
	)

	if metadata.EphemeralRole != nil {
		if handler.memberHasRole(metadata.Member, metadata.EphemeralRole) {
			return
		}
	}

	err = handler.removeEphemeralRoles(metadata)
	if err != nil {
		log.WithError(err).Error(voiceStateUpdateEventError)
	}

	if metadata.Channel == nil {
		return
	}

	err = handler.addEphemeralRole(metadata)
	if err != nil {
		if operations.ShouldLogDebug(err) {
			log.WithError(err).Debug(voiceStateUpdateEventError)
			return
		}

		log.WithError(err).Error(voiceStateUpdateEventError)

		return
	}
}

func (handler *Handler) parseEvent(
	session *discordgo.Session,
	voiceState *discordgo.VoiceStateUpdate,
) (*voiceStateUpdateMetadata, error) {
	guild, err := operations.LookupGuild(session, voiceState.GuildID)
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

	if voiceState.ChannelID == "" {
		return &voiceStateUpdateMetadata{
			Session: session,
			Guild:   guild,
			Member:  member,
		}, nil
	}

	channel, err := session.State.Channel(voiceState.ChannelID)
	if err != nil {
		return nil, &ChannelNotFound{
			Guild:  guild,
			Member: member,
			Err:    err,
		}
	}

	err = operations.BotHasChannelPermission(session, channel)
	if err != nil {
		return nil, &InsufficientPermissions{
			Guild:   guild,
			Member:  member,
			Channel: channel,
			Err:     err,
		}
	}

	ephemeralRole, err := handler.lookupGuildRole(guild, channel)
	if errors.Is(err, &RoleNotFound{}) {
		ephemeralRole, err = handler.createRole(guild, handler.RoleNameFromChannel(channel.Name))
		if err != nil {
			switch {
			case operations.IsDeadlineExceeded(err):
				return nil, &DeadlineExceeded{Guild: guild, Member: member, Channel: channel, Err: err}
			case operations.IsForbiddenResponse(err):
				return nil, &InsufficientPermissions{Guild: guild, Member: member, Channel: channel, Err: err}
			case operations.IsMaxGuildsResponse(err):
				return nil, &MaxNumberOfRoles{Guild: guild, Member: member, Channel: channel, Err: err}
			default:
				return nil, err
			}
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

func (handler *Handler) handleParseEventError(session *discordgo.Session, err error) {
	var (
		memberNotFoundErr          *MemberNotFound
		channelNotFoundErr         *ChannelNotFound
		insufficientPermissionsErr *InsufficientPermissions
		maxNumberOfRolesErr        *MaxNumberOfRoles
		deadlineExceededErr        *DeadlineExceeded
	)

	switch {
	case errors.As(err, &memberNotFoundErr):
		handler.logCleanup(session, memberNotFoundErr)
	case errors.As(err, &channelNotFoundErr):
		handler.logCleanup(session, channelNotFoundErr)
	case errors.As(err, &insufficientPermissionsErr):
		handler.logCleanup(session, insufficientPermissionsErr)
	case errors.As(err, &maxNumberOfRolesErr):
		handler.logCleanup(session, maxNumberOfRolesErr)
	case errors.As(err, &deadlineExceededErr):
		handler.logParseEventError(deadlineExceededErr)
	default:
		handler.Log.WithError(err).Error(voiceStateUpdateEventError)
	}
}

func (handler *Handler) logCleanup(session *discordgo.Session, callbackError CallbackError) {
	handler.logParseEventError(callbackError)
	handler.cleanupParseEventError(session, callbackError)
}

func (handler *Handler) logParseEventError(callbackError CallbackError) {
	handler.newCallbackErrorLogger(callbackError).WithError(callbackError).Debug(voiceStateUpdateEventError)
}

func (handler *Handler) cleanupParseEventError(session *discordgo.Session, callbackError CallbackError) {
	metadata := &voiceStateUpdateMetadata{
		Session: session,
		Guild:   callbackError.InGuild(),
		Member:  callbackError.ForMember(),
	}

	if metadata.Member == nil {
		return
	}

	err := handler.removeEphemeralRoles(metadata)
	if err != nil {
		handler.newCallbackErrorLogger(callbackError).WithError(err).Debug(voiceStateUpdateEventError)
	}
}

func (handler *Handler) newCallbackErrorLogger(callbackError CallbackError) *logrus.Entry {
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

	return log
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

	return index != len(memberRoles) && memberRoles[index] == role.ID
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

	if index != len(guildRoles) && guildRoles[index].Name == ephemeralRoleName {
		return guildRoles[index], nil
	}

	return nil, &RoleNotFound{}
}

func (handler *Handler) createRole(guild *discordgo.Guild, roleName string) (*discordgo.Role, error) {
	resultChannel := operations.NewResultChannel()

	handler.OperationsGateway.Process(resultChannel, &operations.Request{
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

func (handler *Handler) addEphemeralRole(metadata *voiceStateUpdateMetadata) error {
	return operations.AddRoleToMember(metadata.Session, metadata.Guild.ID, metadata.Member.User.ID, metadata.EphemeralRole.ID)
}

func (handler *Handler) removeEphemeralRoles(metadata *voiceStateUpdateMetadata) error {
	var err error

	for _, roleID := range metadata.Member.Roles {
		removeError := handler.removeEphemeralRole(metadata, roleID)
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

func (handler *Handler) removeEphemeralRole(metadata *voiceStateUpdateMetadata, roleID string) error {
	role, err := metadata.Session.State.Role(metadata.Guild.ID, roleID)
	if err != nil {
		if errors.Is(err, discordgo.ErrStateNotFound) {
			return nil
		}

		return fmt.Errorf("unable to remove ephemeral role: %w", err)
	}

	if !strings.HasPrefix(role.Name, handler.RolePrefix) {
		return nil
	}

	err = operations.RemoveRoleFromMember(metadata.Session, metadata.Guild.ID, metadata.Member.User.ID, role.ID)
	if err != nil {
		if !operations.IsForbiddenResponse(err) {
			return err
		}
	}

	return nil
}
