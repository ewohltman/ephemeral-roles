.PHONY: lint test build docker

lint:
	golangci-lint run

test:
	go test -v -race -coverprofile=coverage.out ./...

build:
	CGO_ENABLED=0 go build -o build/package/ephemeral-roles cmd/ephemeral-roles/ephemeral-roles.go

docker: build
	docker image build -t ewohltman/ephemeral-roles:latest .
