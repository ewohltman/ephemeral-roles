.PHONY: generate fmt lint test build build-debug pull-parent-image image push deploy

MAKEFILE_PATH=$(shell readlink -f "${0}")
MAKEFILE_DIR=$(shell dirname "${MAKEFILE_PATH}")

version=$(shell grep ' .*version: .*' deployments/kubernetes/statefulset.yml | awk '{print $$3}')

parentImage=alpine:latest

protocDirectory=internal/pkg/distributed/api
protocArgumentsIncludes=-I ${protocDirectory}
protocArgumentsGo=--go_out=${protocDirectory} --go_opt=paths=source_relative
protocArgumentsGoGRPC=--go-grpc_out=${protocDirectory} --go-grpc_opt=paths=source_relative
protocArguments=${protocArgumentsIncludes} ${protocArgumentsGo} ${protocArgumentsGoGRPC}

generate:
	protoc ${protocArguments} ${protocDirectory}/api.proto

fmt:
	gofmt -s -w . && goimports -w .

lint:
	golangci-lint run ./...

test:
	go test -v -race -coverprofile=coverage.out ./...

build:
	CGO_ENABLED=0 go build -o build/package/ephemeral-roles/ephemeral-roles cmd/ephemeral-roles/ephemeral-roles.go

build-debug:
	CGO_ENABLED=0 go build -gcflags "all=-N -l" -o build/package/ephemeral-roles-debug/ephemeral-roles-debug cmd/ephemeral-roles/ephemeral-roles.go

pull-parent-image:
	docker pull ${parentImage}

image: pull-parent-image
	@echo "Build image: ${version}"
	docker image build -t ewohltman/ephemeral-roles:${version} build/package/ephemeral-roles

push:
	docker login -u ${DOCKER_USER} -p ${DOCKER_PASS}
	docker push ewohltman/ephemeral-roles:${version}
	docker tag ewohltman/ephemeral-roles:${version} ewohltman/ephemeral-roles:latest
	docker logout

deploy:
	${MAKEFILE_DIR}/scripts/deploy.sh
