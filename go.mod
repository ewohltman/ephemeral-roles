module github.com/ewohltman/ephemeral-roles

go 1.14

replace github.com/bwmarrin/discordgo => github.com/ewohltman/discordgo v0.20.3-0.20200601122938-573e5411dae8

require (
	github.com/bwmarrin/discordgo v0.20.3
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/kz/discordrus v1.2.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.6.0
	github.com/prometheus/common v0.10.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/uber/jaeger-client-go v2.23.1+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	go.uber.org/atomic v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37 // indirect
	golang.org/x/sys v0.0.0-20200602100848-8d3cce7afc34 // indirect
	google.golang.org/protobuf v1.24.0 // indirect
)
