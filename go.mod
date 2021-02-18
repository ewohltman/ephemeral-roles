module github.com/ewohltman/ephemeral-roles

go 1.15

// replace github.com/bwmarrin/discordgo => ../discordgo

replace github.com/bwmarrin/discordgo => github.com/ewohltman/discordgo v0.20.3-0.20210128144321-553847c12fc0

require (
	github.com/HdrHistogram/hdrhistogram-go v0.9.0 // indirect
	github.com/bwmarrin/discordgo v0.23.1
	github.com/caarlos0/env v3.5.0+incompatible
	github.com/kz/discordrus v1.2.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/prometheus/client_golang v1.9.0
	github.com/sirupsen/logrus v1.8.0
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.4.0+incompatible
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad // indirect
	golang.org/x/sys v0.0.0-20201231184435-2d18734c6014 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
)
