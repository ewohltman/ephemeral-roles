package callbacks

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
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

	guildRoles, err := getGuildRoles(s, vsu.GuildID)
	if err != nil {
		log.WithError(err).Debugf("Unable to determine guild roles in VoiceStateUpdate")
	}

	// Revoke all ephemeral roles from this user and start clean
	guildRoleMemberCleanup(s, user, guild, guildRoles)

	// User disconnect event
	if vsu.ChannelID == "" {
		log.WithFields(logrus.Fields{
			"user":  user.Username,
			"guild": guild.Name,
		}).Debugf("User disconnected from voice channels and ephemeral roles revoked")

		return
	}

	channel, err := s.Channel(vsu.ChannelID)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"user":  user.Username,
			"guild": guild.Name,
		}).Debugf("Unable to determine channel in VoiceStateUpdate")

		return
	}

	var ephRole *discordgo.Role

	// Check to see if the role already exists
	for _, role := range guildRoles {
		if role.Name == ROLEPREFIX+channel.Name { // Found role
			ephRole = role

			// Add member to ephemeral role
			err = s.GuildMemberRoleAdd(guild.ID, user.ID, ephRole.ID)
			if err != nil {
				log.WithError(err).WithFields(logrus.Fields{
					"user":    user.Username,
					"guild":   guild.Name,
					"channel": channel.Name,
					"role":    ephRole.Name,
				}).Debugf("Unable to add user to ephemeral role")

				return
			}

			log.WithFields(logrus.Fields{
				"user":    user.Username,
				"guild":   guild.Name,
				"channel": channel.Name,
				"role":    ephRole.Name,
			}).Debugf("User connected to voice channel and added to role")

			return
		}
	}

	// Role does not exist
	if ephRole == nil {
		var err error

		// Create and edit a new role
		ephRole, err = guildRoleCreateEdit(s, ROLEPREFIX+channel.Name, guild)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"guild": guild.Name,
				"role":  ROLEPREFIX + channel.Name,
			}).Debugf("Error managing ephemeral role")

			return
		}

		// Add our member to role
		err = s.GuildMemberRoleAdd(guild.ID, user.ID, ephRole.ID)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"user":    user.Username,
				"guild":   guild.Name,
				"channel": channel.Name,
				"role":    ephRole.Name,
			}).Debugf("Unable to add user to ephemeral role")

			return
		}

		log.WithFields(logrus.Fields{
			"user":    user.Username,
			"guild":   guild.Name,
			"channel": channel.Name,
			"role":    ephRole.Name,
		}).Debugf("User connected to voice channel and added to role")
	}

	return
}

// getGuildRoles handles role lookups using dErr *discordError as a means to
// provide context to API errors
func getGuildRoles(
	s *discordgo.Session,
	guildID string,
) (roles []*discordgo.Role, dErr *discordError) {

	roles, err := s.GuildRoles(guildID)
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

	origVoiceChannelOrder := orderedChannels(guildChannels).voiceChannels()
	sort.Stable(origVoiceChannelOrder)

	log.WithField("channels", origVoiceChannelOrder).Debugf("Original voice channel order")

	guildRoles, err := getGuildRoles(s, guildID)
	if err != nil {
		return err
	}

	origRoleOrder := orderedRoles(guildRoles)
	sort.Sort(origRoleOrder)

	log.WithField("roles", orderedRoles(origRoleOrder)).Debugf("Original role order")

	/*roleNameMap := make(map[string]*discordgo.Role)

	botRolePosition := 0
	numEphRoles := 0

	// Search for BOTNAME role, assuming the role is at or near the top of the list...
	for i := len(origRoleOrder) - 1; i >= 0; i-- {
		// Found the BOTNAME role
		if origRoleOrder[i].Name == BOTNAME {
			botRolePosition = i

			continue
		}

		// Found an ephemeral role
		if strings.HasPrefix(origRoleOrder[i].Name, ROLEPREFIX) {
			numEphRoles++
		}
	}

	if botRolePosition == 0 {
		err = fmt.Errorf("unable to get find bot role in guild: %s", err.Error())

		return
	}

	ephRolesOrdered := make([]*discordgo.Role, 0, len(guildRoles))

	for i := 0; i < len(voiceChannelOrder); i++ {
		roleName := ROLEPREFIX + voiceChannelOrder[i].Name

		if ephRole, found := roleNameMap[roleName]; found {
			ephRolesOrdered = append(ephRolesOrdered, ephRole)
		}
	}*/

	/*newRoleOrder := make(orderedRoles, 0, len(guildRoles))

	// roleOrder[0] == @everybody
	newRoleOrder = append(newRoleOrder, origRoleOrder[0])

	for i := 1; i <= botRolePosition-numEphRoles-1; i++ {
		roleOrder[i].Position = i - 1
		newRoleOrder = append(newRoleOrder, roleOrder[i])
	}

	// Add our ordered roles
	for i := 0; i < len(ephRolesOrdered); i++ {
		ephRolesOrdered[i].Position = (botRolePosition - numEphRoles) + i
		newRoleOrder = append(newRoleOrder, ephRolesOrdered[i])
	}

	// Add the remaining roles above us
	for i := botRolePosition; i < len(roleOrder); i++ {
		roleOrder[i].Position = i
		newRoleOrder = append(newRoleOrder, roleOrder[i])
	}

	log.WithField("roles", orderedRoles(newRoleOrder)).Debugf("New role order")

	reorderedRoles, err := s.GuildRoleReorder(guild.ID, newRoleOrder)
	if err != nil {
		log.WithError(err).
			WithField("newRoleOrder", newRoleOrder).
			Debugf("Unable to order new channel")
	}

	log.WithField("roles", orderedRoles(reorderedRoles)).Debugf("Reordered role order")*/

	return
}

// guildRoleMemberCleanup revokes all ephemeral roles from user in guild
func guildRoleMemberCleanup(
	s *discordgo.Session,
	user *discordgo.User,
	guild *discordgo.Guild,
	guildRoles []*discordgo.Role,
) {

	guildMember, err := s.GuildMember(guild.ID, user.ID)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"user":  user.Username,
			"guild": guild.Name,
		}).Debugf("Unable to determine member in VoiceStateUpdate")

		return
	}

	// Map our member roles
	memberRoleIDs := make(map[string]bool)
	for _, roleID := range guildMember.Roles {
		memberRoleIDs[roleID] = true
	}

	for _, role := range guildRoles {
		if strings.HasPrefix(role.Name, ROLEPREFIX) { // An ephemeral role
			if memberRoleIDs[role.ID] { // Our user belongs to this role
				err := s.GuildMemberRoleRemove(guild.ID, user.ID, role.ID)
				if err != nil {
					log.WithError(err).WithFields(logrus.Fields{
						"user":  user.Username,
						"guild": guild.Name,
						"role":  role.Name,
					}).Debugf("Unable to remove role on VoiceStateUpdate")

					return
				}

				log.WithFields(logrus.Fields{
					"user":  user.Username,
					"guild": guild.Name,
					"role":  role.Name,
				}).Debugf("Removed role")
			}
		}
	}
}
