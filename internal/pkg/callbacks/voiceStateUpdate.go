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
)

const (
	voiceStateUpdate = "VoiceStateUpdate"

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
func (config *Config) VoiceStateUpdate(session *discordgo.Session, voiceState *discordgo.VoiceStateUpdate) {
	config.VoiceStateUpdateCounter.Inc()

	span := config.JaegerTracer.StartSpan(voiceStateUpdate)
	defer span.Finish()

	ctx, cancelCtx := context.WithTimeout(context.Background(), config.ContextTimeout)
	defer cancelCtx()

	ctx = opentracing.ContextWithSpan(ctx, span)

	metadata, err := config.parseEvent(ctx, session, voiceState)
	if err != nil {
		config.handleParseEventError(ctx, session, err)
		return
	}

	log := config.Log.WithFields(
		logrus.Fields{
			"guild":  metadata.Guild.Name,
			"member": metadata.Member.User.Username,
		},
	)

	err = config.revokeEphemeralRoles(ctx, metadata)
	if err != nil {
		log.WithError(err).Error(voiceStateUpdateEventError)
	}

	log.Debug("Revoked Ephemeral Roles")

	err = config.grantEphemeralRole(ctx, metadata)
	if err != nil {
		if forbiddenResponse(err) {
			log.WithError(err).Debug(voiceStateUpdateEventError)
			return
		}

		log.WithError(err).Error(voiceStateUpdateEventError)

		return
	}

	log.WithField("role", metadata.EphemeralRole.Name).Debug("Granted Ephemeral Role")
}

func (config *Config) parseEvent(
	ctx context.Context,
	session *discordgo.Session,
	voiceState *discordgo.VoiceStateUpdate,
) (*voiceStateUpdateMetadata, error) {
	guild, err := lookupGuild(ctx, session, voiceState.GuildID)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup guild: %w", err)
	}

	member, err := session.State.Member(voiceState.GuildID, voiceState.UserID)
	if err != nil {
		return nil, &memberNotFound{
			guild: guild,
			err:   err,
		}
	}

	channel, err := session.State.GuildChannel(voiceState.GuildID, voiceState.ChannelID)
	if err != nil {
		return nil, &channelNotFound{
			guild:  guild,
			member: member,
			err:    err,
		}
	}

	err = config.botHasChannelPermission(ctx, session, channel)
	if err != nil {
		return nil, &insufficientPermissions{
			guild:   guild,
			member:  member,
			channel: channel,
			err:     err,
		}
	}

	ephemeralRole, err := config.lookupRole(ctx, session, guild, channel)
	if err != nil {
		if !forbiddenResponse(err) {
			return nil, err
		}

		return nil, &insufficientPermissions{
			guild:   guild,
			member:  member,
			channel: channel,
			err:     err,
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

func (config *Config) handleParseEventError(ctx context.Context, session *discordgo.Session, err error) {
	var (
		memberNotFoundErr          *memberNotFound
		channelNotFoundErr         *channelNotFound
		insufficientPermissionsErr *insufficientPermissions
	)

	switch {
	case errors.As(err, &memberNotFoundErr):
		config.logRevoke(ctx, session, memberNotFoundErr)
	case errors.As(err, &channelNotFoundErr):
		config.logRevoke(ctx, session, channelNotFoundErr)
	case errors.As(err, &insufficientPermissionsErr):
		config.logRevoke(ctx, session, insufficientPermissionsErr)
	default:
		config.Log.WithError(err).Error(voiceStateUpdateEventError)
	}
}

func (config *Config) logRevoke(ctx context.Context, session *discordgo.Session, customErr customError) {
	log := logrus.NewEntry(config.Log.WrappedLogger())

	guild := customErr.Guild()
	if guild != nil {
		log = log.WithField("guild", guild.Name)
	}

	member := customErr.Member()
	if member != nil {
		log = log.WithField("member", member.User.Username)
	}

	channel := customErr.Channel()
	if channel != nil {
		log = log.WithField("channel", channel.Name)
	}

	log.WithError(customErr).Debug(voiceStateUpdateEventError)

	if member == nil {
		return
	}

	metadata := &voiceStateUpdateMetadata{
		Session: session,
		Guild:   customErr.Guild(),
		Member:  customErr.Member(),
	}

	err := config.revokeEphemeralRoles(ctx, metadata)
	if err != nil {
		log.WithError(err).Debug(voiceStateUpdateEventError)
	}
}

func (config *Config) botHasChannelPermission(
	ctx context.Context,
	session *discordgo.Session,
	channel *discordgo.Channel,
) error {
	bot, err := session.UserWithContext(ctx, "@me")
	if err != nil {
		return fmt.Errorf("unable to determine bot user: %w", err)
	}

	permissions, err := session.UserChannelPermissions(bot.ID, channel.ID)
	if err != nil {
		return fmt.Errorf("unable to determine channel permissions: %w", err)
	}

	if permissions&discordgo.PermissionViewChannel != discordgo.PermissionViewChannel {
		return fmt.Errorf("insufficient channel permissions: channel: %s", channel.Name)
	}

	return nil
}

// lookupRole sorts roles in O(n*log(n)) time and does a binary search for the
// associated ephemeral role.
func (config *Config) lookupRole(
	ctx context.Context,
	session *discordgo.Session,
	guild *discordgo.Guild,
	channel *discordgo.Channel,
) (*discordgo.Role, error) {
	guildRoles := make(discordgo.Roles, len(guild.Roles))

	copy(guildRoles, guild.Roles)

	sort.Slice(
		guildRoles,
		func(i, j int) bool {
			return guildRoles[i].Name < guildRoles[j].Name
		},
	)

	ephemeralRoleName := config.RolePrefix + " " + channel.Name

	index := sort.Search(
		len(guildRoles),
		func(i int) bool {
			return guildRoles[i].Name >= ephemeralRoleName
		},
	)

	if index < len(guildRoles) && guildRoles[index].Name == ephemeralRoleName {
		return guildRoles[index], nil
	}

	return createGuildRole(ctx, session, guild, ephemeralRoleName, config.RoleColor)
}

func (config *Config) grantEphemeralRole(ctx context.Context, metadata *voiceStateUpdateMetadata) error {
	return addRoleToMember(ctx, metadata.Session, metadata.Guild.ID, metadata.Member.User.ID, metadata.EphemeralRole.ID)
}

func (config *Config) revokeEphemeralRoles(ctx context.Context, metadata *voiceStateUpdateMetadata) error {
	var err error

	for _, memberRoleID := range metadata.Member.Roles {
		revokeError := config.revokeEphemeralRole(ctx, metadata, memberRoleID)
		if revokeError != nil {
			if err == nil {
				err = revokeError
				continue
			}

			err = fmt.Errorf("%s: %w", err, revokeError)
		}
	}

	return err
}

func (config *Config) revokeEphemeralRole(ctx context.Context, metadata *voiceStateUpdateMetadata, memberRoleID string) error {
	role, err := metadata.Session.State.Role(metadata.Guild.ID, memberRoleID)
	if err != nil {
		if errors.Is(err, discordgo.ErrStateNotFound) {
			return nil
		}

		return fmt.Errorf("unable to revoke ephemeral role: %w", err)
	}

	if !strings.HasPrefix(role.Name, config.RolePrefix) {
		return nil
	}

	err = removeRoleFromMember(ctx, metadata.Session, metadata.Guild.ID, metadata.Member.User.ID, role.ID)
	if err != nil {
		if !forbiddenResponse(err) {
			return err
		}
	}

	return nil
}
