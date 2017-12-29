package callbacks

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/ewohltman/discordgo"
	"github.com/sirupsen/logrus"
)

// VoiceStateUpdate is the callback function for the VoiceStateUpdate event from Discord
func VoiceStateUpdate(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	// Get the user
	user, err := s.User(vsu.UserID)
	if err != nil {
		log.WithError(err).Debugf("Unable to determine user in VoiceStateUpdate")

		return
	}

	// Get the guild
	guild, err := s.Guild(vsu.GuildID)
	if err != nil {
		log.WithError(err).Debugf("Unable to determine guild in VoiceStateUpdate")

		return
	}

	// Revoke all ephemeral roles from this user and start clean
	guildCleanupMemberEphemeralRoles(s, user, guild)

	if vsu.ChannelID == "" { // User disconnect
		log.WithFields(logrus.Fields{
			"user":  user.Username,
			"guild": guild.Name,
		}).Debugf("User disconnected from voice channels and ephemeral roles revoked")

		return
	}

	// Get the channel
	channel, err := s.Channel(vsu.ChannelID)
	if err != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"user":  user.Username,
				"guild": guild.Name,
			}).
			Debugf("Unable to determine channel in VoiceStateUpdate")

		return
	}

	// Get the roles
	guildRoles, dErr := getGuildRoles(s, vsu.GuildID)
	if dErr != nil {
		log.WithError(dErr).
			Debugf("Unable to determine guild roles in VoiceStateUpdate")

		return
	}

	ephRoleName := ROLEPREFIX + channel.Name
	var ephRole *discordgo.Role

	// Check to see if the role already exists
	for _, role := range guildRoles {
		if role.Name == ephRoleName { // Found role
			ephRole = role

			// Add member to ephemeral role
			err = s.GuildMemberRoleAdd(guild.ID, user.ID, ephRole.ID)
			if err != nil {
				log.WithError(err).WithFields(logrus.Fields{
					"user":  user.Username,
					"role":  ephRole.Name,
					"guild": guild.Name,
				}).Debugf("Unable to add user to ephemeral role")

				return
			}

			log.WithFields(logrus.Fields{
				"user":  user.Username,
				"role":  ephRole.Name,
				"guild": guild.Name,
			}).Debugf("Added role")

			return
		}
	}

	// Role does not exist
	if ephRole == nil {
		var err error

		// Create and edit a new role
		ephRole, err = guildRoleCreateEdit(s, ephRoleName, guild)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"role":  ephRoleName,
				"guild": guild.Name,
			}).Debugf("Unable to manage ephemeral role")

			return
		}

		// Add our member to role
		err = s.GuildMemberRoleAdd(guild.ID, user.ID, ephRole.ID)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"user":  user.Username,
				"role":  ephRole.Name,
				"guild": guild.Name,
			}).Debugf("Unable to add user to ephemeral role")

			return
		}

		log.WithFields(logrus.Fields{
			"user":  user.Username,
			"role":  ephRole.Name,
			"guild": guild.Name,
		}).Debugf("Added role")
	}

	return
}

// getGuildRoles handles role lookups using dErr *discordError as a means to
// provide context to API errors
func getGuildRoles(
	s *discordgo.Session,
	guildID string,
) (roles []*discordgo.Role, dErr *discordError) {

	var err error

	roles, err = s.GuildRoles(guildID)
	if err != nil {
		// Find the JSON with regular expressions
		rx := regexp.MustCompile("{.*}")
		errHTTPString := rx.ReplaceAllString(err.Error(), "")
		errJSONString := rx.FindString(err.Error())

		dAPIResp := &DiscordAPIResponse{}

		dErr = &discordError{
			HTTPResponseMessage: errHTTPString,
			APIResponse:         dAPIResp,
			CustomMessage:       "",
		}

		unmarshalErr := json.Unmarshal([]byte(errJSONString), dAPIResp)
		if unmarshalErr != nil {
			dAPIResp.Code = -1
			dAPIResp.Message = "Unable to unmarshal Discord API JSON response: " + errJSONString

			return
		}

		// Add CustomMessage as appropriate
		switch dErr.APIResponse.Code {
		case 50013: // Code 50013: "Missing Permissions"
			dErr.CustomMessage = "Insufficient role permission to query guild roles"
		}
	}

	return
}

