# discordrus | a [Discord](https://discordapp.com/) hook for [Logrus](https://github.com/Sirupsen/logrus) <img src="http://i.imgur.com/hTeVwmJ.png" width="40" height="40" alt=":walrus:" class="emoji" title=":walrus:"/> [![Travis CI](https://api.travis-ci.org/kz/discordrus.svg?branch=master)](https://travis-ci.org/kz/discordrus) [![GoDoc](https://godoc.org/github.com/puddingfactory/logentrus?status.svg)](https://godoc.org/github.com/kz/discordrus)

![Screenshot of discordrus in action](https://i.imgur.com/q8Tcmjn.png?1)

## Install

`go get -u github.com/kz/discordrus`

## Setup

In order to use this package, a Discord webhook URL is required. Find out how to obtain one [here](https://support.discordapp.com/hc/en-us/articles/228383668-Intro-to-Webhooks). You will need to be a server administrator to do this.

## Usage

Below is an example of how this package may be used. The options below are used only for the purpose of demonstration and chances are that you will not need to use any options at all (or if any, only the `Username` option).


```go
package main

import (
	"github.com/sirupsen/logrus"
	"os"
	"github.com/kz/discordrus"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetOutput(os.Stderr)
	logrus.SetLevel(logrus.TraceLevel)

	logrus.AddHook(discordrus.NewHook(
		// Use environment variable for security reasons
		os.Getenv("DISCORDRUS_WEBHOOK_URL"),
		// Set minimum level to DebugLevel to receive all log entries
		logrus.TraceLevel,
		&discordrus.Opts{
			Username:           "Test Username",
			Author:             "",                         // Setting this to a non-empty string adds the author text to the message header
			DisableTimestamp:   false,                      // Setting this to true will disable timestamps from appearing in the footer
			TimestampFormat:    "Jan 2 15:04:05.00000 MST", // The timestamp takes this format; if it is unset, it will take logrus' default format
			TimestampLocale:    nil,                        // The timestamp uses this locale; if it is unset, it will use time.Local
			EnableCustomColors: true,                       // If set to true, the below CustomLevelColors will apply
			CustomLevelColors: &discordrus.LevelColors{
				Trace: 3092790,
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

func main() {
	logrus.WithFields(logrus.Fields{"String": "hi", "Integer": 2, "Boolean": false}).Debug("Check this out! Awesome, right?")
}
```

All discordrus.Opts fields are optional.

Option | Description | Default | Valid options
--- | --- | --- | ---
Username | Replaces the default username of the webhook bot for the sent message only | Username unchanged | Any non-empty string (2-32 chars. inclusive)
Author | Adds an author field to the header if set | Author not set | Any non-empty string (1-256 chars inclusive)
DisableInlineFields | Inline means whether Discord will display the field in a column (with maximum three columns to a row). Setting this to `true` will cause Discord to display the field in its own row. | false | bool 
DisableTimestamp | Specifies whether the timestamp in the footer should be disabled | false | bool
TimestampFormat | Change the timestamp format | logrus's default time format | `"Jan 2 15:04:05.00000 MST"`, or any format accepted by Golang
TimestampLocale | Change the timestamp locale | `nil` | nil == time.Local, time.UTC, time.LoadLocation("America/New_York"), etc
EnableCustomColors | Specifies whether the `CustomLevelColors` opt value should be used instead of `discordrus.DefaultLevelColors`. If `true`, `CustomLevelColors` must be specified (or all colors will be set to the nil value of `0`, therefore displayed as white) | false | bool
CustomLevelColors | Replaces `discordrus.DefaultLevelColors`. All fields must be entered or they will default to the nil value of `0`. | Pointer to struct instance of `discordrus.LevelColors`
	
In addition to the above character count constraints, Discord has a maximum of 25 fields with their name and value limits being 256 and 1024 respectively. Furthermore, the description (i.e., logrus' error message) must be a maximum of 2048. All of these constraints, including the option constraints above, will automatically be truncated with no further action required.
 
## Acknowledgements
The following repositories have been helpful in creating this package: [puddingfactory/logentrus](https://github.com/puddingfactory/logentrus) for Logentries, [johntdyer/slackrus](https://github.com/johntdyer/slackrus) for Slack and [nubo/hiprus](https://github.com/nubo/hiprus) for Hipchat. Check them out!
