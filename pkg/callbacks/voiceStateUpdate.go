package callbacks

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
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

	if vsu.ChannelID == "" { // User disconnect
		log.WithFields(logrus.Fields{
			"user":  user.Username,
			"guild": guild.Name,
		}).Debugf("User disconnected from voice channels and ephemeral roles revoked")

		cleanupMemberEphemeralRoles(s, user, guild)

		return
	}

	// Get the guild's roles
	guildRoles, dErr := guildRoles(s, vsu.GuildID)
	if dErr != nil {
		log.WithError(dErr).
			Debugf("Unable to determine guild roles in VoiceStateUpdate")

		return
	}

	// Get the guild member's roles
	memberRoles, err := guildMemberRoles(s, user, guild)
	if err != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"user":  user.Username,
				"guild": guild.Name,
			}).
			Debugf("Unable to determine guild member roles")

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

	ephRoleName := ROLEPREFIX + channel.Name
	var ephRole *discordgo.Role

	// Check to see if the role already exists in the guild
	for _, gRole := range guildRoles {
		if gRole.Name == ephRoleName {
			// Found role
			ephRole = gRole

			// Check to see if the member already has the role
			for _, mRole := range memberRoles {
				if mRole.ID == gRole.ID {
					// No effective change
					return
				}
			}

			// Revoke any previous ephemeral roles
			cleanupMemberEphemeralRoles(s, user, guild)

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

		// Revoke any previous ephemeral roles
		cleanupMemberEphemeralRoles(s, user, guild)

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

func guildMemberRoles(s *discordgo.Session, user *discordgo.User, guild *discordgo.Guild) ([]*discordgo.Role, error) {
	// Get guild member
	guildMember, err := s.GuildMember(guild.ID, user.ID)
	if err != nil {
		retErr := errors.New(
			"unable to determine member in VoiceStateUpdate: " +
				err.Error(),
		)

		return make([]*discordgo.Role, 0), retErr
	}

	// Get guild roles
	guildRoles, dErr := guildRoles(s, guild.ID)
	if dErr != nil {
		retErr := errors.New(
			"unable to determine roles in VoiceStateUpdate: " +
				dErr.Error(),
		)

		return make([]*discordgo.Role, 0), retErr
	}

	// Map our member roles
	memberRoleIDs := make(map[string]bool)
	for _, roleID := range guildMember.Roles {
		memberRoleIDs[roleID] = true
	}

	memberRoles := make([]*discordgo.Role, 0)

	for _, role := range guildRoles {
		if memberRoleIDs[role.ID] {
			memberRoles = append(memberRoles, role)
		}
	}

	return memberRoles, nil
}

// cleanupMemberEphemeralRoles revokes all ephemeral roles from user in guild
func cleanupMemberEphemeralRoles(s *discordgo.Session, user *discordgo.User, guild *discordgo.Guild) {
	guildRoles, err := guildMemberRoles(s, user, guild)
	if err != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"user":  user.Username,
				"guild": guild.Name,
			}).
			Debugf("Unable to determine guild member roles")

		return
	}

	// Check all guild roles for ephemeral roles.  If our member has this role,
	// revoke it from them
	for _, role := range guildRoles {
		if strings.HasPrefix(role.Name, ROLEPREFIX) { // Found ephemeral role
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

// guildRoles handles role lookups using dErr *discordError as a means to
// provide context to API errors
func guildRoles(s *discordgo.Session, guildID string) (roles []*discordgo.Role, dErr *discordError) {
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
	guildRoles, dErr := guildRoles(s, guildID)
	if dErr != nil {
		return errors.New(dErr.Error())
	}

	roles := orderedRoles(guildRoles)

	log.WithField("roles", roles).Debugf("Old role order")

	sort.SliceStable(
		roles,
		func(i, j int) bool {
			return roles[i].Position < roles[j].Position
		},
	)

	// Alignment correction if Discord is slow to update
	for index, role := range roles {
		if role.Position != index {
			role.Position = index
		}
	}

	for index, role := range roles {
		if role.Name == "@everyone" && role.Position != 0 { // @everyone should be the lowest
			roles.swap(index, 0)
		}

		if role.Name == BOTNAME && role.Position != len(roles)-1 { // BOTNAME should be the highest
			roles.swap(index, len(roles)-1)
		}
	}

	// Bubble the ephemeral roles up
	for index, role := range roles {
		if strings.HasPrefix(role.Name, ROLEPREFIX) {
			for j := index; j < len(roles)-2; j++ {
				// Stop bubbling at the bottom of the top-most group
				if !strings.HasPrefix(roles[j+1].Name, ROLEPREFIX) {
					roles.swap(j, j+1)
				}
			}
		}
	}

	log.WithField("roles", roles).Debugf("New role order")

	_, err := s.GuildRoleReorder(guildID, roles)
	if err != nil {
		err = fmt.Errorf("unable to reorder guild roles from API: %s", err.Error())

		return err
	}

	return nil
}
