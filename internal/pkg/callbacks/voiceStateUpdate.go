package callbacks

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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
	Session           *discordgo.Session
	Guild             *discordgo.Guild
	GuildMember       *discordgo.Member
	GuildRoleMap      roleIDMap
	Channel           *discordgo.Channel
	EphemeralRole     *discordgo.Role
	EphemeralRoleName string
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
		var memberNotFoundErr *memberNotFound

		if errors.As(err, &memberNotFoundErr) {
			config.Log.WithError(memberNotFoundErr).Debug(voiceStateUpdateEventError)
			return
		}

		config.Log.WithError(err).Error(voiceStateUpdateEventError)

		return
	}

	log := config.Log.WithFields(
		logrus.Fields{
			"member": metadata.GuildMember.User.Username,
			"guild":  metadata.Guild.Name,
		},
	)

	log.Debug("Revoking Ephemeral Roles")

	err = config.revokeEphemeralRoles(ctx, metadata)
	if err != nil {
		if forbiddenResponse(err) {
			log.WithError(err).Debug(voiceStateUpdateEventError)
		} else {
			log.WithError(err).Error(voiceStateUpdateEventError)
		}
	}

	if metadata.Channel == nil {
		return
	}

	log.WithField("role", metadata.EphemeralRoleName).Debug("Granting Ephemeral Role")

	err = config.grantEphemeralRole(ctx, metadata)
	if err != nil {
		if forbiddenResponse(err) {
			log.WithError(err).Debug(voiceStateUpdateEventError)
		} else {
			log.WithError(err).Error(voiceStateUpdateEventError)
		}
	}
}

func (config *Config) parseEvent(
	ctx context.Context,
	session *discordgo.Session,
	voiceState *discordgo.VoiceStateUpdate,
) (*voiceStateUpdateMetadata, error) {
	guild, err := lookupGuild(ctx, session, voiceState.GuildID)
	if err != nil {
		return nil, fmt.Errorf("unable to parse event: %w", err)
	}

	guildRoleMap := mapGuildRoleIDs(guild.Roles)

	var guildMember *discordgo.Member

	for _, member := range guild.Members {
		if member.User.ID == voiceState.UserID {
			guildMember = member
			break
		}
	}

	if guildMember == nil {
		return nil, &memberNotFound{}
	}

	var guildChannel *discordgo.Channel

	for _, channel := range guild.Channels {
		if channel.ID == voiceState.ChannelID {
			guildChannel = channel
			break
		}
	}

	if guildChannel == nil {
		return &voiceStateUpdateMetadata{
			Session:      session,
			Guild:        guild,
			GuildMember:  guildMember,
			GuildRoleMap: guildRoleMap,
		}, nil
	}

	err = config.botHasChannelPermission(ctx, session, guildChannel)
	if err != nil {
		if errors.Is(err, &insufficientPermission{}) {
			config.Log.WithError(err).WithFields(
				logrus.Fields{
					"guild":   guild.Name,
					"channel": guildChannel.Name,
				},
			).Debugf("")

			return &voiceStateUpdateMetadata{
				Session:      session,
				Guild:        guild,
				GuildMember:  guildMember,
				GuildRoleMap: guildRoleMap,
			}, nil
		}

		return nil, err
	}

	ephemeralRole, ephemeralRoleName := config.lookupRole(guildChannel, guildRoleMap)

	return &voiceStateUpdateMetadata{
		Session:           session,
		Guild:             guild,
		GuildMember:       guildMember,
		GuildRoleMap:      guildRoleMap,
		Channel:           guildChannel,
		EphemeralRole:     ephemeralRole,
		EphemeralRoleName: ephemeralRoleName,
	}, nil
}

func (config *Config) botHasChannelPermission(ctx context.Context, session *discordgo.Session, channel *discordgo.Channel) error {
	bot, err := session.UserWithContext(ctx, "@me")
	if err != nil {
		return fmt.Errorf("unable to determine bot user: %w", err)
	}

	permissions, err := session.UserChannelPermissions(bot.ID, channel.ID)
	if err != nil {
		return fmt.Errorf("unable to determine channel permissions: %w", err)
	}

	if permissions&discordgo.PermissionViewChannel != discordgo.PermissionViewChannel {
		return &insufficientPermission{}
	}

	return nil
}

func (config *Config) lookupRole(channel *discordgo.Channel, roleMap roleIDMap) (ephemeralRole *discordgo.Role, ephemeralRoleName string) {
	ephemeralRoleName = config.RolePrefix + " " + channel.Name

	for _, role := range roleMap {
		if role.Name == ephemeralRoleName {
			ephemeralRole = role
			break
		}
	}

	return
}

func (config *Config) revokeEphemeralRoles(ctx context.Context, metadata *voiceStateUpdateMetadata) error {
	var revokeErrors []error

	for _, memberRoleID := range metadata.GuildMember.Roles {
		role := metadata.GuildRoleMap[roleID(memberRoleID)]

		if strings.HasPrefix(role.Name, config.RolePrefix) {
			err := removeRoleFromMember(ctx, metadata.Session, metadata.Guild.ID, metadata.GuildMember.User.ID, role.ID)
			if err != nil {
				revokeErrors = append(revokeErrors, fmt.Errorf("unable to revoke %s: %w", role.Name, err))
			}
		}
	}

	if revokeErrors != nil {
		var err error

		for _, revokeError := range revokeErrors {
			if err != nil {
				err = fmt.Errorf("%s, %w", err, revokeError)
				continue
			}

			err = revokeError
		}

		return err
	}

	return nil
}

func (config *Config) grantEphemeralRole(ctx context.Context, metadata *voiceStateUpdateMetadata) error {
	if metadata.EphemeralRole == nil {
		newRole, err := createGuildRole(ctx, metadata.Session, metadata.Guild.ID, metadata.EphemeralRoleName, config.RoleColor)
		if err != nil {
			return err
		}

		metadata.EphemeralRole = newRole
	}

	return addRoleToMember(ctx, metadata.Session, metadata.Guild.ID, metadata.GuildMember.User.ID, metadata.EphemeralRole.ID)
}

func forbiddenResponse(err error) bool {
	var restErr *discordgo.RESTError

	if errors.As(err, &restErr) {
		if restErr.Response.StatusCode == http.StatusForbidden {
			return true
		}
	}

	return false
}
