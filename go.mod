module github.com/ewohltman/ephemeral-roles

go 1.15

replace github.com/bwmarrin/discordgo => github.com/ewohltman/discordgo v0.20.3-0.20200827215255-9e7ef1322884

require (
	github.com/bwmarrin/discordgo v0.22.0
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/kz/discordrus v1.2.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/sirupsen/logrus v1.6.0
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	go.uber.org/atomic v1.6.0 // indirect
)
