package callbacks

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

const (
	defaultRoleColor      = 16753920 // Default to orange hex #FFA500 in decimal
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
	// Increment the total number of VoiceStateUpdate events
	config.VoiceStateUpdateCounter.Inc()

	event, err := config.parseEvent(s, vsu)
	if err != nil {
		config.Log.WithError(err).Error(voiceStateUpdateError)
		return
	}

	logWithFields := config.Log.WithFields(logrus.Fields{
		"user":  event.GuildMember.User.Username,
		"guild": event.Guild.Name,
	})

	if config.userDisconnectEvent(vsu, event) {
		logWithFields.Debugf("User disconnected from voice channels and ephemeral roles revoked")
		return
	}

	// Get the channel
	channel, err := s.Channel(vsu.ChannelID)
	if err != nil {
		err = fmt.Errorf("unable to determine channel: %w", err)

		logWithFields.WithError(err).Error(voiceStateUpdateError)

		return
	}

	ephRoleName := config.RolePrefix + " " + channel.Name

	if config.userHasRole(event, ephRoleName) {
		return
	}

	logWithFields = config.Log.WithFields(logrus.Fields{
		"user":    event.GuildMember.User.Username,
		"guild":   event.Guild.Name,
		"guildID": event.Guild.ID,
		"role":    ephRoleName,
	})

	// Add role to member
	err = config.grantEphemeralRole(event, ephRoleName)
	if err != nil {
		err = fmt.Errorf("unable to grant ephemeral role: %w", err)

		logWithFields.WithError(err).Error(voiceStateUpdateError)

		return
	}

	logWithFields.Debugf("Ephemeral role granted")
}

func (config *Config) parseEvent(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) (*vsuEvent, error) {
	user, guild, err := config.getUserGuild(s, vsu)
	if err != nil {
		return nil, err
	}

	var guildMember *discordgo.Member

	for _, member := range guild.Members {
		if member.User.ID == user.ID {
			guildMember = member
			break
		}
	}

	if guildMember == nil {
		return nil, errors.New("user not found in guild members")
	}

	guildRoleMap := make(map[string]*discordgo.Role)

	for _, role := range guild.Roles {
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

func (config *Config) getUserGuild(
	s *discordgo.Session,
	vsu *discordgo.VoiceStateUpdate,
) (*discordgo.User, *discordgo.Guild, error) {
	// Get the user
	user, err := s.User(vsu.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to determine user: %w", err)
	}

	// Get the guild
	guild, err := s.Guild(vsu.GuildID)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to determine guild: %w", err)
	}

	return user, guild, nil
}

func (config *Config) userDisconnectEvent(vsu *discordgo.VoiceStateUpdate, event *vsuEvent) bool {
	if vsu.ChannelID != "" {
		return false
	}

	config.revokeEphemeralRoles(event)

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

func (config *Config) grantEphemeralRole(event *vsuEvent, ephRoleName string) error {
	// Revoke any previous ephemeral roles
	config.revokeEphemeralRoles(event)

	ephRole, err := config.getGuildRole(event, ephRoleName)
	if err != nil {
		return err
	}

	// Add our member to role
	err = event.Session.GuildMemberRoleAdd(event.Guild.ID, event.GuildMember.User.ID, ephRole.ID)
	if err != nil {
		return err
	}

	return nil
}

func (config *Config) getGuildRole(event *vsuEvent, ephRoleName string) (*discordgo.Role, error) {
	// Check to see if the role already exists in the guild
	for _, guildRole := range event.GuildRoleMap {
		if guildRole.Name == ephRoleName {
			return guildRole, nil
		}
	}

	ephRole, err := config.guildRoleCreate(event, ephRoleName)
	if err != nil {
		return nil, err
	}

	return ephRole, nil
}

func (config *Config) guildRoleCreate(event *vsuEvent, ephRoleName string) (*discordgo.Role, error) {
	// Create a new blank role
	ephRole, err := event.Session.GuildRoleCreate(event.Guild.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to create ephemeral role: %w", err)
	}

	roleColor := defaultRoleColor

	// Check for role color override
	if colorString, found := os.LookupEnv("ROLE_COLOR_HEX2DEC"); found {
		roleColor, err = strconv.Atoi(colorString)
		if err != nil {
			config.Log.
				WithError(err).
				WithField("ROLE_COLOR_HEX2DEC", colorString).
				Warnf("Error parsing ROLE_COLOR_HEX2DEC from environment")

			roleColor = defaultRoleColor
		}
	}

	// Edit the new role
	ephRole, err = event.Session.GuildRoleEdit(
		event.Guild.ID,
		ephRole.ID,
		ephRoleName,
		roleColor,
		true,
		ephRole.Permissions,
		ephRole.Mentionable,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to edit ephemeral role: %w", err)
	}

	return ephRole, nil
}

func (config *Config) revokeEphemeralRoles(event *vsuEvent) {
	for _, roleID := range event.GuildMember.Roles {
		role := event.GuildRoleMap[roleID]

		if strings.HasPrefix(role.Name, config.RolePrefix) {
			// Found ephemeral role, revoke it
			err := event.Session.GuildMemberRoleRemove(event.Guild.ID, event.GuildMember.User.ID, role.ID)
			if err != nil {
				config.Log.WithError(err).
					WithFields(logrus.Fields{
						"user":  event.GuildMember.User.Username,
						"role":  role.Name,
						"guild": event.Guild.Name,
					}).Debugf("Unable to remove role on VoiceStateUpdate")

				return
			}

			config.Log.WithError(err).WithFields(logrus.Fields{
				"user":  event.GuildMember.User.Username,
				"role":  role.Name,
				"guild": event.Guild.Name,
			}).Debugf("Revoked role in VoiceStateUpdate")
		}
	}
}
