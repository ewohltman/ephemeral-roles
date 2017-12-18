package callbacks

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// VoiceStateUpdate is the callback function for the "voice state update" event from Discord
func VoiceStateUpdate(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	user, err := s.User(vsu.UserID)
	if err != nil {
		log.WithError(err).Errorf("Unable to determine user in VoiceStateUpdate")

		return
	}

	// User disconnect?
	if vsu.ChannelID == "" {
		log.WithFields(logrus.Fields{
			"user": user.Username,
		}).Debugf("User disconnected from voice channel")

		// TODO: Remove any temp roles

		return
	}

	channel, err := s.Channel(vsu.ChannelID)
	if err != nil {
		log.WithError(err).Errorf("Unable to determine channel in VoiceStateUpdate")

		return
	}

	log.WithFields(logrus.Fields{
		"user":    user.Username,
		"channel": channel.Name,
	}).Debugf("User connected to voice channel")

	// TODO: Check and add any temp roles
}
