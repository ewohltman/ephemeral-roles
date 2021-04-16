module github.com/ewohltman/ephemeral-roles

go 1.16

// replace github.com/bwmarrin/discordgo => ../discordgo

replace github.com/bwmarrin/discordgo => github.com/ewohltman/discordgo v0.20.3-0.20210128144321-553847c12fc0

require (
	github.com/HdrHistogram/hdrhistogram-go v0.9.0 // indirect
	github.com/bwmarrin/discordgo v0.23.2
	github.com/caarlos0/env v3.5.0+incompatible
	github.com/kz/discordrus v1.3.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/prometheus/client_golang v1.10.0
	github.com/sirupsen/logrus v1.8.1
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/automaxprocs v1.4.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad // indirect
	google.golang.org/protobuf v1.25.0 // indirect
)
