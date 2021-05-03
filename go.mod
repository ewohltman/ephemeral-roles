module github.com/ewohltman/ephemeral-roles

go 1.16

replace github.com/bwmarrin/discordgo => github.com/ewohltman/discordgo v0.23.3-0.20210503155028-0043af39f282

require (
	github.com/HdrHistogram/hdrhistogram-go v0.9.0 // indirect
	github.com/bwmarrin/discordgo v0.23.3-0.20210410202908-577e7dd4f6cc
	github.com/caarlos0/env v3.5.0+incompatible
	github.com/goccy/go-json v0.4.13
	github.com/kz/discordrus v1.3.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/prometheus/client_golang v1.10.0
	github.com/sirupsen/logrus v1.8.1
	github.com/uber/jaeger-client-go v2.27.0+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/automaxprocs v1.4.0
	google.golang.org/protobuf v1.25.0 // indirect
)
