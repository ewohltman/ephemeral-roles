.PHONY: lint test build docker push deploy

COMMIT=$(shell git rev-parse --short HEAD)

lint:
	golangci-lint run ./...

test:
	go test -v -race -coverprofile=coverage.out ./...

build:
	CGO_ENABLED=0 go build -o build/package/ephemeral-roles cmd/ephemeral-roles/ephemeral-roles.go

image:
	docker image build -t ewohltman/ephemeral-roles:latest .

push:
	docker login -u "${DOCKER_USER}" -p "${DOCKER_PASS}"
	docker push ewohltman/ephemeral-roles:latest
	docker logout

deploy:
	kubectl apply -f deployments/kubernetes/service.yml
	kubectl apply -f deployments/kubernetes/ingress.yml
	sed "s/{COMMIT}/${COMMIT}/g" deployments/kubernetes/deployment.yml
