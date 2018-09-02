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

var (
	token            string
	port             string
	discordBotsToken string
	botID            string
	log              = logging.Instance()
)

func checkRequired() {
	// Check for string from slice, these are not needed now, but are needed in the callbacks
	for _, envVar := range []string{"BOT_NAME", "BOT_KEYWORD", "ROLE_PREFIX"} {
		v, found := os.LookupEnv(envVar)
		if !found || v == "" {
			log.Fatalf("%s not defined in environment variables", envVar)
		}
	}

	var found bool

	// Check for BOT_TOKEN, we need this to connect to Discord
	token, found = os.LookupEnv("BOT_TOKEN")
	if !found || token == "" {
		log.Fatalf("BOT_TOKEN not defined in environment variables")
	}

	// Check for PORT, we need this to for our HTTP server in our container
	port, found = os.LookupEnv("PORT")
	if !found || port == "" {
		port = "8080"
	}
}

func checkOptional() {
	var found bool

	// Check for DISCORDBOTS_ORG_TOKEN and BOT_ID, we need these for optional discordbots.org integration
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
}

func checkEnvironment() {
	checkRequired()
	checkOptional()
}

func runHTTPServer(dgBotSession *discordgo.Session) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGHUP)
	signal.Notify(stop, os.Interrupt)

	httpServer := server.New(port, dgBotSession, discordBotsToken, botID)

	log.Debugf("Starting internal HTTP server instance")
	go func() {
		if serverError := httpServer.ListenAndServe(); serverError != nil {
			log.WithError(serverError).Warnf("Internal HTTP server")
		}
	}()

	// Block until the OS signal
	<-stop

	log.Warnf("Caught graceful shutdown signal")

	// Cleanly shutdown the HTTP server
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	err := httpServer.Shutdown(ctx)
	if err != nil {
		log.Error(err)
	}
}

func main() {
	log.Debugf("Bot starting up")

	checkEnvironment()

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
	defer dgBotSession.Close()

	runHTTPServer(dgBotSession)
}
