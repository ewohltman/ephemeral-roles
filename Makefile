.PHONY: lint test build buildBeanstalk

lint:
	golangci-lint run --enable-all --deadline=5m ./...

test:
	go test -v -race -coverprofile=coverage.out ./...

build:
	go build -o build/package/ephemeral-roles cmd/ephemeral-roles/ephemeral-roles.go

buildBeanstalk:
	go build -o bin/ephemeral-roles cmd/ephemeral-roles/ephemeral-roles.go
	chmod +x bin/ephemeral-roles

push:
	docker push ewohltman/ephemeral-roles:latest
