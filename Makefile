MAKEFLAGS += --no-print-directory

SHELL := /bin/bash
.SHELLFLAGS := -o pipefail -c

MODULE := $(shell grep '^module' go.mod | awk '{print $$2}')
PROJECT_DIRS := $(shell go list -f "{{.Dir}}" ./...)

VERSION := $(shell grep 'image: ewohltman/ephemeral-roles:' deployments/kubernetes/statefulset.yml | awk -F: '{print $$3}')

ifneq ($(OS),Windows_NT)
  GO_TEST_RACE_FLAG := -race
endif

GOIMPORTS := go tool -modfile=tools/go.mod goimports -local $(MODULE)/ -w
GOLANGCI_LINT := go tool -modfile=tools/go.mod golangci-lint run ./...
GOTESTSUM := go tool -modfile=tools/go.mod gotestsum --jsonfile test-report.json -- $(GO_TEST_RACE_FLAG) -coverprofile=coverage.out

.PHONY: tidy
tidy:
	go mod tidy
	cd tools && go mod tidy

.PHONY: fmt
fmt: tidy
	@echo 'gofmt -s -w'
	@gofmt -s -w $(PROJECT_DIRS)
	@echo "$(GOIMPORTS)"
	@$(GOIMPORTS) $(PROJECT_DIRS)

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint: vet
	$(GOLANGCI_LINT)

.PHONY: nochanges
nochanges:
	git status --porcelain
	test -z "$$(git status --porcelain)"

.PHONY: test
test:
	@echo "$(GOTESTSUM)"
	@$(GOTESTSUM) $(shell go list ./... | grep -v '/cmd/')
	@echo -n "Coverage: "
	@go tool cover -func=coverage.out | grep 'total:' | column -t

.PHONY: build
build:
	CGO_ENABLED=0 go build -trimpath -gcflags=all=-trimpath=$(PWD) -asmflags=all=-trimpath=$(PWD) -o build/package/ephemeral-roles/ephemeral-roles cmd/ephemeral-roles/ephemeral-roles.go

.PHONY: image
image:
	podman image build -t ewohltman/ephemeral-roles:$(VERSION) build/package/ephemeral-roles

.PHONY: push
push:
	podman login -u $(DOCKER_USER) -p $(DOCKER_PASS)
	podman push ewohltman/ephemeral-roles:$(VERSION)
	podman tag ewohltman/ephemeral-roles:$(VERSION) ewohltman/ephemeral-roles:latest
	podman logout
