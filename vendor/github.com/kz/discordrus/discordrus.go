package discordrus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	maxFieldNum         = 25
	minUsernameChars    = 2
	maxUsernameChars    = 32
	maxAuthorChars      = 256
	maxFieldNameChars   = 256
	maxFieldValueChars  = 1024
	maxDescriptionChars = 2048
	usernameTooShortMsg = " (USERNAME TOO SHORT)"
)

// Opts contains the options available for the hook
type Opts struct {
	// Username replaces the default username of the webhook bot for the sent message only if set (default: none)
	Username string
	// Author adds an author field if set (default: none)
	Author string
	// DisableInlineFields causes fields to be displayed one per line as opposed to being inline (i.e., in columns) (default: false)
	DisableInlineFields bool
	// EnableCustomColors specifies whether CustomLevelColors should be used instead of DefaultLevelColors (default: true)
	EnableCustomColors bool
	// CustomLevelColors is a LevelColors struct which replaces DefaultLevelColors if EnableCustomColors is set to true (default: none)
	CustomLevelColors *LevelColors
	// DisableTimestamp specifies whether the timestamp in the footer should be disabled (default: false)
	DisableTimestamp bool
	// TimestampFormat specifies a custom format for the footer
	TimestampFormat string
}

// Hook is a hook to send logs to Discord
type Hook struct {
	// WebhookURL is the full Discord webhook URL
	WebhookURL string
	// MinLevel is the minimum priority level to enable logging for
	MinLevel logrus.Level
	// Opts contains the options available for the hook
	Opts *Opts
}

// NewHook creates a new instance of a hook, ensures correct string lengths and returns its pointer
func NewHook(webhookURL string, minLevel logrus.Level, opts *Opts) *Hook {
	hook := Hook{
		WebhookURL: webhookURL,
		MinLevel:   minLevel,
		Opts:       opts,
	}

	// Ensure correct username length
	if hook.Opts.Username != "" && len(hook.Opts.Username) < minUsernameChars {
		// Append "(USERNAME TOO SHORT)" in order not to disrupt logging operations
		hook.Opts.Username = hook.Opts.Username + usernameTooShortMsg
	} else if len(hook.Opts.Username) > maxUsernameChars {
		hook.Opts.Username = hook.Opts.Username[:maxUsernameChars]
	}

	// Truncate author
	if len(hook.Opts.Author) > maxAuthorChars {
		hook.Opts.Author = hook.Opts.Author[:maxAuthorChars]
	}

	return &hook
}

func (hook *Hook) Fire(entry *logrus.Entry) error {
	// Parse the entry to a Discord webhook object in JSON form
	webhookObject, err := hook.parseToJson(entry)
	if err != nil {
		return err
	}

	err = hook.send(webhookObject)
	if err != nil {
		return err
	}

	return nil
}

func (hook *Hook) Levels() []logrus.Level {
	return LevelThreshold(hook.MinLevel)
}

func (hook *Hook) parseToJson(entry *logrus.Entry) (*[]byte, error) {
	// Create struct mapping to Discord webhook object
	var data = map[string]interface{}{
		"embeds": []map[string]interface{}{},
	}
	var embed = map[string]interface{}{
		"title": strings.ToUpper(entry.Level.String()),
	}
	var fields = []map[string]interface{}{}

	// Add username to data
	if hook.Opts.Username != "" {
		data["username"] = hook.Opts.Username
	}

	// Add description to embed
	if len(entry.Message) > maxDescriptionChars {
		entry.Message = entry.Message[:maxDescriptionChars]
	}
	embed["description"] = entry.Message

	// Add color to embed
	if hook.Opts.EnableCustomColors {
		embed["color"] = hook.Opts.CustomLevelColors.LevelColor(entry.Level)
	} else {
		embed["color"] = DefaultLevelColors.LevelColor(entry.Level)
	}

	// Add author to embed
	if hook.Opts.Author != "" {
		embed["author"] = map[string]interface{}{"name": hook.Opts.Author}
	}

	// Add footer to embed
	if !hook.Opts.DisableTimestamp {
		timestamp := ""
		if hook.Opts.TimestampFormat != "" {
			timestamp = entry.Time.Format(hook.Opts.TimestampFormat)
		} else {
			timestamp = entry.Time.String()
		}
		embed["footer"] = map[string]interface{}{
			"text": timestamp,
		}
	}

	// Add fields to embed
	counter := 0
	for name, value := range entry.Data {
		// Ensure that the maximum field number is not exceeded
		if counter > maxFieldNum {
			break
		}

		// Make value a string
		valueStr := fmt.Sprintf("%v", value)

		// Truncate names and values which are too long
		if len(name) > maxFieldNameChars {
			name = name[:maxFieldNameChars]
		}
		if len(valueStr) > maxFieldValueChars {
			valueStr = valueStr[:maxFieldValueChars]
		}

		var embedField = map[string]interface{}{
			"name":   name,
			"value":  valueStr,
			"inline": !hook.Opts.DisableInlineFields,
		}
		fields = append(fields, embedField)
		counter++
	}

	// Merge fields and embed into data
	embed["fields"] = fields
	data["embeds"] = []map[string]interface{}{embed}

	marshaled, err := json.Marshal(data)
	return &marshaled, err
}

func (hook *Hook) send(webhookObject *[]byte) error {
	_, err := http.Post(hook.WebhookURL, "application/json; charset=utf-8", bytes.NewBuffer(*webhookObject))
	if err != nil {
		return err
	}
	return nil
}
