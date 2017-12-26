package callbacks

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
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
		log.WithError(err).Errorf("Unable to determine user in VoiceStateUpdate")

		return
	}

	// Get the guild
	guild, err := s.Guild(vsu.GuildID)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"user": user.Username,
		}).Errorf("Unable to determine guild in VoiceStateUpdate")

		return
	}

	// Context-appropriate logging is handled within getGuildRoles
	guildRoles, err := getGuildRoles(s, guild)
	if err != nil {
		return
	}

	// User disconnect event
	if vsu.ChannelID == "" {
		log.WithFields(logrus.Fields{
			"user":  user.Username,
			"guild": guild.Name,
		}).Debugf("User disconnected from voice channels and ephemeral roles revoked")

		// Revoke all ephemeral roles from this user
		guildRoleMemberCleanup(s, user, guild, guildRoles)

		return
	}

	channel, err := s.Channel(vsu.ChannelID)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"user":  user.Username,
			"guild": guild.Name,
		}).Errorf("Unable to determine channel in VoiceStateUpdate")

		return
	}

	var ephRole *discordgo.Role

	for _, role := range guildRoles {
		// Role already exists
		if role.Name == ROLE_PREFIX+channel.Name {
			ephRole = role

			// User not in role, revoke any existing ephemeral roles
			guildRoleMemberCleanup(s, user, guild, guildRoles)

			// Add member to ephemeral role
			err = s.GuildMemberRoleAdd(guild.ID, user.ID, ephRole.ID)
			if err != nil {
				log.WithError(err).WithFields(logrus.Fields{
					"user":    user.Username,
					"guild":   guild.Name,
					"channel": channel.Name,
					"role":    ephRole.Name,
				}).Errorf("Unable to add user to ephemeral role")

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
		ephRole, err = guildRoleCreateEdit(s, ROLE_PREFIX+channel.Name, guild)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"guild": guild.Name,
				"role":  ROLE_PREFIX + channel.Name,
			}).Errorf("Error managing ephemeral role")

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
			}).Errorf("Unable to add user to ephemeral role")

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
			}).Errorf("Unable to unmarshal Discord API JSON response while determining roles in guild")

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
		}).Errorf("Unable to determine roles in guild")

		return
	}

	return
}

func guildRoleCreateEdit(
	s *discordgo.Session,
	ephRoleName string,
	guild *discordgo.Guild) (ephRole *discordgo.Role, err error) {

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

	// Organize the role within the existing roles
	// Find all voice channels
	guildChannels, err := s.GuildChannels(guild.ID)
	if err != nil {
		err = fmt.Errorf("unable to get guild channels: %s", err.Error())

		return
	}

	currentChannelOrder := make(map[int]*discordgo.Channel)

	for _, channel := range guildChannels {
		if channel.Type != discordgo.ChannelTypeGuildVoice {
			continue
		}

		currentChannelOrder[channel.Position] = channel
	}

	channelOrderListString := ""
	for i := 0; i < len(currentChannelOrder); i++ {
		channelOrderListString = fmt.Sprintf(
			"%s\norder: %d, name: %s",
			channelOrderListString,
			i,
			currentChannelOrder[i].Name,
		)
	}

	log.WithField("channels", channelOrderListString).Debugf("Current channel order")

	guildRoles, err := s.GuildRoles(guild.ID)
	if err != nil {
		err = fmt.Errorf("unable to get guild roles: %s", err.Error())

		return
	}

	currentRoleOrder := make(map[int]*discordgo.Role)
	botRolePosition := -1

	for _, role := range guildRoles {
		currentRoleOrder[role.Position] = role

		if role.Name == BOT_NAME {
			botRolePosition = role.Position
		}
	}

	if botRolePosition == -1 {
		err = fmt.Errorf("unable to get find bot role in guild: %s", err.Error())

		return
	}

	newRoleOrder := make([]*discordgo.Role, 0, len(currentRoleOrder)+1)

	// Add the pre-existing lower roles
	for i := 0; i < botRolePosition; i++ {
		if currentRoleOrder[i] != nil {
			if i == 0 { // @everybody
				newRoleOrder = append(newRoleOrder, currentRoleOrder[i])
				continue
			}

			if currentRoleOrder[i].Name == ephRole.Name {
				continue // Manually insert later
			}

			currentRoleOrder[i].Position = currentRoleOrder[i].Position - 1
			newRoleOrder = append(newRoleOrder, currentRoleOrder[i])
		}
	}

	// Add our new role
	ephRole.Position = botRolePosition - 1
	newRoleOrder = append(newRoleOrder, ephRole)

	// Add the bot role and remaining higher roles
	for j := botRolePosition; j < len(newRoleOrder); j++ {
		newRoleOrder = append(newRoleOrder, currentRoleOrder[j])
	}

	roleOrderListString := ""
	for i := 0; i < len(currentRoleOrder); i++ {
		if currentRoleOrder[i] != nil {
			roleOrderListString = fmt.Sprintf(
				"%s\norder: %d, name: %s",
				roleOrderListString,
				i,
				currentRoleOrder[i].Name,
			)
		}
	}

	log.WithField("roles", roleOrderListString).Debugf("Current role order")

	newRoleOrderListString := ""
	for i := 0; i < len(newRoleOrder); i++ {
		if newRoleOrder[i] != nil {
			newRoleOrderListString = fmt.Sprintf(
				"%s\norder: %d, name: %s",
				newRoleOrderListString,
				i,
				newRoleOrder[i].Name,
			)
		}
	}

	log.WithField("roles", newRoleOrderListString).Debugf("New role order")

	_, err = s.GuildRoleReorder(guild.ID, newRoleOrder)
	if err != nil {
		log.WithError(err).WithField("newRoleOrder", newRoleOrder).Errorf("Unable to order new channel")
	}

	return
}

// guildRoleMemberCleanup revokes all ephemeral roles from user in guild
func guildRoleMemberCleanup(s *discordgo.Session, user *discordgo.User, guild *discordgo.Guild, guildRoles []*discordgo.Role) {
	guildMember, err := s.GuildMember(guild.ID, user.ID)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"user":  user.Username,
			"guild": guild.Name,
		}).Errorf("Unable to determine member in VoiceStateUpdate")

		return
	}

	// Map our member roles
	memberRoleIDs := make(map[string]bool)
	for _, roleID := range guildMember.Roles {
		memberRoleIDs[roleID] = true
	}

	for _, role := range guildRoles {
		if strings.HasPrefix(role.Name, ROLE_PREFIX) { // An ephemeral role
			if memberRoleIDs[role.ID] { // Our user belongs to this role
				err := s.GuildMemberRoleRemove(guild.ID, user.ID, role.ID)
				if err != nil {
					log.WithError(err).WithFields(logrus.Fields{
						"user":  user.Username,
						"guild": guild.Name,
						"role":  role.Name,
					}).Errorf("Unable to remove role on VoiceStateUpdate")
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
