package callbacks

import (
	"os"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const defaultRoleColor = 16753920 // Default to orange hex #FFA500 in decimal

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

	// Get the user
	user, err := s.User(vsu.UserID)
	if err != nil {
		config.Log.WithError(err).Debugf("Unable to determine user in VoiceStateUpdate")
		return
	}

	// Get the guild
	guild, err := s.Guild(vsu.GuildID)
	if err != nil {
		config.Log.WithError(err).Debugf("Unable to determine guild in VoiceStateUpdate")
		return
	}

	found := false
	var guildMember *discordgo.Member

	for _, member := range guild.Members {
		if member.User.ID == user.ID {
			found = true
			guildMember = member
			break
		}
	}

	if !found {
		config.Log.WithFields(logrus.Fields{
			"user":  user.Username,
			"guild": guild.Name,
		}).Debugf("User not found in guild members")

		return
	}

	event := &vsuEvent{
		Session:      s,
		Guild:        guild,
		GuildMember:  guildMember,
		GuildRoleMap: make(map[string]*discordgo.Role),
	}

	for _, role := range event.Guild.Roles {
		event.GuildRoleMap[role.ID] = role
	}

	// Check if user disconnect event
	if vsu.ChannelID == "" {
		config.revokeEphemeralRoles(event)

		config.Log.WithFields(logrus.Fields{
			"user":  user.Username,
			"guild": guild.Name,
		}).Debugf("User disconnected from voice channels and ephemeral roles revoked")

		return
	}

	// Get the channel
	channel, err := s.Channel(vsu.ChannelID)
	if err != nil {
		config.Log.WithError(err).WithFields(logrus.Fields{
			"user":  user.Username,
			"guild": guild.Name,
		}).Debugf("Unable to determine channel in VoiceStateUpdate")

		return
	}

	ephRoleName := config.RolePrefix + " " + channel.Name

	// Check to see if the member already has the role
	for _, memberRoleID := range event.GuildMember.Roles {
		if event.GuildRoleMap[memberRoleID].Name == ephRoleName {
			return
		}
	}

	// Check to see if the role already exists in the guild
	for _, guildRole := range event.GuildRoleMap {
		if guildRole.Name == ephRoleName {
			// Ephemeral role exists, add member to it
			config.grantEphemeralRole(event, guildRole)
			return
		}
	}

	// Ephemeral role does not exist, create and edit it
	ephRole, err := config.guildRoleCreateEdit(event, ephRoleName)
	if err != nil {
		config.Log.WithError(err).WithFields(logrus.Fields{
			"role":  ephRoleName,
			"guild": guild.Name,
		}).Debugf("Unable to manage ephemeral role")

		return
	}

	// Add role to member
	config.grantEphemeralRole(event, ephRole)
}

func (config *Config) guildRoleCreateEdit(event *vsuEvent, ephRoleName string) (*discordgo.Role, error) {
	// Create a new blank role
	ephRole, err := event.Session.GuildRoleCreate(event.Guild.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create ephemeral role: "+err.Error())
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
		return nil, errors.New("unable to edit ephemeral role: " + err.Error())
	}

	/*err = guildRolesReorder(s, guild.ID)
	if err != nil {
		return nil, errors.New("unable to reorder ephemeral role: " + err.Error())
	}*/

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
						"guild": event.Guild.Name,
						"role":  role.Name,
					}).Debugf("Unable to remove role on VoiceStateUpdate")

				return
			}

			config.Log.WithFields(logrus.Fields{
				"user":  event.GuildMember.User.Username,
				"guild": event.Guild.Name,
				"role":  role.Name,
			}).Debugf("Removed role")
		}
	}
}

func (config *Config) grantEphemeralRole(event *vsuEvent, ephRole *discordgo.Role) {
	// Revoke any previous ephemeral roles
	config.revokeEphemeralRoles(event)

	// Add our member to role
	err := event.Session.GuildMemberRoleAdd(event.Guild.ID, event.GuildMember.User.ID, ephRole.ID)
	if err != nil {
		config.Log.WithError(err).WithFields(logrus.Fields{
			"user":  event.GuildMember.User.Username,
			"role":  ephRole.Name,
			"guild": event.Guild.Name,
		}).Debugf("Unable to add user to ephemeral role")

		return
	}

	config.Log.WithFields(logrus.Fields{
		"user":  event.GuildMember.User.Username,
		"role":  ephRole.Name,
		"guild": event.Guild.Name,
	}).Debugf("Added role")
}

/*func guildRolesReorder(s *discordgo.Session, guildID string) error {
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
		err = errors.New("unable to reorder guild roles from API: " + err.Error())

		return err
	}

	return nil
}*/
