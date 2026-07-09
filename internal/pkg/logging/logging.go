// Package logging provides a log/slog logging implementation. Configuration is
// determined via environment variables upon startup and the logging level may
// be changed at runtime. Logs are written to stdout and, when a Discord webhook
// is configured, to Discord as well.
package logging

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	slogdiscord "github.com/Bufferoverflovv/slog-discord"
)

// Logging level strings.
const (
	DebugLevel   = "debug"
	InfoLevel    = "info"
	WarningLevel = "warning"
	ErrorLevel   = "error"
	FatalLevel   = "fatal"
	PanicLevel   = "panic"
)

// Logging level color constants.
const (
	DebugColor   = 10170623
	InfoColor    = 3581519
	WarningColor = 14327864
	ErrorColor   = 13631488
	PanicColor   = 13631488
	FatalColor   = 13631488
)

// OptionFunc is used to configure options for a *Logger.
type OptionFunc func(*Logger)

// Logger owns the slog logging configuration and exposes an embedded
// *slog.Logger that fans out to stdout and, when configured, a Discord webhook.
type Logger struct {
	*slog.Logger

	level       *slog.LevelVar
	location    *time.Location
	webhookURL  string
	baseAttrs   []any
	badTimezone string
	output      io.Writer
}

// New returns a new *Logger instance configured with the OptionFunc arguments
// provided.
func New(options ...OptionFunc) *Logger {
	log := &Logger{
		level:    &slog.LevelVar{},
		location: time.UTC,
		output:   os.Stdout,
	}

	log.level.Set(slog.LevelInfo)

	for _, option := range options {
		option(log)
	}

	log.build()

	if log.badTimezone != "" {
		log.Warn("error parsing timezone location", "location", log.badTimezone)
	}

	return log
}

// OptionalOutput returns an OptionFunc to configure a *Logger to set where log
// messages should output to.
func OptionalOutput(output io.Writer) OptionFunc {
	return func(log *Logger) {
		log.output = output
	}
}

// OptionalShardID returns an OptionFunc to configure a *Logger to include a
// shardID field.
func OptionalShardID(shardID int) OptionFunc {
	return func(log *Logger) {
		log.baseAttrs = append(log.baseAttrs, slog.Int("shardID", shardID))
	}
}

// OptionalLogLevel returns an OptionFunc to configure a *Logger log level.
func OptionalLogLevel(logLevel string) OptionFunc {
	return func(log *Logger) {
		log.level.Set(parseLevel(logLevel))
	}
}

// OptionalTimezoneLocation returns an OptionFunc to configure a *Logger
// timezone location.
func OptionalTimezoneLocation(timezoneLocation string) OptionFunc {
	return func(log *Logger) {
		location, err := time.LoadLocation(timezoneLocation)
		if err != nil {
			log.badTimezone = timezoneLocation
			log.location = time.UTC
			return
		}

		log.location = location
	}
}

// OptionalDiscordWebhook returns an OptionFunc to configure a *Logger to also
// log to a Discord webhook URL.
func OptionalDiscordWebhook(webhookURL string) OptionFunc {
	return func(log *Logger) {
		log.webhookURL = webhookURL
	}
}

// UpdateLevel allows for runtime updates of the logging level. It adjusts the
// stdout handler live; the Discord handler keeps its startup level.
func (l *Logger) UpdateLevel(level string) {
	l.level.Set(parseLevel(level))
}

// build (re)constructs the *slog.Logger from the current configuration, fanning
// out to Discord when a webhook is configured.
func (l *Logger) build() {
	var handler slog.Handler = slog.NewTextHandler(l.output, &slog.HandlerOptions{
		Level:       l.level,
		ReplaceAttr: l.replaceAttr,
	})

	if l.webhookURL != "" {
		discordHandler := slogdiscord.NewDiscordHandler(slogdiscord.DiscordWebhookConfig{
			WebhookURL: l.webhookURL,
			MinLevel:   l.level.Level(),
			LevelColors: slogdiscord.LevelColors{
				slog.LevelDebug.String(): DebugColor,
				slog.LevelInfo.String():  InfoColor,
				slog.LevelWarn.String():  WarningColor,
				slog.LevelError.String(): ErrorColor,
			},
			CustomEmbed: discordEmbed,
		})

		handler = &fanoutHandler{handlers: []slog.Handler{handler, discordHandler}}
	}

	slogLogger := slog.New(handler)

	if len(l.baseAttrs) > 0 {
		slogLogger = slogLogger.With(l.baseAttrs...)
	}

	l.Logger = slogLogger
}

// replaceAttr rewrites the record timestamp into the configured location.
func (l *Logger) replaceAttr(_ []string, attr slog.Attr) slog.Attr {
	if l.location != nil && attr.Key == slog.TimeKey && attr.Value.Kind() == slog.KindTime {
		attr.Value = slog.TimeValue(attr.Value.Time().In(l.location))
	}

	return attr
}

// fanoutHandler is a slog.Handler that dispatches each record to every wrapped
// handler, so a single logger can write to both stdout and Discord.
type fanoutHandler struct {
	handlers []slog.Handler
}

// Enabled reports whether any wrapped handler is enabled for the level.
func (f *fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range f.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}

	return false
}

// Handle dispatches the record to every wrapped handler that is enabled for it.
//
//nolint:gocritic // slog.Handler requires slog.Record to be passed by value.
func (f *fanoutHandler) Handle(ctx context.Context, record slog.Record) error {
	var errs []error

	for _, handler := range f.handlers {
		if !handler.Enabled(ctx, record.Level) {
			continue
		}

		if err := handler.Handle(ctx, record.Clone()); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// WithAttrs returns a new fanoutHandler with the attributes applied to each
// wrapped handler.
func (f *fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(f.handlers))

	for i, handler := range f.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}

	return &fanoutHandler{handlers: handlers}
}

// WithGroup returns a new fanoutHandler with the group applied to each wrapped
// handler.
func (f *fanoutHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(f.handlers))

	for i, handler := range f.handlers {
		handlers[i] = handler.WithGroup(name)
	}

	return &fanoutHandler{handlers: handlers}
}

// discordEmbed builds the Discord embed for a record. The library's default
// embed drops the log message, so it is set as the embed description here.
//
//nolint:gocritic // slogdiscord.CustomEmbed requires slog.Record by value.
func discordEmbed(record slog.Record, levelColors slogdiscord.LevelColors) *slogdiscord.DiscordEmbed {
	embed := slogdiscord.DefaultEmbed(record, levelColors)
	embed.Description = record.Message

	return embed
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case DebugLevel:
		return slog.LevelDebug
	case InfoLevel:
		return slog.LevelInfo
	case WarningLevel:
		return slog.LevelWarn
	case ErrorLevel, FatalLevel, PanicLevel:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
