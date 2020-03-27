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

	voiceStateUpdateError = "Unable to process VoiceStateUpdate"
)

type vsuEvent struct {
	Session      *discordgo.Session
	Guild        *discordgo.Guild
	GuildMember  *discordgo.Member
	GuildRoleMap map[string]*discordgo.Role
}

// VoiceStateUpdate is the callback function for the VoiceStateUpdate event from Discord
func (config *Config) VoiceStateUpdate(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	config.VoiceStateUpdateCounter.Inc()

	ctx, cancelCtx := context.WithTimeout(context.Background(), contextTimeout)
	defer cancelCtx()

	event, err := config.parseEvent(s, vsu)
	if err != nil {
		if errors.Is(err, &userNotFoundError{}) {
			config.Log.WithError(err).Debug(voiceStateUpdateError)
			return
		}

		config.Log.WithError(err).Error(voiceStateUpdateError)

		return
	}

	logWithFields := config.Log.WithFields(logrus.Fields{
		"user":  event.GuildMember.User.Username,
		"guild": event.Guild.Name,
	})

	if event.Guild.Name == discordBotList {
		logWithFields.Debug("Ignoring VoiceStateUpdate event")
		return
	}

	if config.userDisconnectEvent(ctx, vsu, event) {
		logWithFields.Debug("User disconnected from voice channels and ephemeral roles revoked")
		return
	}

	channel, err := s.State.Channel(vsu.ChannelID)
	if err != nil {
		var restErr *discordgo.RESTError

		if errors.As(err, &restErr) {
			if restErr.Response.StatusCode == http.StatusForbidden {
				logWithFields.WithError(err).Debug(voiceStateUpdateError)
				return
			}
		}

		logWithFields.WithError(err).Error(voiceStateUpdateError)

		return
	}

	ephRoleName := config.RolePrefix + " " + channel.Name

	if config.userHasRole(event, ephRoleName) {
		return
	}

	logWithFields = config.Log.WithFields(logrus.Fields{
		"user":  event.GuildMember.User.Username,
		"guild": event.Guild.Name,
		"role":  ephRoleName,
	})

	err = config.grantEphemeralRole(ctx, event, ephRoleName)
	if err != nil {
		var restErr *discordgo.RESTError

		if errors.As(err, &restErr) {
			if restErr.Response.StatusCode == http.StatusForbidden {
				logWithFields.WithError(err).Debug(voiceStateUpdateError)
				return
			}
		}

		err = fmt.Errorf("unable to grant ephemeral role: %w", err)

		logWithFields.WithError(err).Error(voiceStateUpdateError)

		return
	}

	logWithFields.Debugf("Ephemeral role granted")
}

func (config *Config) parseEvent(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) (*vsuEvent, error) {
	guild, err := s.State.Guild(vsu.GuildID)
	if err != nil {
		return nil, fmt.Errorf("unable to determine guild: %w", err)
	}

	guildMember, err := s.GuildMember(vsu.GuildID, vsu.UserID)
	if err != nil {
		return nil, fmt.Errorf("unable to determine guild member: %w", err)
	}

	guildRoles, err := s.GuildRoles(vsu.GuildID)
	if err != nil {
		return nil, fmt.Errorf("unable to determine guild roles: %w", err)
	}

	guildRoleMap := make(map[string]*discordgo.Role)

	for _, role := range guildRoles {
		guildRoleMap[role.ID] = role
	}

	event := &vsuEvent{
		Session:      s,
		Guild:        guild,
		GuildMember:  guildMember,
		GuildRoleMap: guildRoleMap,
	}

	return event, nil
}

func (config *Config) userDisconnectEvent(ctx context.Context, vsu *discordgo.VoiceStateUpdate, event *vsuEvent) bool {
	if vsu.ChannelID != "" {
		return false
	}

	config.revokeEphemeralRoles(ctx, event)

	return true
}

func (config *Config) userHasRole(event *vsuEvent, ephRoleName string) bool {
	for _, memberRoleID := range event.GuildMember.Roles {
		if event.GuildRoleMap[memberRoleID].Name == ephRoleName {
			return true
		}
	}

	return false
}

func (config *Config) grantEphemeralRole(ctx context.Context, event *vsuEvent, ephRoleName string) error {
	config.revokeEphemeralRoles(ctx, event)

	ephRole, err := config.getGuildRole(ctx, event, ephRoleName)
	if err != nil {
		return err
	}

	err = event.Session.GuildMemberRoleAddWithContext(ctx, event.Guild.ID, event.GuildMember.User.ID, ephRole.ID)
	if err != nil {
		return err
	}

	return nil
}

func (config *Config) getGuildRole(ctx context.Context, event *vsuEvent, ephRoleName string) (*discordgo.Role, error) {
	for _, guildRole := range event.GuildRoleMap {
		if guildRole.Name == ephRoleName {
			return guildRole, nil
		}
	}

	ephRole, err := config.guildRoleCreate(ctx, event, ephRoleName)
	if err != nil {
		return nil, err
	}

	return ephRole, nil
}

func (config *Config) guildRoleCreate(ctx context.Context, event *vsuEvent, ephRoleName string) (*discordgo.Role, error) {
	ephRole, err := event.Session.GuildRoleCreateWithContext(ctx, event.Guild.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to create ephemeral role: %w", err)
	}

	ephRole, err = event.Session.GuildRoleEditWithContext(
		ctx,
		event.Guild.ID,
		ephRole.ID,
		ephRoleName,
		config.RoleColor,
		true,
		ephRole.Permissions,
		ephRole.Mentionable,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to edit ephemeral role: %w", err)
	}

	return ephRole, nil
}

func (config *Config) revokeEphemeralRoles(ctx context.Context, event *vsuEvent) {
	for _, roleID := range event.GuildMember.Roles {
		role := event.GuildRoleMap[roleID]

		if strings.HasPrefix(role.Name, config.RolePrefix) {
			err := event.Session.GuildMemberRoleRemoveWithContext(ctx, event.Guild.ID, event.GuildMember.User.ID, role.ID)
			if err != nil {
				config.Log.WithError(err).
					WithFields(logrus.Fields{
						"user":  event.GuildMember.User.Username,
						"role":  role.Name,
						"guild": event.Guild.Name,
					}).Debugf("Unable to remove role on VoiceStateUpdate")

				return
			}

			config.Log.WithFields(logrus.Fields{
				"user":  event.GuildMember.User.Username,
				"role":  role.Name,
				"guild": event.Guild.Name,
			}).Debugf("Revoked role in VoiceStateUpdate")
		}
	}
}
