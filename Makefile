# METADATA
VERSION=v0.1.0

# CMD
BIN_DIR=bin
BIN_NAME=govcourier
MAIN_PATH=cmd/main.go
BIN_PATH=$(BIN_DIR)/$(BIN_NAME)

# DOCKER
IMAGE_NAME=govcourier

GO=go

.PHONY: all version test fmt vet generate prepare

all: build

version:
	@$(GO) version
	@$(GO) env

COVERAGE=cover.out
COVERAGE_ARGS=-covermode count -coverprofile $(COVERAGE)

test:
	$(GO) test -v -cover $(COVERAGE_ARGS) ./...

coverage:
	$(GO) tool cover -html $(COVERAGE)

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

generate:
	$(GO) generate ./...

prepare: fmt vet

.PHONY: gen cleangen dev clean build-bin build

GENSRC=$(shell find . -name '*_gen.go')

gen: generate fmt

cleangen:
	rm $(GENSRC)

dev:
	$(GO) run -ldflags "-X main.GitHash=$$(git rev-parse --verify HEAD)" $(MAIN_PATH) --config configdev

clean:
	if [ -d $(BIN_DIR) ]; then rm -r $(BIN_DIR); fi

build-bin:
	mkdir -p $(BIN_DIR)
	if [ -f $(BIN_PATH) ]; then rm $(BIN_PATH); fi
	CGO_ENABLED=0 $(GO) build -a -tags netgo -ldflags "-w -s -X main.GitHash=$$(git rev-parse --verify HEAD)" -o $(BIN_PATH) $(MAIN_PATH)

build: clean build-bin

## docker
.PHONY: build-docker produp proddown devup devdown docker-clean

build-docker:
	docker build -f ./cmd/Dockerfile -t $(IMAGE_NAME):$(VERSION) -t $(IMAGE_NAME):latest .

produp:
	docker-compose -f dc.main.yaml -f dc.prod.yaml -f dc.compose.yaml up -d

proddown:
	docker-compose -f dc.main.yaml -f dc.prod.yaml -f dc.compose.yaml down

devup:
	docker-compose -f dc.main.yaml -f dc.dev.yaml up -d

devdown:
	docker-compose -f dc.main.yaml -f dc.dev.yaml down

docker-clean:
	if [ "$$(docker ps -q -f status=running)" ]; \
		then docker stop $$(docker ps -q -f status=running); fi
	if [ "$$(docker ps -q -f status=restarting)" ]; \
		then docker stop $$(docker ps -q -f status=restarting); fi
	if [ "$$(docker ps -q -f status=exited)" ]; \
		then docker rm $$(docker ps -q -f status=exited); fi
	if [ "$$(docker ps -q -f status=created)" ]; \
		then docker rm $$(docker ps -q -f status=created); fi
