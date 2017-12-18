package main

import "github.com/ewohltman/discordEphemeralRolesProject/pkg/logging"

func main() {
	log := logging.Instance()

	log.Infof("Hello, World!")
}
