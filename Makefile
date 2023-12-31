VERSION ?= $(shell git rev-parse --short HEAD)
GO?=$(shell which go)
export GOBIN := $(BIN)
export PATH := $(BIN):$(PATH)

# proxy
HTTP_PROXY ?= "http://192.168.1.5:17890"
HTTPS_PROXY ?= "http://192.168.1.5:17890"

.PHONY: build/docker
build/docker: ## Build the docker image.
	DOCKER_BUILDKIT=1 \
	docker build \
		-f ./Dockerfile \
		-t hyperdot/fronted:$(VERSION) \
		--build-arg "HTTP_PROXY=$(HTTP_PROXY)" \
		--build-arg "HTTPS_PROXY=$(HTTPS_PROXY)" \

.PHONY: build/docker
build/docker: ## Build the docker image.
	DOCKER_BUILDKIT=1 \
	docker build \
		-f ./Dockerfile \
		-t hyperdot/node:latest \
		--build-arg "HTTP_PROXY=$(HTTP_PROXY)" \
		--build-arg "HTTPS_PROXY=$(HTTPS_PROXY)" \
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

.PHONY: up-infra
up-infra: 
	sudo docker-compose -f orchestration/docker-compose/infra-docker-compose.yml up -d

.PHONY: stop-infra
stop-infra: 
	sudo docker-compose -f orchestration/docker-compose/infra-docker-compose.yml stop

.PHONY: rm-infra
rm-infra: 
	sudo docker-compose -f orchestration/docker-compose/infra-docker-compose.yml down

.PHONY: up
up: 
	sudo docker-compose -f orchestration/docker-compose/docker-compose.yml up -d

.PHONY: stop
stop: 
	sudo docker-compose -f orchestration/docker-compose/docker-compose.yml stop

.PHONY: rm
rm: 
	sudo docker-compose -f orchestration/docker-compose/docker-compose.yml down


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

