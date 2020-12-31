APPNAME:=ard2rss
IMAGE_NAME:=ard2rss
RELEASE?=0

DOCKER_CMD := $(shell command -v podman 2> /dev/null || echo docker)

ifeq ($(RELEASE), 1)
	# Strip debug information from the binary
	GO_LDFLAGS+=-s -w
endif
GO_LDFLAGS:=-ldflags="$(GO_LDFLAGS)"


.PHONY: default
default: test

.PHONY: build
build:
	go build $(GO_LDFLAGS) -o ./build/$(APPNAME) -v cmd/main.go

.PHONY: docker
docker:
	$(DOCKER_CMD) build -t $(IMAGE_NAME) .

.PHONY: test
test: build
	go test -v -race ./...

.PHONY: clean
clean:
	rm -rf ./build