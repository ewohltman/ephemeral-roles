package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/discordEphemeralRolesProject/pkg/callbacks"
	"github.com/ewohltman/discordEphemeralRolesProject/pkg/logging"
)

var log = logging.Instance()

func main() {
	// Check for DERP_BOT_TOKEN, we need this to connect to Discord
	token, found := os.LookupEnv("DERP_BOT_TOKEN")
	if !found || token == "" {
		log.Fatalf("DERP_BOT_TOKEN not defined in environment variables")
	}

	// Check for DERP_BOT_KEYWORD, we don't need it now but it's required in the callbacks
	_, found = os.LookupEnv("DERP_BOT_KEYWORD")
	if !found {
		log.Fatalf("DERP_BOT_KEYWORD not defined in environment variables")
	}

	// Check for DERP_CHANNEL_PREFIX, we don't need it now but it's required in the callbacks
	_, found = os.LookupEnv("DERP_CHANNEL_PREFIX")
	if !found {
		log.Fatalf("DERP_CHANNEL_PREFIX not defined in environment variables")
	}

	// Create a new Discord session using the provided bot token.
	dgBot, err := discordgo.New("Bot " + token)
	if err != nil {
		log.WithError(err).Fatalf("Error creating Discord session")
	}

	// Add event handlers
	dgBot.AddHandler(callbacks.Ready)
	dgBot.AddHandler(callbacks.MessageCreate)
	dgBot.AddHandler(callbacks.VoiceStateUpdate)

	// Open the websocket and begin listening.
	err = dgBot.Open()
	if err != nil {
		log.WithError(err).Fatalf("Error opening Discord session")
	}
	defer dgBot.Close() // Cleanly close down the Discord session.

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("dERP is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	log.Infof("Caught signal for graceful shutdown")

	return
}
