package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/ephemeral-roles/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/ewohltman/ephemeral-roles/pkg/server"
)

func main() {
	var log = logging.Instance()

	log.Debugf("Bot starting up")

	// Check for BOT_TOKEN, we need this to connect to Discord
	token, found := os.LookupEnv("BOT_TOKEN")
	if !found || token == "" {
		log.Fatalf("BOT_TOKEN not defined in environment variables")
	}

	// Check for string from slice, these are not needed now, but are needed in the callbacks
	for _, envVar := range []string{"BOT_NAME", "BOT_KEYWORD", "ROLE_PREFIX"} {
		v, found := os.LookupEnv(envVar)
		if !found || v == "" {
			log.Fatalf("%s not defined in environment variables", envVar)
		}
	}

	// Check for PORT, we need this to for our HTTP server in our container
	port, found := os.LookupEnv("PORT")
	if !found || port == "" {
		port = "8080"
	}

	// Check for DISCORDBOTS_ORG_TOKEN and BOT_ID, we need these for optional discordbots.org integration
	discordBotsToken := ""
	botID := ""

	discordBotsToken, found = os.LookupEnv("DISCORDBOTS_ORG_TOKEN")

	if !found || discordBotsToken == "" {
		log.WithField("warn", "DISCORDBOTS_ORG_TOKEN not defined in environment variables").
			Warnf("Integration with discordbots.org integration disabled")
	} else {
		botID, found = os.LookupEnv("BOT_ID")

		if !found || botID == "" {
			log.WithField("warn", "BOT_ID not defined in environment variables").
				Warnf("Integration with discordbots.org disabled")
		}
	}

	// Create a new Discord session using the provided bot token
	dgBotSession, err := discordgo.New("Bot " + token)
	if err != nil {
		log.WithError(err).Fatalf("Error creating Discord session")
	}

	// Add event handlers
	dgBotSession.AddHandler(callbacks.Ready)            // Connection established with Discord
	dgBotSession.AddHandler(callbacks.MessageCreate)    // Chat messages with BOT_KEYWORD
	dgBotSession.AddHandler(callbacks.VoiceStateUpdate) // Updates to voice channel state

	// Open the websocket and begin listening
	err = dgBotSession.Open()
	if err != nil {
		log.WithError(err).Fatalf("Error opening Discord session")
	}

	// Cleanly close down the Discord session
	defer dgBotSession.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGHUP)
	signal.Notify(stop, os.Interrupt)

	go server.MonitorGuildsUpdate(dgBotSession, discordBotsToken, botID)

	httpServer := server.New(port)

	log.Debugf("Starting internal HTTP server instance")
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.WithError(err).Warnf("Internal HTTP server")
		}
	}()

	// Block until the OS signal
	<-stop

	log.Warnf("Caught graceful shutdown signal")

	// Cleanly shutdown the HTTP server
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	httpServer.Shutdown(ctx)

	return
}
