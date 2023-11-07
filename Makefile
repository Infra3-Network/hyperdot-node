VERSION ?= $(shell git rev-parse --short HEAD)
GO?=$(shell which go)
export GOBIN := $(BIN)
export PATH := $(BIN):$(PATH)

.PHONY: build/docker
build/docker: ## Build the docker image.
	DOCKER_BUILDKIT=1 \
	docker build \
		-f ./Dockerfile \
		-t hyperdot/node:$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

.PHONY: build/docker-arm
build-arm:
	DCOMPOSE_DOCKER_CLI_BUILD=1 DOCKER_FILE="Dockerfile-arm" docker build \
		-t hyperdot/node:$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

.PHONY: up-docker
up-docker: 
	sudo docker-compose -f orchestration/docker-compose/docker-compose.yml up -d

.PHONY: stop-docker
stop-docker: 
	sudo docker-compose -f orchestration/docker-compose/docker-compose.yml stop

.PHONY: up-test
up-test: 
	docker-compose -f tests/docker-compose.yaml up -d


.PHONY: tests
tests:
	go test -v ./tests/ --count=1

.PHONY: lint
lint: 
	  golangci-lint run
	

.PHONY: clean
clean:
	rm -rf tests/hyperdot.db
	docker-compose -f tests/docker-compose.yaml down 

