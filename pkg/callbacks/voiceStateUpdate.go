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

// VoiceStateUpdate is the callback function for the "voice state update" event from Discord
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

	// Context-appropriate logging is handled within getGuildRoles()
	roles, err := getGuildRoles(s, guild)
	if err != nil {
		return
	}

	// Map out this guild's roles
	guildRoleNameToID := make(map[string]string)
	guildRoleIDToRole := make(map[string]*discordgo.Role)
	guildRoleOriginalOrder := make(map[int]*discordgo.Role)

	for _, role := range roles {
		guildRoleNameToID[role.Name] = role.ID
		guildRoleIDToRole[role.ID] = role
		guildRoleOriginalOrder[role.Position] = role
	}

	channelName := vsu.ChannelID

	if channelName != "" {
		// User connect or change event
		channel, err := s.Channel(vsu.ChannelID)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"user":  user.Username,
				"guild": guild.Name,
			}).Errorf("Unable to determine channel in VoiceStateUpdate")

			return
		}

		channelName = channel.Name
	}

	ephRoleName := ROLE_PREFIX + " " + channelName
	var ephRole *discordgo.Role

	// Does this guild have our intended ephemeral role?
	if existingRole, found := guildRoleNameToID[ephRoleName]; found && channelName != "" {
		ephRole = guildRoleIDToRole[existingRole]
	}

	// If we did not find it, create it
	if ephRole == nil && channelName != "" {
		var err error

		ephRole, err = guildRoleCreateEdit(s, ephRoleName, guild)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"guild": guild.Name,
				"role":  ephRoleName,
			}).Errorf("Error managing ephemeral role")

			return
		}
	}

	// Check to see if we need to add this user to the ephemeral role
	foundInMemberRoles := false
	for _, member := range guild.Members {
		if member.User.ID != user.ID {
			continue
		}

		// Found our member, check roles
		for _, memberRoleID := range member.Roles {
			role, found := guildRoleIDToRole[memberRoleID]
			if !found {
				log.WithFields(logrus.Fields{
					"user":  user.Username,
					"guild": guild.Name,
				}).Debugf("Role not found in current guild")

				continue
			}

			// Is this the role we're looking for?
			if role == ephRole {
				foundInMemberRoles = true

				continue
			}

			// While we're here, let's check to see if we can clean up
			if strings.HasPrefix(role.Name, ROLE_PREFIX+" ") {
				err = s.GuildMemberRoleRemove(guild.ID, user.ID, role.ID)
				if err != nil {
					log.WithError(err).WithFields(logrus.Fields{
						"user":  user.Username,
						"guild": guild.Name,
						"role":  role.Name,
					}).Errorf("Unable to remove role on VoiceStateUpdate")
				}

				log.WithFields(logrus.Fields{
					"user":     user.Username,
					"guild":    guild.Name,
					"roleName": memberRoleID,
					"roleID":   guildRoleNameToID[memberRoleID],
				}).Debugf("Removed role")
			}
		}

		break
	}

	// User disconnect event
	if channelName == "" {
		log.WithFields(logrus.Fields{
			"user":  user.Username,
			"guild": guild.Name,
		}).Debugf("User disconnected from voice channels and ephemeral roles revoked")

		return
	}

	// Add user to role
	if !foundInMemberRoles {
		err = s.GuildMemberRoleAdd(guild.ID, user.ID, ephRole.ID)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"user":    user.Username,
				"guild":   guild.Name,
				"channel": channelName,
				"role":    ephRole.Name,
			}).Errorf("Unable to add user to ephemeral role")

			return
		}

		log.WithFields(logrus.Fields{
			"user":    user.Username,
			"guild":   guild.Name,
			"channel": channelName,
			"role":    ephRole.Name,
		}).Debugf("User connected to voice channel and added to role")

		return
	}

	// User already in role
	log.WithFields(logrus.Fields{
		"user":    user.Username,
		"guild":   guild.Name,
		"channel": channelName,
		"status": logrus.Fields{
			"suppress": vsu.Suppress,
			"deaf":     vsu.Deaf,
			"mute":     vsu.Mute,
			"sealDeaf": vsu.SelfDeaf,
			"selfMute": vsu.SelfMute,
		},
	}).Debugf("User changed status in voice channel")
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
