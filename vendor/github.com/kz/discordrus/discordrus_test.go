package discordrus

import (
	"bytes"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetOutput(os.Stderr)
	logrus.SetLevel(logrus.DebugLevel)

	logrus.AddHook(NewHook(
		// Use environment variable for security reasons
		os.Getenv("DISCORDRUS_WEBHOOK_URL"),
		// Set minimum level to DebugLevel to receive all log entries
		logrus.DebugLevel,
		&Opts{
			Username:           "Test Username",
			Author:             "",                         // Setting this to a non-empty string adds the author text to the message header
			DisableTimestamp:   false,                      // Setting this to true will disable timestamps from appearing in the footer
			TimestampFormat:    "Jan 2 15:04:05.00000 MST", // The timestamp takes this format; if it is unset, it will take logrus' default format
			TimestampLocale:    nil,                        // The timestamp uses this locale; if it is unset, it will use time.Local
			EnableCustomColors: true,                       // If set to true, the below CustomLevelColors will apply
			CustomLevelColors: &LevelColors{
				Debug: 10170623,
				Info:  3581519,
				Warn:  14327864,
				Error: 13631488,
				Panic: 13631488,
				Fatal: 13631488,
			},
			DisableInlineFields: false, // If set to true, fields will not appear in columns ("inline")
		},
	))
}

