module github.com/ewohltman/ephemeral-roles

go 1.14

replace github.com/bwmarrin/discordgo => github.com/ewohltman/discordgo v0.20.3-0.20200407200000-0cc8c7081932

require (
	github.com/bwmarrin/discordgo v0.20.2
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/golang/protobuf v1.3.5 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kz/discordrus v1.2.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.5.1
	github.com/prometheus/procfs v0.0.11 // indirect
	github.com/sirupsen/logrus v1.5.0
	github.com/uber/jaeger-client-go v2.22.1+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	go.uber.org/atomic v1.6.0 // indirect
	golang.org/x/sys v0.0.0-20200408040146-ea54a3c99b9b // indirect
)
