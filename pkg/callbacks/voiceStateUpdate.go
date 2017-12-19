package callbacks

import (
	"os"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

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

	// Get the roles for this guild
	roles, err := s.GuildRoles(vsu.GuildID)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"user":  user.Username,
			"guild": guild.Name,
		}).Errorf("Unable to determine roles from guild in VoiceStateUpdate")

		return
	}

	channelName := vsu.ChannelID

	if channelName != "" {
		// User connect/change event
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

	// Map out this guild's roles
	guildRoleNameToID := make(map[string]string)
	guildRoleIDToRole := make(map[string]*discordgo.Role)
	guildRoleOriginalOrder := make(map[int]*discordgo.Role)

	for _, role := range roles {
		guildRoleNameToID[role.Name] = role.ID
		guildRoleIDToRole[role.ID] = role
		guildRoleOriginalOrder[role.Position] = role
	}

	var ephRole *discordgo.Role
	ephRoleName := botChannelPrefix + " " + channelName

	// Does this guild have our intended ephemeral role?
	if intendedRole, found := guildRoleNameToID[ephRoleName]; found && channelName != "" {
		ephRole = guildRoleIDToRole[intendedRole]
	}

	// If we did not find it, create it
	if ephRole == nil && channelName != "" {
		// Create a new blank role
		ephRole, err = s.GuildRoleCreate(vsu.GuildID)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"user":    user.Username,
				"guild":   guild.Name,
				"channel": channelName,
			}).Errorf("Unable to create ephemeral role")

			return
		}

		// Edit the blank role
		color := 16753920 // Default to orange hex #FFA500 in decimal
		if colorString, found := os.LookupEnv("EPH_CHANNEL_COLOR_HEX2DEC"); found {
			parsedString, err := strconv.Atoi(colorString)
			if err != nil {
				log.WithError(err).
					WithField("EPH_CHANNEL_COLOR_HEX2DEC", colorString).
					Warnf("Error parsing EPH_CHANNEL_COLOR_HEX2DEC from environment")
			} else {
				color = parsedString
			}
		}

		ephRole, err = s.GuildRoleEdit(
			vsu.GuildID,
			ephRole.ID,
			ephRoleName,
			color, // Orange hex #FFA500 to decimal
			true,
			ephRole.Permissions,
			ephRole.Mentionable,
		)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"user":    user.Username,
				"guild":   guild.Name,
				"channel": channelName,
			}).Errorf("Unable to edit ephemeral role")

			return
		}

		// TODO: Figure out why the following section on ordering roles does not work
		/*
			// Check if EPH_DISPLAY_AFTER was provided for ordering, otherwise we're done
			roleDisplayAfter := os.Getenv("EPH_DISPLAY_AFTER")

			if roleDisplayAfter != "" {
				// Reorder them to be below slot roleDisplayAfter
				guildNewRoleOrder := make([]*discordgo.Role, 0, len(roles)+1)

				// Find our slot under in the "stack" for our role
				// The highest in the UI list has the greatest positional parameter
				for i := len(guildRoleOriginalOrder); i >= 0; i-- {
					if guildRoleOriginalOrder[i] != nil && guildRoleOriginalOrder[i].Name == roleDisplayAfter {
						// Add the lower positions
						for j := 0; j < i; j++ {
							if guildRoleOriginalOrder[j] != nil {
								originalCopy := *guildRoleOriginalOrder[j]
								originalCopy.Position = j

								guildNewRoleOrder = append(guildNewRoleOrder, &originalCopy)
							}
						}

						// Add our new position
						ephRole.Position = i
						guildNewRoleOrder = append(guildNewRoleOrder, ephRole)

						// Update the remaining positions
						for k := i + 1; k <= len(guildRoleOriginalOrder)+1; k++ {
							if guildRoleOriginalOrder[k-1] != nil {
								originalCopy := *guildRoleOriginalOrder[k-1]
								originalCopy.Position = k

								guildNewRoleOrder = append(guildNewRoleOrder, &originalCopy)
							}
						}

						break // No need to go any further after the critical section
					}
				}

				for testIndex := 0; testIndex < len(guildNewRoleOrder); testIndex++ {
					log.Infof("guildNewRoleOrder[%d] = %+v", testIndex, *guildNewRoleOrder[testIndex])
				}

				// Set the new role order
				orderedRoles, err := s.GuildRoleReorder(vsu.GuildID, guildNewRoleOrder)
				if err != nil {
					log.WithError(err).WithFields(logrus.Fields{
						"user":    user.Username,
						"guild":   guild.Name,
						"channel": channelName,
					}).Errorf("Unable to order new ephemeral role")
				} else {
					// Reset our maps
					guildRoleNameToID = make(map[string]string)
					guildRoleIDToRole = make(map[string]*discordgo.Role)

					for _, role := range orderedRoles {
						guildRoleNameToID[role.Name] = role.ID
						guildRoleIDToRole[role.ID] = role
					}
				}
			}
		*/
	}

	// Check to see if we need to add this user to the ephemeral role
	foundInMemberRoles := false
	for _, member := range guild.Members {
		if member.User.ID != user.ID {
			continue
		}

		// Found our member, check roles
		for _, memberRoleID := range member.Roles {
			guildRole, found := guildRoleIDToRole[memberRoleID]
			if !found {
				log.WithFields(logrus.Fields{
					"user":   user.Username,
					"guild":  guild.Name,
					"roleID": memberRoleID,
				}).Debugf("Role ID not found in guild")

				continue
			}

			// Is this the role we're looking for?
			if guildRole == ephRole {
				foundInMemberRoles = true

				continue
			}

			// While we're here, let's check to see if we can clean up
			if strings.HasPrefix(guildRole.Name, botChannelPrefix+" ") {
				err = s.GuildMemberRoleRemove(vsu.GuildID, vsu.UserID, guildRole.ID)
				if err != nil {
					log.WithError(err).WithFields(logrus.Fields{
						"user":  user.Username,
						"guild": guild.Name,
						"role":  guildRole.Name,
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
		}).Infof("User disconnected from voice channels and ephemeral roles revoked")

		return
	}

	// Add user to role
	if !foundInMemberRoles {
		err = s.GuildMemberRoleAdd(vsu.GuildID, vsu.UserID, ephRole.ID)
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
		}).Infof("User connected to voice channel and added to role")

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
