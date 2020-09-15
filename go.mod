module github.com/ewohltman/ephemeral-roles

go 1.15

replace github.com/bwmarrin/discordgo => github.com/ewohltman/discordgo v0.20.3-0.20200827215255-9e7ef1322884

require (
	github.com/bwmarrin/discordgo v0.22.0
	github.com/kz/discordrus v1.2.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/common v0.13.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a // indirect
	golang.org/x/sys v0.0.0-20200909081042-eff7692f9009 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
)
