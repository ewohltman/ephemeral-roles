package main

import (
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/ephemeral-roles/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/pkg/logging"
)

var log = logging.Instance()

func main() {
	// Check for PORT, we need this to connect to Discord
	port, found := os.LookupEnv("PORT")
	if !found || port == "" {
		port = "8080"
	}

	// Check for EPH_BOT_TOKEN, we need this to connect to Discord
	token, found := os.LookupEnv("EPH_BOT_TOKEN")
	if !found || token == "" {
		log.Fatalf("EPH_BOT_TOKEN not defined in environment variables")
	}

	// Check for EPH_BOT_NAME, we don't need it now but it's required in the callbacks
	_, found = os.LookupEnv("EPH_BOT_NAME")
	if !found {
		log.Fatalf("EPH_BOT_NAME not defined in environment variables")
	}

	// Check for EPH_BOT_KEYWORD, we don't need it now but it's required in the callbacks
	_, found = os.LookupEnv("EPH_BOT_KEYWORD")
	if !found {
		log.Fatalf("EPH_BOT_KEYWORD not defined in environment variables")
	}

	// Check for EPH_CHANNEL_PREFIX, we don't need it now but it's required in the callbacks
	_, found = os.LookupEnv("EPH_ROLE_PREFIX")
	if !found {
		log.Fatalf("EPH_CHANNEL_PREFIX not defined in environment variables")
	}

	// Create a new Discord session using the provided bot token
	dgBot, err := discordgo.New("Bot " + token)
	if err != nil {
		log.WithError(err).Fatalf("Error creating Discord session")
	}

	// Add event handlers
	dgBot.AddHandler(callbacks.Ready)
	dgBot.AddHandler(callbacks.MessageCreate)
	dgBot.AddHandler(callbacks.VoiceStateUpdate)

	// Open the websocket and begin listening
	err = dgBot.Open()
	if err != nil {
		log.WithError(err).Fatalf("Error opening Discord session")
	}
	defer dgBot.Close() // Cleanly close down the Discord session

	// Set up handler funcs and an HTTP server to live in a container
	http.HandleFunc(
		"/",
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
		},
	)

	// Do something special with /status later?
	http.HandleFunc(
		"/status",
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
		},
	)

	log.WithError(http.ListenAndServe(":"+port, nil)).Fatalf("HTTP server error")

	return
}
