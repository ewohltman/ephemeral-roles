.PHONY: lint test build buildBeanstalk buildDocker

lint:
	golangci-lint run --enable-all --deadline=5m ./...

test:
	go test -v -race -coverprofile=coverage.out ./...

build:
	go build -o build/package/ephemeral-roles cmd/ephemeral-roles/ephemeral-roles.go

buildBeanstalk:
	go build -o bin/ephemeral-roles cmd/ephemeral-roles/ephemeral-roles.go

buildDocker: build
	docker image build -t ephemeral-roles build/package

push:
	docker push ewohltman/ephemeral-roles:latest