func guildRoleCreateEdit(
	s *discordgo.Session,
	ephRoleName string,
	guild *discordgo.Guild,
) (ephRole *discordgo.Role, err error) {

	// Create a new blank role
	ephRole, err = s.GuildRoleCreate(guild.ID)
	if err != nil {
		err = fmt.Errorf("unable to create ephemeral role: %s", err.Error())

		return
	}

	// Check for role color override
	roleColor := 16753920 // Default to orange hex #FFA500 in decimal
	if colorString, found := os.LookupEnv("ROLE_COLOR_HEX2DEC"); found {
		parsedString, err := strconv.Atoi(colorString)
		if err != nil {
			log.WithError(err).
				WithField("ROLE_COLOR_HEX2DEC", colorString).
				Warnf("Error parsing ROLE_COLOR_HEX2DEC from environment")
		} else {
			roleColor = parsedString
		}
	}

	// Edit the new role
	ephRole, err = s.GuildRoleEdit(
		guild.ID,
		ephRole.ID,
		ephRoleName,
		roleColor, // Orange hex #FFA500 to decimal
		true,
		ephRole.Permissions,
		ephRole.Mentionable,
	)
	if err != nil {
		err = fmt.Errorf("unable to edit ephemeral role: %s", err.Error())

		return
	}

	err = guildRoleReorder(s, guild.ID)

	return
}

// guildRoleReorder orders roles in the order in which the channels appear
func guildRoleReorder(s *discordgo.Session, guildID string) error {
	// Get channels via API using s.GuildChannels() for most up-to-date data
	guildChannels, err := s.GuildChannels(guildID)
	if err != nil {
		err = fmt.Errorf("unable to get guild from API: %s", err.Error())

		return err
	}

	origVoiceChannelOrder := orderedChannels(guildChannels).voiceChannelsSort()
	log.WithField("channels", origVoiceChannelOrder).Debugf("Original voice channel order")

	guildRoles, dErr := getGuildRoles(s, guildID)
	if dErr != nil {
		return err
	}

	origRoleOrder := orderedRoles(guildRoles).sort()
	log.WithField("roles", origRoleOrder).Debugf("Original role order")

	return nil
}

// guildCleanupMemberEphemeralRoles revokes all ephemeral roles from user in guild
func guildCleanupMemberEphemeralRoles(
	s *discordgo.Session,
	user *discordgo.User,
	guild *discordgo.Guild,
) {

	// Get guild member
	guildMember, err := s.GuildMember(guild.ID, user.ID)
	if err != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"user":  user.Username,
				"guild": guild.Name,
			}).
			Debugf("Unable to determine member in VoiceStateUpdate")

		return
	}

	// Map our member roles
	memberRoleIDs := make(map[string]bool)
	for _, roleID := range guildMember.Roles {
		memberRoleIDs[roleID] = true
	}

	// Get guild roles
	guildRoles, dErr := getGuildRoles(s, guild.ID)
	if dErr != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"user":  user.Username,
				"guild": guild.Name,
			}).
			Debugf("Unable to determine member in VoiceStateUpdate")

		return
	}

	// Check all guild roles for ephemeral roles.  If our member has this role,
	// revoke it from them
	for _, role := range guildRoles {
		if strings.HasPrefix(role.Name, ROLEPREFIX) { // Found ephemeral role
			if memberRoleIDs[role.ID] { // Member has this ephemeral role
				// Remove the role
				err := s.GuildMemberRoleRemove(guild.ID, user.ID, role.ID)
				if err != nil {
					log.WithError(err).
						WithFields(logrus.Fields{
							"user":  user.Username,
							"role":  role.Name,
							"guild": guild.Name,
						}).
						Debugf("Unable to remove role on VoiceStateUpdate")

					return
				}

				log.WithFields(logrus.Fields{
					"user":  user.Username,
					"role":  role.Name,
					"guild": guild.Name,
				}).Debugf("Removed role")
			}
		}
	}
}
