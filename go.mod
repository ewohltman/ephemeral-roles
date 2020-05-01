module github.com/ewohltman/ephemeral-roles

go 1.14

replace github.com/bwmarrin/discordgo => github.com/ewohltman/discordgo v0.20.3-0.20200413094755-5a68d35e19f2

require (
	github.com/bwmarrin/discordgo v0.20.3
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kz/discordrus v1.2.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.6.0
	github.com/sirupsen/logrus v1.5.0
	github.com/uber/jaeger-client-go v2.23.1+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	go.uber.org/atomic v1.6.0 // indirect
)
