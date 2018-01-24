package server

import (
	"bytes"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/ephemeral-roles/pkg/discordBotsOrg"
	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/sirupsen/logrus"
)

// InternalStateCache is a mutex protected cache of values
type InternalStateCache struct {
	mu        sync.RWMutex
	guildList []*discordgo.Guild
	numGuilds int
}

// S is the struct for the internal HTTP server
type S struct {
	Logger *logrus.Logger
	mux    *http.ServeMux
}

// (s *S) ServeHTTP satisfies the http.Handler interface
func (s *S) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

var isc = &InternalStateCache{}
var log = logging.Instance()
var serverTest bool

// New returns a new pre-configured server instance
func New(port string) *http.Server {
	return &http.Server{
		Addr: ":" + port,
		Handler: server(
			func(s *S) {
				s.Logger = log
			},
		),
	}
}

func server(options ...func(*S)) *S {
	s := &S{
		Logger: log,
		mux:    http.NewServeMux(),
	}

	for _, f := range options {
		f(s)
	}

	// List the guilds our bot is a member of
	s.mux.HandleFunc(
		"/guilds",
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()

			isc.mu.RLock()
			defer isc.mu.RUnlock()

			buf := bytes.NewBuffer([]byte{})
			for _, guild := range isc.guildList {
				buf.Write([]byte(guild.Name + "\n"))
			}

			response := buf.Bytes()

			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Content-Length", strconv.Itoa(len(response)))

			_, err := w.Write(response)
			if err != nil {
				log.WithError(err).Errorf("Error writing /guilds HTTP response")
				return
			}
		},
	)

	// Default handler
	s.mux.HandleFunc(
		"/",
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
		},
	)

	return s
}

// MonitorGuildsUpdate will monitor for changes to isc.numGuilds and update
// discordbots.org appropriately
func MonitorGuildsUpdate(dgBotSession *discordgo.Session, token string, botID string) {
	// Initialize
	discordBotsOrgUpdate(dgBotSession, token, botID)

	if serverTest {
		isc.mu.Lock()
		isc.numGuilds = len(dgBotSession.State.Guilds)
		isc.mu.Unlock()

		monitorGuilds(dgBotSession, token, botID)

		isc.mu.Lock()
		isc.numGuilds = len(dgBotSession.State.Guilds) + 1
		isc.mu.Unlock()

		monitorGuilds(dgBotSession, token, botID)

		isc.mu.Lock()
		isc.numGuilds = len(dgBotSession.State.Guilds) - 1
		isc.mu.Unlock()

		monitorGuilds(dgBotSession, token, botID)

		return
	}

	for {
		monitorGuilds(dgBotSession, token, botID)

		time.Sleep(time.Second * 5)
	}
}

func monitorGuilds(dgBotSession *discordgo.Session, token string, botID string) {
	isc.mu.RLock()
	checkNum := isc.numGuilds
	isc.mu.RUnlock()

	guildsNum := len(dgBotSession.State.Guilds)

	switch {
	case guildsNum > checkNum:
		log.WithField(
			"guild",
			dgBotSession.State.Guilds[len(dgBotSession.State.Guilds)-1].Name,
		).Infof(dgBotSession.State.User.Username + " joined new guild")

		discordBotsOrgUpdate(dgBotSession, token, botID)
	case guildsNum < checkNum:
		log.Infof(dgBotSession.State.User.Username + " removed from guild")

		discordBotsOrgUpdate(dgBotSession, token, botID)
	}
}

func discordBotsOrgUpdate(dgBotSession *discordgo.Session, token string, botID string) {
	isc.mu.Lock()
	defer isc.mu.Unlock()

	isc.guildList = dgBotSession.State.Guilds
	isc.numGuilds = len(isc.guildList)

	if token != "" && botID != "" {
		response, err := discordBotsOrg.Update(token, botID, isc.numGuilds)
		if err != nil {
			log.WithError(err).Warnf("unable to update guild count")

			return
		}

		if response != "{}" {
			log.WithField("response", response).Warnf("discordbots.org integration: abnormal response")
		}
	}
}
