package callbacks

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

const (
	discordBotList = "Discord Bot List"

	voiceStateUpdateError     = "Unable to process VoiceStateUpdate"
	revokeEphemeralRolesError = "unable to revoke ephemeral roles"
)

type vsuEvent struct {
	Session       *discordgo.Session
	Guild         *discordgo.Guild
	GuildMember   *discordgo.Member
	GuildRoleMap  map[string]*discordgo.Role
	Channel       *discordgo.Channel
	GuildRole     *discordgo.Role
	GuildRoleName string
}

// VoiceStateUpdate is the callback function for the VoiceStateUpdate event from Discord.
func (config *Config) VoiceStateUpdate(session *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	config.VoiceStateUpdateCounter.Inc()

	ctx, cancelCtx := context.WithTimeout(context.Background(), contextTimeout)
	defer cancelCtx()

	config.Log.Debug("Parsing VoiceStateUpdate event")

	event, err := config.parseEvent(ctx, session, vsu)
	if err != nil {
		for _, customError := range []error{&memberNotFound{}, &userNotFound{}} {
			if errors.Is(err, customError) {
				config.Log.WithError(customError).Debug(voiceStateUpdateError)
				return
			}
		}

		config.Log.WithError(err).Error(voiceStateUpdateError)

		return
	}

	log := config.Log.WithFields(
		logrus.Fields{
			"member": event.GuildMember.User.Username,
			"guild":  event.Guild.Name,
		},
	)

	if event.Guild.Name == discordBotList {
		log.Debug("Ignoring VoiceStateUpdate event")
		return
	}

	log.Debug("Revoking Ephemeral Roles")

	err = config.revokeEphemeralRoles(ctx, event)
	if err != nil {
		log.WithError(err).Error(revokeEphemeralRolesError)
	}

	if event.Channel == nil {
		return
	}

	log.WithField("role", event.GuildRoleName).Debug("Granting Ephemeral Role")

	err = config.grantEphemeralRole(ctx, event)
	if err != nil {
		var restErr *discordgo.RESTError

		if errors.As(err, &restErr) {
			if restErr.Response.StatusCode == http.StatusForbidden {
				log.WithError(err).Debug(voiceStateUpdateError)
				return
			}
		}

		log.WithError(err).Error(voiceStateUpdateError)

		return
	}
}

func (config *Config) parseEvent(ctx context.Context, session *discordgo.Session, vsu *discordgo.VoiceStateUpdate) (*vsuEvent, error) {
	guild, err := lookupGuild(ctx, session, vsu.GuildID)
	if err != nil {
		return nil, fmt.Errorf("unable to determine guild: %w", err)
	}

	guildMember, err := lookupGuildMember(ctx, session, vsu.GuildID, vsu.UserID)
	if err != nil {
		return nil, err
	}

	guildRoles, err := lookupGuildRoles(ctx, session, vsu.GuildID)
	if err != nil {
		return nil, fmt.Errorf("unable to determine guild roles: %w", err)
	}

	channel, err := lookupGuildChannel(ctx, session, vsu.GuildID, vsu.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("unable to determine guild channel: %w", err)
	}

	guildRoleMap := mapGuildRoleIDs(guildRoles)

	if channel == nil {
		return &vsuEvent{
			Session:      session,
			Guild:        guild,
			GuildMember:  guildMember,
			GuildRoleMap: guildRoleMap,
		}, nil
	}

	guildRoleName := config.RolePrefix + " " + channel.Name

	var guildRole *discordgo.Role

	for _, role := range guildRoleMap {
		if role.Name == guildRoleName {
			guildRole = role
		}
	}

	return &vsuEvent{
		Session:       session,
		Guild:         guild,
		GuildMember:   guildMember,
		GuildRoleMap:  guildRoleMap,
		Channel:       channel,
		GuildRole:     guildRole,
		GuildRoleName: guildRoleName,
	}, nil
}

func (config *Config) grantEphemeralRole(ctx context.Context, event *vsuEvent) error {
	if event.GuildRole == nil {
		newRole, err := createGuildRole(ctx, event.Session, event.Guild.ID, event.GuildRoleName, config.RoleColor)
		if err != nil {
			return err
		}

		event.GuildRole = newRole
	}

	return addRoleToMember(ctx, event.Session, event.Guild.ID, event.GuildMember.User.ID, event.GuildRole.ID)
}

func (config *Config) revokeEphemeralRoles(ctx context.Context, event *vsuEvent) error {
	var revokeErrors []error

	for _, roleID := range event.GuildMember.Roles {
		role := event.GuildRoleMap[roleID]

		if strings.HasPrefix(role.Name, config.RolePrefix) {
			err := removeRoleFromMember(ctx, event.Session, event.Guild.ID, event.GuildMember.User.ID, role.ID)
			if err != nil {
				revokeErrors = append(revokeErrors, fmt.Errorf("unable to revoke %s: %w", role.Name, err))
			}
		}
	}

	if revokeErrors != nil {
		var err error

		for _, revokeError := range revokeErrors {
			err = fmt.Errorf("%s, %w", revokeError, err)
		}

		return err
	}

	return nil
}
