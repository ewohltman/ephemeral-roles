.PHONY: lint test build docker push deploy

MAKEFILE_PATH=$(shell readlink -f "${0}")
MAKEFILE_DIR=$(shell dirname "${MAKEFILE_PATH}")

parentImage=alpine:latest

lint:
	golangci-lint run ./...

test:
	go test -v -race -coverprofile=coverage.out ./...

build:
	CGO_ENABLED=0 go build -o build/package/ephemeral-roles cmd/ephemeral-roles/ephemeral-roles.go

image:
	docker pull "${parentImage}"
	docker image build -t ewohltman/ephemeral-roles:latest .

push:
	docker login -u "${DOCKER_USER}" -p "${DOCKER_PASS}"
	docker push ewohltman/ephemeral-roles:latest
	docker logout

deploy:
	${MAKEFILE_DIR}/scripts/deploy.sh
