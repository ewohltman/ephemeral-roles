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

// DiscordAPIResponse is a helper struct for encapsulating API responses in use
// with logging
type DiscordAPIResponse struct {
	Code    int
	Message string
}

type discordError struct {
	HTTPResponseMessage string
	APIResponse         *DiscordAPIResponse
}

// String implements the Stringer interface for field names in logs
func (dAR *DiscordAPIResponse) String() string {
	return fmt.Sprintf("Code: %d, Message: %s", dAR.Code, dAR.Message)
}

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
		log.WithError(err).WithFields(logrus.Fields{
			"user": user.Username,
		}).Debugf("Unable to determine guild in VoiceStateUpdate")

		return
	}

	guildRoles, err := getGuildRoles(s, guild)
	if err != nil {
		// Context-appropriate logging is handled within getGuildRoles
		return
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

	for _, role := range guildRoles {
		// Role already exists
		if role.Name == ROLEPREFIX+channel.Name {
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

		return
	}

	return
}

// getGuildRoles handles role lookups is a graceful way
//
// Logging is handled within this function so the caller should handle the
// error in some other way
func getGuildRoles(
	s *discordgo.Session,
	guild *discordgo.Guild,
) (roles []*discordgo.Role, err error) {

	roles, err = s.GuildRoles(guild.ID)
	if err != nil {
		// Find the JSON with regular expressions
		rx := regexp.MustCompile("{.*}")
		errHTTPString := rx.ReplaceAllString(err.Error(), "")
		errJSONString := rx.FindString(err.Error())

		dErr := &discordError{
			HTTPResponseMessage: errHTTPString,
			APIResponse:         &DiscordAPIResponse{},
		}

		unmarshalErr := json.Unmarshal([]byte(errJSONString), dErr.APIResponse)

		// Unable to unmarshal the API response
		if unmarshalErr != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"guild": guild.Name,
				"json":  errJSONString,
			}).Debugf("Unable to unmarshal Discord API JSON response while determining roles in guild")

			return
		}

		// Code 50013: "Missing Permissions"
		if dErr.APIResponse.Code == 50013 {
			log.WithFields(logrus.Fields{
				"guild":     guild.Name,
				"api_error": *dErr.APIResponse,
			}).Debugf("Insufficient privileged role to query guild roles")

			return
		}

		// Catch all other error codes
		log.WithFields(logrus.Fields{
			"guild":      guild.Name,
			"http_error": dErr.HTTPResponseMessage,
			"api_error":  *dErr.APIResponse,
		}).Debugf("Unable to determine roles in guild")

		return
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

type orderedRoles []*discordgo.Role

// String satisfies the Stringer interface for orderedRoles
func (oR orderedRoles) String() string {
	str := ""

	positions := make(map[int]*discordgo.Role)

	for _, role := range oR {
		positions[role.Position] = role
	}

	for i := len(positions) - 1; i >= 0; i-- {
		if positions[i] != nil {
			str = fmt.Sprintf(
				"%s\nPosition: %d, Name: %s",
				str,
				positions[i].Position,
				positions[i].Name,
			)
		} else {
			str = fmt.Sprintf(
				"%s\nPosition: %s, Name: %s",
				str,
				"nil",
				"nil",
			)
		}
	}

	return str
}

// guildRoleReorder orders roles in the order in which the channels appear
func guildRoleReorder(s *discordgo.Session, guildID string) (err error) {
	// Get guild from our internal state
	guild, err := s.State.Guild(guildID)
	if err != nil {
		err = fmt.Errorf("unable to get guild from internal state: %s", err.Error())

		return
	}

	log.WithField("roles", orderedRoles(guild.Roles)).Debugf("Original role order")

	voiceChannelOrder := make(map[int]*discordgo.Channel)

	// Find the order of all voice channels
	for _, channel := range guild.Channels {
		if channel.Type != discordgo.ChannelTypeGuildVoice {
			continue
		}

		voiceChannelOrder[channel.Position] = channel
	}

	voiceChannelOrderString := ""
	for i := 0; i < len(voiceChannelOrder); i++ {
		voiceChannelOrderString = fmt.Sprintf(
			"%s\norder: %d, name: %s",
			voiceChannelOrderString,
			i,
			voiceChannelOrder[i].Name,
		)
	}

	log.WithField("channels", voiceChannelOrderString).Debugf("Current channel order")

	roleOrder := make(map[int]*discordgo.Role)
	roleNameMap := make(map[string]*discordgo.Role)

	botRolePosition := 0
	numEphRoles := 0

	log.WithField("roles", orderedRoles(guild.Roles)).Debugf("Original role order")

	for _, role := range guild.Roles {
		roleOrder[role.Position] = role
		roleNameMap[role.Name] = role

		// Found the BOTNAME role
		if role.Name == BOTNAME {
			botRolePosition = role.Position

			continue
		}

		// Found an ephemeral role
		if strings.HasPrefix(role.Name, ROLEPREFIX) {
			numEphRoles++
		}
	}

	if botRolePosition == 0 {
		err = fmt.Errorf("unable to get find bot role in guild: %s", err.Error())

		return
	}

	ephRolesOrdered := make([]*discordgo.Role, 0, len(guild.Roles))

	for i := 0; i < len(voiceChannelOrder); i++ {
		roleName := ROLEPREFIX + voiceChannelOrder[i].Name

		if ephRole, found := roleNameMap[roleName]; found {
			ephRolesOrdered = append(ephRolesOrdered, ephRole)
		}
	}

	newRoleOrder := make([]*discordgo.Role, 0, len(guild.Roles))

	// roleOrder[0] == @everybody
	newRoleOrder = append(newRoleOrder, roleOrder[0])

	for i := 1; i <= botRolePosition-numEphRoles; i++ {
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

	log.WithField("roles", orderedRoles(reorderedRoles)).Debugf("Reordered role order")

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