// TestMaxLengths ensures that the specified maximum lengths match the maximum lengths of the Discord API
func TestMaxLengths(t *testing.T) {
	type testType struct {
		name               string
		json               string
		expectedStatusCode int
	}
	// Set up tests
	tests := []testType{
		{
			name:               "minimal",
			json:               "{\"embeds\":[{\"author\":{\"name\":\"A\"},\"description\":\"A\",\"fields\":[{\"name\":\"A\",\"value\":\"A\"}],\"title\":\"A\"}],\"username\":\"AA\"}",
			expectedStatusCode: 204,
		},
		{
			name:               "authorMax",
			json:               "{\"embeds\":[{\"author\":{\"name\":\"" + strings.Repeat("A", maxAuthorChars) + "\"},\"description\":\"A\",\"fields\":[{\"name\":\"A\",\"value\":\"A\"}],\"title\":\"A\"}],\"username\":\"AA\"}",
			expectedStatusCode: 204,
		},
		{
			name:               "authorMaxPlusOne",
			json:               "{\"embeds\":[{\"author\":{\"name\":\"" + strings.Repeat("A", maxAuthorChars+1) + "\"},\"description\":\"A\",\"fields\":[{\"name\":\"A\",\"value\":\"A\"}],\"title\":\"A\"}],\"username\":\"AA\"}",
			expectedStatusCode: 400,
		},
		{
			name:               "fieldNumMax",
			json:               "{\"embeds\":[{\"author\":{\"name\":\"A\"},\"description\":\"A\",\"fields\":[" + strings.Repeat("{\"name\":\"A\",\"value\":\"A\"},", maxFieldNum) + "{\"name\":\"A\",\"value\":\"A\"}],\"title\":\"A\"}],\"username\":\"AA\"}",
			expectedStatusCode: 204,
		},
		{
			name:               "fieldNumMaxPlusOne",
			json:               "{\"embeds\":[{\"author\":{\"name\":\"A\"},\"description\":\"A\",\"fields\":[" + strings.Repeat("{\"name\":\"b\",\"value\":\"A\"},", maxFieldNum+1) + "{\"name\":\"B\",\"value\":\"A\"}],\"title\":\"A\"}],\"username\":\"AA\"}",
			expectedStatusCode: 204, // For some reason, Discord does the truncation on their side to enforce the limit
		},
		{
			name:               "usernameEqualsOne",
			json:               "{\"embeds\":[{\"author\":{\"name\":\"A\"},\"description\":\"A\",\"fields\":[{\"name\":\"A\",\"value\":\"A\"}],\"title\":\"A\"}],\"username\":\"A\"}",
			expectedStatusCode: 400,
		},
		{
			name:               "usernameMin",
			json:               "{\"embeds\":[{\"author\":{\"name\":\"A\"},\"description\":\"A\",\"fields\":[{\"name\":\"A\",\"value\":\"A\"}],\"title\":\"A\"}],\"username\":\"" + strings.Repeat("A", minUsernameChars) + "\"}",
			expectedStatusCode: 204, // If this changes, then usernameTooShortMsg must be changed
		},
		{
			name:               "usernameMax",
			json:               "{\"embeds\":[{\"author\":{\"name\":\"A\"},\"description\":\"A\",\"fields\":[{\"name\":\"A\",\"value\":\"A\"}],\"title\":\"A\"}],\"username\":\"" + strings.Repeat("A", maxUsernameChars) + "\"}",
			expectedStatusCode: 204,
		},
		{
			name:               "usernameMaxPlusOne",
			json:               "{\"embeds\":[{\"author\":{\"name\":\"A\"},\"description\":\"A\",\"fields\":[{\"name\":\"A\",\"value\":\"A\"}],\"title\":\"A\"}],\"username\":\"" + strings.Repeat("A", maxUsernameChars+1) + "\"}",
			expectedStatusCode: 400,
		},
		{
			name:               "fieldNameMax",
			json:               "{\"embeds\":[{\"author\":{\"name\":\"A\"},\"description\":\"A\",\"fields\":[{\"name\":\"" + strings.Repeat("A", maxFieldNameChars) + "\",\"value\":\"A\"}],\"title\":\"A\"}],\"username\":\"AA\"}",
			expectedStatusCode: 204,
		},
		{
			name:               "fieldNameMaxPlusOne",
			json:               "{\"embeds\":[{\"author\":{\"name\":\"A\"},\"description\":\"A\",\"fields\":[{\"name\":\"" + strings.Repeat("A", maxFieldNameChars+1) + "\",\"value\":\"A\"}],\"title\":\"A\"}],\"username\":\"AA\"}",
			expectedStatusCode: 400,
		},
		{
			name:               "fieldValueMax",
			json:               "{\"embeds\":[{\"author\":{\"name\":\"A\"},\"description\":\"A\",\"fields\":[{\"name\":\"A\",\"value\":\"" + strings.Repeat("A", maxFieldValueChars) + "\"}],\"title\":\"A\"}],\"username\":\"AA\"}",
			expectedStatusCode: 204,
		},
		{
			name:               "fieldValueMaxPlusOne",
			json:               "{\"embeds\":[{\"author\":{\"name\":\"A\"},\"description\":\"A\",\"fields\":[{\"name\":\"A\",\"value\":\"" + strings.Repeat("A", maxFieldValueChars+1) + "\"}],\"title\":\"A\"}],\"username\":\"AA\"}",
			expectedStatusCode: 400,
		},
		{
			name:               "descriptionMax",
			json:               "{\"embeds\":[{\"author\":{\"name\":\"A\"},\"description\":\"" + strings.Repeat("A", maxDescriptionChars) + "\",\"fields\":[{\"name\":\"A\",\"value\":\"A\"}],\"title\":\"A\"}],\"username\":\"AA\"}",
			expectedStatusCode: 204,
		},
		{
			name:               "descriptionMaxPlusOne",
			json:               "{\"embeds\":[{\"author\":{\"name\":\"A\"},\"description\":\"" + strings.Repeat("A", maxDescriptionChars+1) + "\",\"fields\":[{\"name\":\"A\",\"value\":\"A\"}],\"title\":\"A\"}],\"username\":\"AA\"}",
			expectedStatusCode: 400,
		},
	}

	for _, test := range tests {
		res, err := http.Post(os.Getenv("DISCORDRUS_WEBHOOK_URL"), "application/json; charset=utf-8", bytes.NewBuffer([]byte(test.json)))
		if err != nil {
			t.Errorf("Error occured while sending HTTP request in %s test: %v\n", test.name, err.Error())
		} else if res.StatusCode != test.expectedStatusCode {
			t.Errorf("Unexpected status code when sending JSON in %s test: %v\n", test.name, res.StatusCode)
		}
		// Prevent Discord API rate limiting
		time.Sleep(500 * time.Millisecond)
	}
}

// TestHookIntegration is an integration test to ensure that log entries do send
func TestHookIntegration(t *testing.T) {
	logrus.WithFields(logrus.Fields{"String": "hi", "Integer": 2, "Boolean": false}).Debug("Check this out! Awesome, right?")
	logrus.WithFields(logrus.Fields{"String": "hi", "Integer": 2, "Boolean": false}).Info("Check this out! Awesome, right?")
	logrus.WithFields(logrus.Fields{"String": "hi", "Integer": 2, "Boolean": false}).Warn("Check this out! Awesome, right?")
	logrus.WithFields(logrus.Fields{"String": "hi", "Integer": 2, "Boolean": false}).Error("Check this out! Awesome, right?")
}
