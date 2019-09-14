.PHONY: lint test build buildBeanstalk docker

lint:
	golangci-lint run --enable-all --deadline=5m ./...

test:
	go test -v -race -coverprofile=coverage.out ./...

build:
	go build -tags netgo -a -v -o build/package/ephemeral-roles cmd/ephemeral-roles/ephemeral-roles.go

docker: build
	docker image build -t ewohltman/ephemeral-roles:latest .
